// api/router.go
package api

import (
	"github.com/JkD004/playarena-backend/booking"
	"github.com/JkD004/playarena-backend/notification"
	"github.com/JkD004/playarena-backend/payment"
	"github.com/JkD004/playarena-backend/settings"
	"github.com/JkD004/playarena-backend/team"
	"github.com/JkD004/playarena-backend/user"
	"github.com/JkD004/playarena-backend/venue"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {

	// Global Middleware
	router.Use(MaintenanceMiddleware())

	v1 := router.Group("/api/v1")
	{
		// ==========================================
		//           PUBLIC ROUTES (No Auth)
		// ==========================================

		// --- Health Checks ---
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "alive"})
		})
		v1.HEAD("/health", func(c *gin.Context) {
			c.Status(200)
		})

		// --- Authentication ---
		v1.POST("/register", user.RegisterUserHandler)
		v1.POST("/login", user.LoginUserHandler)

		// --- Venues & Slots ---
		v1.GET("/venues", venue.GetVenuesHandler)
		v1.GET("/venues/:id", venue.GetVenueByIDHandler)
		v1.GET("/venues/:id/photos", venue.GetVenuePhotosHandler)
		v1.GET("/venues/:id/slots", booking.GetBookedSlotsHandler)
		v1.GET("/venues/:id/reviews", venue.GetReviewsHandler)

		// --- General ---
		v1.GET("/terms", settings.GetTermsHandler)
		v1.GET("/debug/email", booking.TestLiveEmailHandler)

		// --- Downloads ---
		// Publicly accessible to allow direct email links to work
		v1.GET("/bookings/:id/ticket", booking.DownloadTicketHandler)

		// ==========================================
		//        USER ROUTES (Player, Owner, Admin)
		// ==========================================

		// --- Profile ---
		v1.GET("/profile/me", AuthMiddleware("player", "owner", "admin"), user.GetProfileHandler)
		v1.PATCH("/profile/me", AuthMiddleware("player", "owner", "admin"), user.UpdateProfileHandler)
		v1.POST("/profile/avatar", AuthMiddleware("player", "owner", "admin"), user.UploadProfilePicHandler)

		// --- Booking & Payments ---
		v1.POST("/bookings", AuthMiddleware("player", "owner", "admin"), booking.CreateBookingHandler)
		v1.GET("/bookings/mine", AuthMiddleware("player", "owner", "admin"), booking.GetUserBookingsHandler)
		v1.PATCH("/bookings/:id/cancel", AuthMiddleware("player", "owner", "admin"), booking.CancelBookingHandler)
		
		v1.POST("/bookings/:id/pay", AuthMiddleware("player", "owner", "admin"), booking.ProcessPaymentHandler) // Manual/Test
		v1.POST("/payment/create-order/:id", AuthMiddleware("player", "owner", "admin"), payment.CreateOrderHandler)
		v1.POST("/payment/verify", AuthMiddleware("player", "owner", "admin"), payment.VerifyPaymentHandler)

		// --- Teams & Chat ---
		v1.POST("/teams", AuthMiddleware("player", "owner", "admin"), team.CreateTeamHandler)
		v1.GET("/teams/mine", AuthMiddleware("player", "owner", "admin"), team.GetMyTeamsHandler)
		v1.PATCH("/teams/:id/status", AuthMiddleware("player", "owner", "admin"), team.UpdateMemberStatusHandler)
		v1.POST("/teams/:id/invite", AuthMiddleware("player", "owner", "admin"), team.InviteMemberHandler)
		v1.GET("/teams/:id/members", AuthMiddleware("player", "owner", "admin"), team.GetTeamMembersHandler)
		
		v1.POST("/teams/:id/chat", AuthMiddleware("player", "owner", "admin"), team.PostMessageHandler)
		v1.GET("/teams/:id/chat", AuthMiddleware("player", "owner", "admin"), team.GetTeamMessagesHandler)

		// --- Notifications & Reviews ---
		v1.GET("/notifications", AuthMiddleware("player", "owner", "admin"), notification.GetMyNotificationsHandler)
		v1.PATCH("/notifications/:id/read", AuthMiddleware("player", "owner", "admin"), notification.MarkReadHandler)
		v1.POST("/venues/:id/reviews", AuthMiddleware("player", "owner", "admin"), venue.CreateReviewHandler)

		// ==========================================
		//          OWNER ROUTES (Owner, Admin)
		// ==========================================

		// --- Venue Management ---
		v1.POST("/venues", AuthMiddleware("player", "owner", "admin"), venue.CreateVenueHandler) // Players can become owners by creating
		v1.PUT("/venues/:id", AuthMiddleware("owner", "admin"), venue.UpdateVenueHandler)
		v1.POST("/venues/:id/photos", AuthMiddleware("owner", "admin"), venue.UploadVenuePhotoHandler)
		v1.DELETE("/photos/:id", AuthMiddleware("owner", "admin"), venue.DeleteVenuePhotoHandler)
		v1.GET("/venues/mine", AuthMiddleware("owner", "admin"), venue.GetOwnerVenuesHandler)
		v1.POST("/reviews/:id/reply", AuthMiddleware("owner", "admin"), venue.ReplyReviewHandler)

		// --- Booking Management ---
		v1.GET("/venues/:id/bookings", AuthMiddleware("owner", "admin"), booking.GetVenueBookingsHandler)
		v1.GET("/owner/bookings/:id", AuthMiddleware("owner", "admin"), booking.GetSingleBookingOwnerHandler)
		v1.PATCH("/owner/bookings/:id/status", AuthMiddleware("owner", "admin"), booking.ManageBookingHandler)
		v1.POST("/bookings/block", AuthMiddleware("owner", "admin"), booking.BlockSlotHandler)
		v1.POST("/owner/bookings/:id/refund-decision", AuthMiddleware("owner"), booking.HandleRefundDecisionHandler)

		// --- Stats ---
		v1.GET("/owner/venues/:id/stats", AuthMiddleware("owner", "admin"), booking.GetOwnerStatsHandler)
		v1.GET("/owner/stats/by-venue", AuthMiddleware("owner", "admin"), booking.GetOwnerGroupedStatsHandler)
		v1.GET("/owner/stats/global", AuthMiddleware("owner"), booking.GetOwnerGlobalStatsHandler)

		// ==========================================
		//            ADMIN ROUTES (Admin Only)
		// ==========================================

		// --- User Management ---
		v1.GET("/admin/users", AuthMiddleware("admin"), user.GetAllUsersHandler)
		v1.DELETE("/admin/users/:id", AuthMiddleware("admin"), user.DeleteUserHandler)
		v1.PATCH("/admin/users/:id/role", AuthMiddleware("admin"), user.UpdateUserRoleHandler)

		// --- Venue Management ---
		v1.GET("/admin/venues", AuthMiddleware("admin"), venue.GetVenuesByStatusHandler) // Filter by pending/approved
		v1.GET("/admin/venues/all", AuthMiddleware("admin"), venue.AdminGetAllVenuesHandler)
		v1.GET("/admin/venues/:id/details", AuthMiddleware("admin"), venue.AdminGetVenueDetailHandler)
		v1.PATCH("/admin/venues/:id/status", AuthMiddleware("admin"), venue.UpdateVenueStatusHandler)
		v1.DELETE("/admin/venues/:id", AuthMiddleware("admin"), venue.AdminDeleteVenueHandler)

		// --- Booking & Stats ---
		v1.GET("/admin/bookings", AuthMiddleware("admin"), booking.GetAllBookingsHandler)
		v1.GET("/admin/stats/by-venue", AuthMiddleware("admin"), booking.GetGroupedStatsHandler)
		v1.GET("/admin/stats/global", AuthMiddleware("admin"), booking.GetAdminGlobalStatsHandler)
		v1.GET("/admin/stats/grouped", AuthMiddleware("admin"), booking.GetGroupedStatsHandler) // (Duplicate alias kept for compatibility)

		// --- System ---
		v1.PUT("/admin/terms", AuthMiddleware("admin"), settings.UpdateTermsHandler)
	}
}