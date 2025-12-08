// booking/booking_service.go
package booking

import (
	"errors"
	"log"
	"time"
	"github.com/JkD004/playarena-backend/notification"
	"github.com/JkD004/playarena-backend/venue"
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

// ProcessPayment simulates payment processing
func ProcessPayment(bookingID int64) error {
	// 1. Fetch the booking details first (to get UserID)
	booking, err := FindBookingByID(bookingID)
	if err != nil {
		return errors.New("booking not found")
	}

	// 2. In a real app, verify payment with Stripe/Razorpay here.

	// 3. Update DB status to 'confirmed'
	err = ConfirmBookingPayment(bookingID)
	if err != nil {
		return err
	}

	// 4. Send Notification
	// We now have booking.UserID from step 1
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
func CancelBooking(bookingID int64, userID int64) error {
	// 1. Fetch the booking to check its details
	booking, err := FindBookingByID(bookingID)
	if err != nil {
		return errors.New("booking not found")
	}

	// 2. POLICY CHECK: Cannot cancel past/started bookings
	// We add a small buffer (e.g., can't cancel if it starts in less than 1 hour)
	// For now, let's just say "cannot cancel if already started"
	if time.Now().After(booking.StartTime) {
		return errors.New("cannot cancel a booking that has already started")
	}

    // 3. Update the status in the database
	err = UpdateBookingStatus(bookingID, userID, "canceled")
	if err != nil {
		return err
	}

	// 4. Send Notification
	message := "Your booking has been successfully canceled. A refund will be processed within 5-7 days."
	_ = notification.CreateNotification(userID, message, "warning")

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

// GetStatisticsForOwner is the service-layer function
// GetStatisticsForOwnerOrAdmin handles logic for both roles
func GetStatisticsForOwner(userID int64, venueID int64, userRole string) (*OwnerStats, error) {
	var bookings int64
	var revenue float64
	var popTime string
	var err error

	if userRole == "admin" {
		// Admin: Fetch stats without checking owner_id
		bookings, revenue, err = GetVenueStatsSimple(venueID)
        if err != nil { return nil, err }
        
		popTime, err = GetVenuePopularTimeSimple(venueID)
        if err != nil { return nil, err }

	} else {
		// Owner: Enforce owner_id check
		bookings, revenue, err = GetOwnerBookingStats(userID, venueID)
        if err != nil { return nil, err }
        
		popTime, err = GetOwnerPopularTime(userID, venueID)
        if err != nil { return nil, err }
	}

	return &OwnerStats{
		TotalBookings: bookings,
		TotalRevenue:  revenue,
		PopularTime:   popTime,
	}, nil
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

// GetGroupedVenueStats is the service-layer function for the admin
func GetGroupedVenueStats() ([]VenueStats, error) {
	return GetVenueStatsGrouped()
}

// Helper to satisfy compiler if "time" is not explicitly used in logic
func init() {
	_ = time.Duration(0)
}

// ManageBookingAttendance handles owner/admin actions
// Added 'userRole' parameter
func ManageBookingAttendance(bookingID int64, userID int64, userRole string, status string) error {
	// Validate status
	if status != "present" && status != "absent" && status != "canceled" {
		return errors.New("invalid status update")
	}

	// Logic: If Admin, bypass ownership check. If Owner, enforce it.
	if userRole == "admin" {
		return UpdateBookingStatusDirect(bookingID, status)
	} else {
		return UpdateBookingStatusByOwner(bookingID, userID, status)
	}
}

