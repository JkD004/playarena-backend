// booking/booking_service.go
package booking

import (
	"errors"
	"log"
	"time"
	"github.com/JkD004/playarena-backend/notification"
	"github.com/JkD004/playarena-backend/venue"
	"github.com/JkD004/playarena-backend/gateway"
)

// CreateNewBooking handles the business logic
func CreateNewBooking(req *CreateBookingRequest, userID int64) (*Booking, error) {
	// 1. Get Venue details for pricing
	venueToBook, err := venue.GetVenueByID(req.VenueID)
	if err != nil {
		return nil, errors.New("venue not found or not available for booking")
	}

	// 2. Calculate Duration
	duration := req.EndTime.Sub(req.StartTime)
	if duration <= 0 {
		return nil, errors.New("end time must be after start time")
	}
	
	// 3. Calculate Price
	totalPrice := duration.Hours() * venueToBook.PricePerHour

	// 4. Check Availability
	available, err := IsSlotAvailable(req.VenueID, req.StartTime, req.EndTime)
	if err != nil {
		return nil, errors.New("error checking slot availability")
	}
	if !available {
		return nil, errors.New("this time slot is no longer available")
	}

	if req.StartTime.Before(time.Now().Add(-2 * time.Minute)) {
		return nil, errors.New("cannot book a time slot in the past")
	}

	// 5. Create Booking Object
	newBooking := &Booking{
		UserID:     userID,
		VenueID:    req.VenueID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		TotalPrice: totalPrice,
		Status:     "pending", // Default to pending until payment
	}

	// 6. Save to DB
	err = CreateBooking(newBooking)
	if err != nil {
		log.Println("Service error creating booking:", err)
		return nil, errors.New("failed to create booking")
	}

	return newBooking, nil
}

// BlockVenueSlot creates a "blocked" booking (Owner/Admin only)
func BlockVenueSlot(req *CreateBookingRequest, userID int64) error {
	// 1. Check availability
	available, err := IsSlotAvailable(req.VenueID, req.StartTime, req.EndTime)
	if err != nil {
		return errors.New("error checking slot availability")
	}
	if !available {
		return errors.New("this slot is already booked or blocked")
	}

	// 2. Create a "Blocked" booking
	// We use 'confirmed' status so it takes up the slot.
	// We set TotalPrice to 0 because it's an internal block.
	newBooking := &Booking{
		UserID:     userID,
		VenueID:    req.VenueID,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		TotalPrice: 0,           // <--- FIX: Set to 0 for blocks
		Status:     "confirmed", // <--- FIX: Confirmed immediately
	}

	err = CreateBooking(newBooking)
	if err != nil {
		log.Println("Service error blocking slot:", err)
		return errors.New("failed to block slot")
	}

	return nil
}

// ProcessPayment (Legacy/Internal helper)
func ProcessPayment(bookingID int64, paymentID string) error { // <--- Added paymentID arg to match
	// 1. Fetch booking
	booking, err := FindBookingByID(bookingID)
	if err != nil {
		return errors.New("booking not found")
	}

	// 2. Update DB
	err = ConfirmBookingPayment(bookingID, paymentID) // <--- Now matches repo signature
	if err != nil {
		return err
	}

	// 3. Notification
	message := "Payment successful! Your booking has been confirmed."
	_ = notification.CreateNotification(booking.UserID, message, "success")

	return nil
}
// --- Getters & Helpers ---

// GetBookingsForUser is the service-layer function
func GetBookingsForUser(userID int64) ([]Booking, error) {
	return FindBookingsByUserID(userID)
}

// CancelBooking handles the logic for canceling a booking
// CancelBooking handles the logic for canceling a booking
func CancelBooking(bookingID int64, userID int64) error {
	booking, err := FindBookingByID(bookingID)
	if err != nil {
		return errors.New("booking not found")
	}

	if booking.UserID != userID {
		return errors.New("unauthorized")
	}

	if time.Now().After(booking.StartTime.Add(-2 * time.Hour)) {
		return errors.New("cannot cancel less than 2 hours before start")
	}

	var newStatus string
	var notifMsg string

	if booking.Status == "pending" {
		newStatus = "canceled"
		notifMsg = "Booking canceled."
	} else if booking.Status == "confirmed" {
		// --- REFUND LOGIC ---
		if booking.PaymentID == "" {
			return errors.New("cannot refund: payment ID missing")
		}

		// Call the new Gateway package (No Import Cycle!)
		err := gateway.InitiateRefund(booking.PaymentID, booking.TotalPrice)
		if err != nil {
			log.Println("Refund Failed:", err)
			return errors.New("unable to process refund")
		}

		newStatus = "refunded"
		notifMsg = "Booking canceled. Refund initiated."
	} else {
		return errors.New("cannot cancel this booking")
	}

	err = UpdateBookingStatus(bookingID, userID, newStatus)
	if err != nil {
		return err
	}

	_ = notification.CreateNotification(userID, notifMsg, "warning")
	return nil
}

