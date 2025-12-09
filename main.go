package main

import (
	"log"
	
	"time" // Added for time.Sleep

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/cloudinary/cloudinary-go/v2"

	"github.com/JkD004/playarena-backend/api"
	"github.com/JkD004/playarena-backend/db"
	"github.com/JkD004/playarena-backend/venue"
	"github.com/JkD004/playarena-backend/user"
	"github.com/JkD004/playarena-backend/payment"
	"github.com/JkD004/playarena-backend/worker"

)

func main() {

	// ✅ Load environment variables (.env)
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  Warning: .env file not found, continuing...")
	}

	// ✅ Initialize Database with Retry Logic (CRITICAL FIX)
	const maxRetries = 15 // Increased retries for slow DB initialization
	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to DB... (Attempt %d/%d)", i+1, maxRetries)
		// Assuming db.InitDB() now returns an error instead of calling log.Fatal
		if err := db.InitDB(); err == nil {
			log.Println("✅ Database connection established and initialized.")
			break // Success, exit loop
		}

		if i == maxRetries-1 {
			log.Fatal("❌ Failed to connect to database after multiple retries. Exiting.")
		}

		// Exponential backoff: 1s, 2s, 4s, 8s...
		wait := time.Duration(1<<(i)) * time.Second
		if wait > 30*time.Second { // Cap waiting time
			wait = 30 * time.Second
		}
		log.Printf("Database not ready, waiting for %v before retrying...", wait)
		time.Sleep(wait)
	}

	// The rest of the application setup relies on the database being connected successfully.

	// ✅ Initialize Cloudinary
	cld, err := cloudinary.New()
	if err != nil {
		// Log.Fatal will exit the application if Cloudinary fails, which is correct behavior.
		log.Fatalf("❌ Failed to initialize Cloudinary: %v", err)
	}

	// ✅ Pass Cloudinary into relevant modules
	venue.SetCloudinary(cld)
	user.SetCloudinary(cld)

	// ✅ Setup Gin Router
	router := gin.Default()

	// ✅ Initialize Payment System
    payment.InitRazorpay()

	// 🚀 START BACKGROUND WORKER (Run in a separate goroutine)
	go worker.StartCleanupTask()


	// ✅ CORS Configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true 
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	router.Use(cors.New(config))

	// ✅ Set up routes
	api.SetupRoutes(router)

	// ✅ Run Server
	router.Run(":8080")
}