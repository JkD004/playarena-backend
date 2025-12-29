// booking/booking_handler.go
package booking

import (
	//"github.com/JkD004/playarena-backend/venue"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JkD004/playarena-backend/gateway"
	"github.com/JkD004/playarena-backend/notification"
	"github.com/JkD004/playarena-backend/pkg/utils"
	"github.com/JkD004/playarena-backend/user"
	"github.com/JkD004/playarena-backend/venue"
)

// booking/booking_handler.go

func TestLiveEmailHandler(c *gin.Context) {
	// 1. Check if API Key is loaded
	apiKey := os.Getenv("RESEND_API_KEY")

	maskedKey := "NOT SET"
	if len(apiKey) > 5 {
		maskedKey = apiKey[:3] + "..." + apiKey[len(apiKey)-3:]
	}

	debugInfo := fmt.Sprintf("API Key: %s", maskedKey)
	fmt.Println("Debug Info:", debugInfo)

	// 2. Define the recipient explicitly (Since we deleted SMTP_EMAIL)
	// MUST match your Resend login email for the free tier
	recipient := "thesportgrid@gmail.com"

	subject := "Test Email from Live Server (Resend)"
	body := "<h1>It Works!</h1><p>Your Resend API is configured correctly.</p>"

	// 3. Try to send
	err := notification.SendEmail(recipient, subject, body)

	if err != nil {
		c.JSON(500, gin.H{
			"status":        "failed",
			"error":         err.Error(),
			"config_loaded": debugInfo,
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "success",
		"message": "Email sent successfully to " + recipient,
		"config":  debugInfo,
	})
}

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

	// ---------------- EMAIL LOGIC ----------------

	// Fetch user details
	userData, err := user.GetUserByID(userID)
	if err != nil {
		fmt.Println("Failed to fetch user:", err)
	} else {

		// Fetch venue details
		venueData, err := venue.GetVenueByID(newBooking.VenueID)
		if err != nil {
			fmt.Println("Failed to fetch venue:", err)
		} else {

			// ================================
			// 1. FORCE IST TIMEZONE (FIXED ZONE)
			// ================================
			// Instead of loading from OS (which might fail on Render),
			// we hardcode the 5hr 30min offset (19800 seconds).
			secondsEastOfUTC := int((5 * 3600) + (30 * 60))
			loc := time.FixedZone("IST", secondsEastOfUTC)

			// ================================
			// 2. CONVERT TIMES TO IST
			// ================================
			// Use the Scheduled Start Time for the "Date" field to ensure accuracy
			startTime := newBooking.StartTime.In(loc)
			endTime := newBooking.EndTime.In(loc)

			// ================================
			// 3. FORMAT STRINGS
			// ================================
			dateStr := startTime.Format("02 Jan 2006")
			startTimeStr := startTime.Format("03:04 PM")
			endTimeStr := endTime.Format("03:04 PM")

			// ================================
			// 4. DOWNLOAD LINK
			// ================================
			//baseURL := "https://playarena-frontend.vercel.app"
			
			baseURL := "http://localhost:3000"


			downloadLink := fmt.Sprintf(
				"%s/bookings/%d/ticket?download=true",
				baseURL,
				newBooking.ID,
			)

			// ================================
			// 5. EMAIL CONTENT
			// ================================
			subject := "Booking Confirmed! - SportGrid"

			body := fmt.Sprintf(`
<h1>Booking Confirmed! ✅</h1>

<p>Hi %s,</p>

<p>Your booking at <strong>%s</strong> is confirmed.</p>

<p><strong>Date:</strong> %s</p>
<p><strong>Time:</strong> %s - %s</p>

<p><strong>Total Price:</strong> ₹%.2f</p>

<br>

<a href="%s" style="background-color:#008CBA;color:white;padding:10px 20px;
text-decoration:none;border-radius:5px;font-weight:bold;">
Download Ticket
</a>

<br><br>
<p>If the button doesn't work, click here:</p>
<p><a href="%s">%s</a></p>

<br>
<p>Thank you for choosing <b>SportGrid</b>!</p>
`,
				userData.FirstName,
				venueData.Name,
				dateStr,
				startTimeStr,
				endTimeStr,
				newBooking.TotalPrice,
				downloadLink,
				downloadLink,
				downloadLink,
			)

			// ================================
			// 6. SEND EMAIL ASYNC
			// ================================
			go func() {
				if err := notification.SendEmail(userData.Email, subject, body); err != nil {
					fmt.Println("Email send failed:", err)
				}
			}()
		}
	}

	// ---------------- RESPONSE ----------------
	c.JSON(http.StatusCreated, newBooking)
}

