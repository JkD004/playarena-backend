// worker/cleanup.go
package worker

import (
	"log"
	"time"

	"github.com/JkD004/playarena-backend/booking"
)

// StartCleanupTask runs a background loop to clean up old bookings
func StartCleanupTask() {
	// Ticker triggers every 1 minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run forever in the background
	for range ticker.C {
		// Cancel bookings pending for more than 10 minutes
		rowsAffected, err := booking.AutoCancelPendingBookings(10)
		if err != nil {
			log.Println("❌ Error running cleanup task:", err)
		} else if rowsAffected > 0 {
			log.Printf("🧹 Cleanup: Canceled %d expired pending bookings.\n", rowsAffected)
		}
	}
}