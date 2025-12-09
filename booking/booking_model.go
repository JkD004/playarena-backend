// booking/booking_model.go
package booking

import "time"

type Booking struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	UserFirstName string    `json:"user_first_name"` // Added for joins
	UserLastName  string    `json:"user_last_name"`  // Added for joins
	VenueID       int64     `json:"venue_id"`
	VenueName     string    `json:"venue_name"`     // Added for joins
	VenueAddress  string    `json:"venue_address"`  // Added for joins
	SportCategory string    `json:"sport_category"` // Added for joins
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	TotalPrice    float64   `json:"total_price"`
	Status        string    `json:"status"`
	PaymentID     string    `json:"razorpay_payment_id"` // <--- ADDED THIS FIELD
	CreatedAt     time.Time `json:"created_at"`
}

// Add a struct for the request body, as users won't send everything
type CreateBookingRequest struct {
	VenueID   int64     `json:"venue_id" binding:"required"`
	StartTime time.Time `json:"start_time" binding:"required"`
	EndTime   time.Time `json:"end_time" binding:"required"`
	// Price will be calculated on the backend
}

// AdminBookingView includes venue name and user name
type AdminBookingView struct {
	BookingID     int64     `json:"booking_id"`
	VenueID       int64     `json:"venue_id"`
	VenueName     string    `json:"venue_name"`
	SportCategory string    `json:"sport_category"`
	UserID        int64     `json:"user_id"`
	UserFirstName string    `json:"user_first_name"` // <-- ADD THIS
	UserLastName  string    `json:"user_last_name"`  // <-- ADD THIS
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	TotalPrice    float64   `json:"total_price"`
	Status        string    `json:"status"`
	UserPhone     string    `json:"user_phone"`
}

// OwnerStats defines the data for the owner's dashboard
type OwnerStats struct {
	TotalBookings     int64   `json:"total_bookings"`     // Sum of all
	ConfirmedBookings int64   `json:"confirmed_bookings"` // New
	PresentBookings   int64   `json:"present_bookings"`   // New
	CanceledBookings  int64   `json:"canceled_bookings"`  // New
	RefundedBookings  int64   `json:"refunded_bookings"`  // New
	TotalRevenue      float64 `json:"total_revenue"`
	PopularTime       string  `json:"popular_time"`
}

// VenueStats defines the stats for a single venue
type VenueStats struct {
	VenueID       int64   `json:"venue_id"`
	VenueName     string  `json:"venue_name"`
	SportCategory string  `json:"sport_category"`
	TotalBookings int64   `json:"total_bookings"`
	TotalRevenue  float64 `json:"total_revenue"`
} // booking/booking_model.go

type BookedSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}
