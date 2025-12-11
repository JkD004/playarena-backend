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

	v1 := router.Group("/api/v1")
	{
		// === Public Routes (No Auth Needed) ===
		v1.POST("/register", user.RegisterUserHandler)
		v1.POST("/login", user.LoginUserHandler)
		v1.GET("/venues", venue.GetVenuesHandler)
		v1.GET("/venues/:id", venue.GetVenueByIDHandler)
		v1.GET("/venues/:id/photos", venue.GetVenuePhotosHandler) // Public can see photos

		// === Admin-Only Routes ===
		v1.GET("/admin/venues", AuthMiddleware("admin"), venue.GetVenuesByStatusHandler)
		v1.PATCH("/admin/venues/:id/status", AuthMiddleware("admin"), venue.UpdateVenueStatusHandler)
		v1.GET("/admin/bookings", AuthMiddleware("admin"), booking.GetAllBookingsHandler)

		// === Owner & Admin Routes ===
		v1.POST("/venues", AuthMiddleware("player", "owner", "admin"), venue.CreateVenueHandler)
		v1.POST("/venues/:id/photos", AuthMiddleware("owner", "admin"), venue.UploadVenuePhotoHandler)
		v1.DELETE("/photos/:id", AuthMiddleware("owner", "admin"), venue.DeleteVenuePhotoHandler)

		// === All Logged-in Users (Player, Owner, Admin) ===
		v1.POST("/bookings", AuthMiddleware("player", "owner", "admin"), booking.CreateBookingHandler)
		v1.POST("/bookings/:id/pay", AuthMiddleware("player", "owner", "admin"), booking.ProcessPaymentHandler)
		v1.GET("/bookings/mine", AuthMiddleware("player", "owner", "admin"), booking.GetUserBookingsHandler)
		v1.PATCH("/bookings/:id/cancel", AuthMiddleware("player", "owner", "admin"), booking.CancelBookingHandler)

		// === Team Routes ===
		// (We'll use "player", "owner", "admin" for now on team routes for simplicity)
		v1.POST("/teams", AuthMiddleware("player", "owner", "admin"), team.CreateTeamHandler)
		v1.GET("/teams/mine", AuthMiddleware("player", "owner", "admin"), team.GetMyTeamsHandler)
		v1.PATCH("/teams/:id/status", AuthMiddleware("player", "owner", "admin"), team.UpdateMemberStatusHandler)
		v1.POST("/teams/:id/invite", AuthMiddleware("player", "owner", "admin"), team.InviteMemberHandler)
		v1.GET("/teams/:id/members", AuthMiddleware("player", "owner", "admin"), team.GetTeamMembersHandler)

		v1.GET("/venues/mine", AuthMiddleware("owner", "admin"), venue.GetOwnerVenuesHandler)

		v1.GET("/venues/:id/bookings", AuthMiddleware("owner", "admin"), booking.GetVenueBookingsHandler)

		v1.GET("/owner/venues/:id/stats", AuthMiddleware("owner", "admin"), booking.GetOwnerStatsHandler)

		v1.GET("/admin/stats/by-venue", AuthMiddleware("admin"), booking.GetGroupedStatsHandler)
		// We will add the chat and profile routes here once we build their handlers.

		v1.POST("/bookings/block", AuthMiddleware("owner", "admin"), booking.BlockSlotHandler)

		v1.GET("/owner/stats/by-venue", AuthMiddleware("owner", "admin"), booking.GetOwnerGroupedStatsHandler)
		v1.POST("/teams/:id/chat", AuthMiddleware("player", "owner", "admin"), team.PostMessageHandler)
		v1.GET("/teams/:id/chat", AuthMiddleware("player", "owner", "admin"), team.GetTeamMessagesHandler)

		v1.GET("/profile/me", AuthMiddleware("player", "owner", "admin"), user.GetProfileHandler)

		v1.PATCH("/profile/me", AuthMiddleware("player", "owner", "admin"), user.UpdateProfileHandler)
		v1.GET("/venues/:id/reviews", venue.GetReviewsHandler) // Anyone can read reviews

		v1.POST("/venues/:id/reviews", AuthMiddleware("player", "owner", "admin"), venue.CreateReviewHandler)

		v1.PUT("/venues/:id", AuthMiddleware("owner", "admin"), venue.UpdateVenueHandler)

		v1.POST("/profile/avatar", AuthMiddleware("player", "owner", "admin"), user.UploadProfilePicHandler)
		v1.GET("/notifications", AuthMiddleware("player", "owner", "admin"), notification.GetMyNotificationsHandler)
		v1.PATCH("/notifications/:id/read", AuthMiddleware("player", "owner", "admin"), notification.MarkReadHandler)
		v1.GET("/venues/:id/slots", booking.GetBookedSlotsHandler)

		v1.POST("/reviews/:id/reply", AuthMiddleware("owner", "admin"), venue.ReplyReviewHandler)

		// Inside v1 group -> Admin-Only Routes
		v1.GET("/admin/users", AuthMiddleware("admin"), user.GetAllUsersHandler)
		v1.DELETE("/admin/users/:id", AuthMiddleware("admin"), user.DeleteUserHandler)
		v1.PATCH("/admin/users/:id/role", AuthMiddleware("admin"), user.UpdateUserRoleHandler)

		// --- Public Routes ---
		v1.GET("/terms", settings.GetTermsHandler) // Public access to terms

		// --- Admin-Only Routes ---
		v1.PUT("/admin/terms", AuthMiddleware("admin"), settings.UpdateTermsHandler) // Admin update

		v1.PATCH("/owner/bookings/:id/status", AuthMiddleware("owner", "admin"), booking.ManageBookingHandler)

		// === Payment Routes (Protected) ===
		v1.POST("/payment/create-order/:id", AuthMiddleware("player", "owner", "admin"), payment.CreateOrderHandler)
		v1.POST("/payment/verify", AuthMiddleware("player", "owner", "admin"), payment.VerifyPaymentHandler)

		v1.GET("/owner/bookings/:id", AuthMiddleware("owner", "admin"), booking.GetSingleBookingOwnerHandler)

		// Inside v1 group -> Owner Routes
		v1.GET("/owner/stats/global", AuthMiddleware("owner"), booking.GetOwnerGlobalStatsHandler)

		// Inside v1 group -> Admin Routes
		v1.GET("/admin/stats/global", AuthMiddleware("admin"), booking.GetAdminGlobalStatsHandler)

		v1.GET("/admin/stats/grouped", AuthMiddleware("admin"), booking.GetGroupedStatsHandler)

		v1.GET("/admin/venues/all", AuthMiddleware("admin"), venue.AdminGetAllVenuesHandler)
		v1.GET("/admin/venues/:id/details", AuthMiddleware("admin"), venue.AdminGetVenueDetailHandler)

		v1.DELETE("/admin/venues/:id", AuthMiddleware("admin"), venue.AdminDeleteVenueHandler)

		v1.POST("/owner/bookings/:id/refund-decision", AuthMiddleware("owner"), booking.HandleRefundDecisionHandler)

	}
}
