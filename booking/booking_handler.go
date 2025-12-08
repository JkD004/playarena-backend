// booking/booking_handler.go
package booking

import (
	//"github.com/JkD004/playarena-backend/venue"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

// CreateBookingHandler handles POST requests to create a booking
func CreateBookingHandler(c *gin.Context) {
	var req CreateBookingRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userID := c.MustGet("userID").(int64)

	newBooking, err := CreateNewBooking(&req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, newBooking)
}

// GetUserBookingsHandler handles fetching all bookings for the logged-in user
func GetUserBookingsHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	bookings, err := GetBookingsForUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch your bookings"})
		return
	}

	if bookings == nil {
		bookings = make([]Booking, 0)
	}

	c.JSON(http.StatusOK, bookings)
}

// CancelBookingHandler handles canceling a booking
func CancelBookingHandler(c *gin.Context) {
	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	userID := c.MustGet("userID").(int64)

	// --- THIS IS THE FIX ---
	// Call CancelBooking, not CancelUserBooking
	err = CancelBooking(bookingID, userID)
	// ---------------------
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking canceled successfully"})
}

// GetAllBookingsHandler handles the admin request to get all bookings
func GetAllBookingsHandler(c *gin.Context) {
	bookings, err := GetAllBookings()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch bookings"})
		return
	}
	c.JSON(http.StatusOK, bookings)
}

// GetVenueBookingsHandler handles fetching all bookings for a specific venue
func GetVenueBookingsHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	// TODO: Verify owner (from c.MustGet("userID")) owns this venueID

	bookings, err := GetBookingsForVenue(venueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch bookings"})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

// GetOwnerStatsHandler handles fetching all stats for the logged-in owner
func GetOwnerStatsHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	userRole := c.MustGet("userRole").(string) // <-- Get Role

	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	// Pass role to service
	stats, err := GetStatisticsForOwner(userID, venueID, userRole)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not calculate statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetAdminStatsHandler handles fetching all stats for the admin
func GetAdminStatsHandler(c *gin.Context) {
	stats, err := GetStatisticsForAdmin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not calculate statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetGroupedStatsHandler handles the admin request to get stats by venue
func GetGroupedStatsHandler(c *gin.Context) {
	stats, err := GetGroupedVenueStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not calculate statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// BlockSlotHandler handles owner requests to block a time slot
func BlockSlotHandler(c *gin.Context) {
	var req CreateBookingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userID := c.MustGet("userID").(int64)

	// Call service (we reuse CreateNewBooking but with a flag or logic)
	// Ideally, we create a specific service function for this.
	// For simplicity, let's call a new service function:
	err := BlockVenueSlot(&req, userID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Slot blocked successfully"})
}

// booking/booking_handler.go

func GetOwnerGroupedStatsHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	stats, err := GetOwnerGroupedStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not calculate statistics"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// booking/booking_handler.go
// ... (keep all other handlers)

// ProcessPaymentHandler handles the payment request
func ProcessPaymentHandler(c *gin.Context) {
	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	// TODO: You might want to check if the user owns this booking

	err = ProcessPayment(bookingID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Payment failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment successful, booking confirmed!"})
}

// booking/booking_handler.go

// GetBookedSlotsHandler handles GET /api/v1/venues/:id/slots?date=YYYY-MM-DD
func GetBookedSlotsHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date query param required (YYYY-MM-DD)"})
		return
	}

	// We skip the service layer for this simple read-only query to keep it quick
	slots, err := GetBookedSlotsForDate(venueID, dateStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch slots"})
		return
	}

	c.JSON(http.StatusOK, slots)
}

// ManageBookingHandler handles PATCH /api/v1/owner/bookings/:id/status
func ManageBookingHandler(c *gin.Context) {
	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	userID := c.MustGet("userID").(int64)
	userRole := c.MustGet("userRole").(string) // <-- Get Role

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status is required"})
		return
	}

	// Pass userRole to service
	err = ManageBookingAttendance(bookingID, userID, userRole, req.Status)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Booking status updated successfully"})
}