// GetBookingsForVenue is the service-layer function
func GetBookingsForVenue(venueID int64) ([]AdminBookingView, error) {
	return FindBookingsByVenueID(venueID)
}

// GetAllBookings is the service-layer function for the admin
func GetAllBookings() ([]AdminBookingView, error) {
	return FindAllBookings()
}


// booking/booking_service.go

// GetStatisticsForOwner handles logic for both roles
func GetStatisticsForOwner(userID int64, venueID int64, userRole string) (*OwnerStats, error) {
	var stats *OwnerStats
	var rawTimes []time.Time
	var err error

	// 1. Fetch Stats (Counts & Revenue)
	if userRole == "admin" {
		stats, err = GetVenueStatsSimple(venueID)
		if err != nil { return nil, err }
		
		// Get times for Admin (Confirmed + Present)
		rawTimes, err = FetchRawStartTimes(venueID)
		if err != nil { return nil, err }

	} else {
		stats, err = GetOwnerBookingStats(userID, venueID)
		if err != nil { return nil, err }
		
		// Get times for Owner (Confirmed + Present)
		// We use the same fetcher since ownership is checked in the handler
		rawTimes, err = FetchRawStartTimes(venueID)
		if err != nil { return nil, err }
	}

	// 2. Calculate Popular Time (Using the Go Helper we added)
	// This uses the list of times to find the busiest hour
	stats.PopularTime = CalculatePopularTime(rawTimes)

	return stats, nil
}

// GetStatisticsForAdmin is the service-layer function
func GetStatisticsForAdmin() (*OwnerStats, error) {
	bookings, revenue, err := GetPlatformBookingStats()
	if err != nil {
		return nil, err
	}

	popTime, err := GetPlatformPopularTime()
	if err != nil {
		return nil, err
	}

	stats := &OwnerStats{
		TotalBookings: bookings,
		TotalRevenue:  revenue,
		PopularTime:   popTime,
	}
	
	return stats, nil
}

// GetOwnerGroupedStats is the service-layer function
func GetOwnerGroupedStats(ownerID int64) ([]VenueStats, error) {
	return GetOwnerVenueStatsGrouped(ownerID)
}


// CalculatePopularTime finds the most frequent hour from a list of times
func CalculatePopularTime(times []time.Time) string {
	if len(times) == 0 {
		return "--:--"
	}

	// 1. Create a frequency map for hours (0-23)
	hourCounts := make(map[int]int)
	
	// 2. Load the timezone for India (IST)
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		// Fallback to Local if timezone data is missing
		loc = time.Local 
	}

	// 3. Count frequencies
	for _, t := range times {
		// Convert UTC to IST in Go memory
		localTime := t.In(loc)
		hourCounts[localTime.Hour()]++
	}

	// 4. Find max
	maxCount := 0
	popularHour := 0
	for h, count := range hourCounts {
		if count > maxCount {
			maxCount = count
			popularHour = h
		}
	}

	// 5. Format
	t := time.Date(0, 1, 1, popularHour, 0, 0, 0, time.UTC)
	return t.Format("03:04 PM")
}
// GetGroupedVenueStats is the service-layer function for the admin
func GetGroupedVenueStats() ([]VenueStats, error) {
	return GetVenueStatsGrouped()
}

// Helper to satisfy compiler if "time" is not explicitly used in logic
func init() {
	_ = time.Duration(0)
}

// ManageBookingAttendance handles OWNER/ADMIN actions
func ManageBookingAttendance(bookingID int64, userID int64, userRole string, action string) error {
	// action can be: 'present', 'absent', 'cancel'

	// 1. Fetch Booking to check current status
	booking, err := FindBookingByID(bookingID)
	if err != nil {
		return errors.New("booking not found")
	}

	var newStatus string

	// 2. Determine Logic
	if action == "present" {
		newStatus = "present"
	} else if action == "absent" {
		// If they didn't show up, we keep it as 'confirmed' (money kept) or mark 'absent'
		// Usually, 'confirmed' is fine, but let's say we mark 'absent' for records.
		newStatus = "absent" // Ensure your DB ENUM allows this, or stick to 'confirmed'
	} else if action == "cancel" {
		// Owner is canceling the booking (e.g., rain, maintenance)
		if booking.Status == "confirmed" {
			newStatus = "refunded" // Owner cancels paid slot -> MUST Refund
		} else {
			newStatus = "canceled"
		}
	} else {
		return errors.New("invalid action")
	}

	// 3. Apply Update
	if userRole == "admin" {
		return UpdateBookingStatusDirect(bookingID, newStatus)
	} else {
		return UpdateBookingStatusByOwner(bookingID, userID, newStatus)
	}
}
