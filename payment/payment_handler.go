// payment/payment_handler.go
package payment

import (
	"net/http"
	"os"
	"strconv"
	"time" // <--- Import time

	"github.com/JkD004/playarena-backend/booking"
	"github.com/JkD004/playarena-backend/gateway"
	"github.com/gin-gonic/gin"
)

// CreateOrderHandler: Frontend asks to start payment process
func CreateOrderHandler(c *gin.Context) {
	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	// 1. Fetch Booking
	b, err := booking.FindBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// ---------------------------------------------------------
	// 🛑 SECURITY CHECK 1: Is the slot in the past?
	// ---------------------------------------------------------
	// We add a small 5-minute buffer to be kind, but generally no.
	if time.Now().After(b.StartTime.Add(5 * time.Minute)) {
		c.JSON(http.StatusConflict, gin.H{"error": "This booking slot has expired. Please book a new slot."})
		return
	}

    // ---------------------------------------------------------
	// 🛑 SECURITY CHECK 2: Is it already paid?
	// ---------------------------------------------------------
    if b.Status == "confirmed" || b.Status == "present" {
        c.JSON(http.StatusConflict, gin.H{"error": "This booking is already paid for."})
		return
    }
    
    // ---------------------------------------------------------
	// 🛑 SECURITY CHECK 3: Is it canceled?
	// ---------------------------------------------------------
    if b.Status == "canceled" {
        c.JSON(http.StatusConflict, gin.H{"error": "This booking was canceled. You cannot pay for it."})
		return
    }

	// 2. Create Razorpay Order
	orderID, err := CreateRazorpayOrder(b.ID, b.TotalPrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 3. Send Order ID
	c.JSON(http.StatusOK, gin.H{
		"order_id": orderID,
		"amount":   b.TotalPrice,
		"key_id":   os.Getenv("RAZORPAY_KEY_ID"),
	})
}



// VerifyPaymentHandler: Frontend sends success data to verify
// VerifyPaymentHandler
func VerifyPaymentHandler(c *gin.Context) {
	var req struct {
		BookingID         int64  `json:"booking_id"`
		RazorpayOrderID   string `json:"razorpay_order_id"`
		RazorpayPaymentID string `json:"razorpay_payment_id"`
		RazorpaySignature string `json:"razorpay_signature"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// FIX: Use the function from the gateway package
	if !gateway.VerifySignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid payment signature"})
		return
	}

	// Confirm Booking
	err := booking.ConfirmBookingPayment(req.BookingID, req.RazorpayPaymentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB update failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Booking confirmed"})
}