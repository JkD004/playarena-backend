// payment/payment_service.go
package payment

import (
	"errors"
	"os"

	"github.com/razorpay/razorpay-go"
)

var client *razorpay.Client

// InitRazorpay initializes the global Razorpay client
func InitRazorpay() {
	client = razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))
}

// CreateRazorpayOrder creates an order ID for the frontend checkout
func CreateRazorpayOrder(bookingID int64, amount float64) (string, error) {
	// Razorpay expects amount in paise (multiply by 100)
	amountInPaise := int(amount * 100)

	data := map[string]interface{}{
		"amount":   amountInPaise,
		"currency": "INR",
		"receipt":  "receipt_booking_" + string(rune(bookingID)), // simple ID conversion
		"payment_capture": 1,
	}

	body, err := client.Order.Create(data, nil)
	if err != nil {
		return "", errors.New("failed to create razorpay order: " + err.Error())
	}

	orderID, ok := body["id"].(string)
	if !ok {
		return "", errors.New("invalid response from razorpay")
	}

	return orderID, nil
}


// NOTE: Refund logic has moved to the 'gateway' package.
// NOTE: Verification logic has moved to 'payment_handler.go' using 'gateway'.