// VerifyTicketHandler checks if a scanned QR code is valid
func VerifyTicketHandler(c *gin.Context) {
	var req struct {
		QRCodeString string `json:"qr_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"valid": false, "message": "Invalid Request"})
		return
	}

	// 1. Split the content "BookingID|Signature"
	parts := strings.Split(req.QRCodeString, "|")
	if len(parts) != 2 {
		c.JSON(http.StatusOK, gin.H{"valid": false, "message": "Fake Ticket: Invalid Format"})
		return
	}

	bookingIDStr := parts[0]
	providedSignature := parts[1]

	// 2. Re-create the signature using OUR secret
	secretKey := os.Getenv("TICKET_SECRET_KEY")
	if secretKey == "" {
		secretKey = "default_unsafe_secret"
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(bookingIDStr))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// 3. Compare signatures
	// We use subtle.ConstantTimeCompare to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(providedSignature), []byte(expectedSignature)) != 1 {
		c.JSON(http.StatusOK, gin.H{"valid": false, "message": "Fake Ticket: Signature Mismatch"})
		return
	}

	// 4. (Optional) Check if booking exists in DB and is for today
	// ... Database logic here ...

	c.JSON(http.StatusOK, gin.H{
		"valid":      true,
		"booking_id": bookingIDStr,
		"message":    "Ticket Verified Successfully ✅",
	})
}

// DownloadTicketHandler generates and serves the PDF
func DownloadTicketHandler(c *gin.Context) {
	bookingID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	// 1. Fetch Booking
	booking, err := FindBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// 2. Fetch User and Venue Details
	venueData, _ := venue.GetVenueByID(booking.VenueID)
	userData, _ := user.GetUserByID(booking.UserID)

	// 3. Generate PDF
	// Ensure you match the parameters exactly as defined in pdf_generator.go
	pdf, err := utils.GenerateTicketPDF(
		userData.FirstName, // userName
		userData.LastName,  // userLastName
		venueData.Name,     // venueName
		venueData.Address,  // venueAddress
		booking.ID,         // bookingID
		userData.ID,        // userID
		booking.StartTime,  // startTime
		booking.EndTime,    // endTime
		booking.CreatedAt,  // bookingCreated
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate ticket: " + err.Error()})
		return
	}

	// 4. Create Filename and Download
	fileName := fmt.Sprintf("TICKET-U%d-V%d-B%d.pdf", userData.ID, venueData.ID, booking.ID)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Content-Type", "application/pdf")

	err = pdf.Output(c.Writer)
	if err != nil {
		fmt.Println("Error outputting PDF:", err)
	}
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

	// ----------------------------------------------------
	// NEW: Generate Signed QR Codes for each booking
	// ----------------------------------------------------
	secretKey := os.Getenv("TICKET_SECRET_KEY")
	if secretKey == "" {
		secretKey = "default_unsafe_secret"
	}

	// Loop through each booking and sign it
	for i := range bookings {
		// 1. Create Data
		data := fmt.Sprintf("%d", bookings[i].ID)

		// 2. Create Signature
		h := hmac.New(sha256.New, []byte(secretKey))
		h.Write([]byte(data))
		signature := hex.EncodeToString(h.Sum(nil))

		// 3. Attach to the struct
		bookings[i].QRCode = fmt.Sprintf("%s|%s", data, signature)
	}
	// ----------------------------------------------------

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

// ProcessPaymentHandler handles the payment request
// ProcessPaymentHandler (Used for testing/manual payment simulation)
func ProcessPaymentHandler(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	// FIX: Pass a placeholder string as the second argument
	// Real Razorpay payments use the VerifyPaymentHandler in the payment package.
	err = ProcessPayment(id, "simulated_payment_id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment processed successfully"})
}

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

// GetSingleBookingOwnerHandler handles GET /api/v1/owner/bookings/:id
func GetSingleBookingOwnerHandler(c *gin.Context) {
	bookingID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid booking ID"})
		return
	}

	ownerID := c.MustGet("userID").(int64)

	bookingDetails, err := GetBookingDetailsForOwner(bookingID, ownerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, bookingDetails)
}

// booking/booking_handler.go

// GetOwnerGlobalStatsHandler: /api/v1/owner/stats/global
func GetOwnerGlobalStatsHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)
	stats, err := GetGlobalStatsForOwner(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch global stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// GetAdminGlobalStatsHandler: /api/v1/admin/stats/global
func GetAdminGlobalStatsHandler(c *gin.Context) {
	stats, err := GetGlobalStatsForPlatform()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch platform stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// HandleRefundDecisionHandler allows owner to Approve/Reject refunds
func HandleRefundDecisionHandler(c *gin.Context) {
	bookingID, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var req struct {
		Decision string `json:"decision"` // 'approve' or 'reject'
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 1. Fetch Booking
	// FIX: Removed 'booking.' prefix (we are already in package booking)
	b, err := FindBookingByID(bookingID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Booking not found"})
		return
	}

	// 2. Validate state
	if b.Status != "refund_requested" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This booking is not pending a refund request"})
		return
	}

	userID := b.UserID
	var newStatus string
	var msg string

	if req.Decision == "approve" {
		// --- CALL RAZORPAY ---
		if b.PaymentID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing payment ID"})
			return
		}
		// FIX: 'gateway' is now imported correctly
		err := gateway.InitiateRefund(b.PaymentID, b.TotalPrice)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Razorpay Refund Failed: " + err.Error()})
			return
		}
		newStatus = "refunded"
		msg = "Your refund request has been APPROVED. Money sent to source."

	} else if req.Decision == "reject" {
		// --- NO REFUND ---
		newStatus = "refund_rejected"
		msg = "Your refund request has been REJECTED by the venue owner."

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid decision"})
		return
	}

	// 3. Update Status
	// FIX: Removed 'booking.' prefix
	err = UpdateBookingStatusDirect(bookingID, newStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	// 4. Notify Player
	// FIX: 'notification' is now imported correctly
	_ = notification.CreateNotification(userID, msg, "info")

	c.JSON(http.StatusOK, gin.H{"message": "Refund decision processed", "status": newStatus})
}
