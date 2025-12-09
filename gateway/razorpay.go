package gateway

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

	"github.com/razorpay/razorpay-go"
)

// InitiateRefund processes a refund via Razorpay
func InitiateRefund(paymentID string, amount float64) error {
	client := razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

	// Razorpay expects amount in paise (int)
	refundAmount := int(amount * 100)

	data := map[string]interface{}{
		"amount": refundAmount,
		"speed":  "normal",
	}

	// FIX: Provide all 4 arguments required by the SDK: 
	// 1. Payment ID (string)
	// 2. Amount (int)
	// 3. Data (map)
	// 4. Headers (map[string]string - nil is fine)
	_, err := client.Payment.Refund(paymentID, refundAmount, data, nil)
	if err != nil {
		return errors.New("failed to process refund: " + err.Error())
	}

	return nil
}

// VerifySignature checks if the Razorpay payment is legitimate
func VerifySignature(orderID, paymentID, signature string) bool {
	secret := os.Getenv("RAZORPAY_KEY_SECRET")
	
	// Create the message to sign
	data := orderID + "|" + paymentID

	// Generate HMAC SHA256 signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	generatedSignature := hex.EncodeToString(h.Sum(nil))

	return generatedSignature == signature
}