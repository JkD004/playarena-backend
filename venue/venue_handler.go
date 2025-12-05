// venue/venue_handler.go
package venue

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"

	"github.com/JkD004/playarena-backend/db"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

var cld *cloudinary.Cloudinary

func SetCloudinary(cloudinaryClient *cloudinary.Cloudinary) {
	cld = cloudinaryClient
}

// -------------------------------------------------------
// CREATE VENUE
// -------------------------------------------------------
func CreateVenueHandler(c *gin.Context) {
	var venue Venue

	if err := c.ShouldBindJSON(&venue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userID := c.MustGet("userID").(int64)

	err := CreateNewVenue(&venue, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create venue"})
		return
	}

	c.JSON(http.StatusCreated, venue)
}

// -------------------------------------------------------
// GET ALL VENUES (admin)
// -------------------------------------------------------
func GetVenuesHandler(c *gin.Context) {
	venues, err := GetAllVenues()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch venues"})
		return
	}
	c.JSON(http.StatusOK, venues)
}

// -------------------------------------------------------
// GET VENUES BY STATUS (admin)
// -------------------------------------------------------
func GetVenuesByStatusHandler(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")

	venues, err := GetVenuesByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch venues"})
		return
	}
	c.JSON(http.StatusOK, venues)
}

// -------------------------------------------------------
// UPDATE VENUE STATUS (admin)
// -------------------------------------------------------
func UpdateVenueStatusHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body, 'status' is required"})
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be 'approved' or 'rejected'"})
		return
	}

	err = UpdateVenueStatus(venueID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update venue status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Venue status updated successfully"})
}

// -------------------------------------------------------
// GET VENUE BY ID (public or owner)
// -------------------------------------------------------
func GetVenueByIDHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	venue, err := GetVenueByID(venueID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found or not approved"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch venue"})
		return
	}

	c.JSON(http.StatusOK, venue)
}

// -------------------------------------------------------
// UPLOAD VENUE PHOTO
// -------------------------------------------------------
func UploadVenuePhotoHandler(c *gin.Context) {
	venueIDStr := c.Param("id")
	venueID, _ := strconv.ParseInt(venueIDStr, 10, 64)

	// 🔒 SECURITY CHECK
	userID := c.MustGet("userID").(int64)
	userRole := c.MustGet("userRole").(string)

	if userRole != "admin" {
		if err := VerifyVenueOwnership(venueID, userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
	}

	// UPLOAD LOGIC
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image file is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		return
	}
	defer src.Close()

	ctx := context.Background()
	params := uploader.UploadParams{
		Folder: "playarena_venues",
	}

	result, err := cld.Upload.Upload(ctx, src, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	imageURL := result.SecureURL

	_, err = db.DB.Exec("INSERT INTO venue_photos (venue_id, image_url) VALUES (?, ?)", venueID, imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store image URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Photo uploaded successfully",
		"url":     imageURL,
	})
}

// -------------------------------------------------------
// GET VENUE PHOTOS
// -------------------------------------------------------
func GetVenuePhotosHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	photos, err := GetVenuePhotos(venueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch photos"})
		return
	}
	c.JSON(http.StatusOK, photos)
}

// -------------------------------------------------------
// DELETE VENUE PHOTO
// -------------------------------------------------------
// venue/venue_handler.go

// DeleteVenuePhotoHandler handles DELETE /api/v1/photos/:id
func DeleteVenuePhotoHandler(c *gin.Context) {
	photoID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid photo ID"})
		return
	}

	userID := c.MustGet("userID").(int64)
	userRole := c.MustGet("userRole").(string)

	// --- SECURITY CHECK ---
	// If not admin, check ownership
	if userRole != "admin" {
		// 1. Find which venue this photo belongs to
		venueID, err := GetVenueIdFromPhoto(photoID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Photo not found"})
			return
		}

		// 2. Check if the user owns that venue
		if err := VerifyVenueOwnership(venueID, userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to delete this photo"})
			return
		}
	}
	// ----------------------

	err = DeleteVenuePhoto(photoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Photo deleted successfully"})
}

// -------------------------------------------------------
// OWNER'S VENUE LIST
// -------------------------------------------------------
func GetOwnerVenuesHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	venues, err := GetVenuesForOwner(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch your venues"})
		return
	}

	c.JSON(http.StatusOK, venues)
}

// -------------------------------------------------------
// REVIEWS
// -------------------------------------------------------
func CreateReviewHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}
	userID := c.MustGet("userID").(int64)

	var req struct {
		Rating  int    `json:"rating" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating is required"})
		return
	}

	err = AddReview(venueID, userID, req.Rating, req.Comment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit review"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Review submitted"})
}

func GetReviewsHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	reviews, err := GetVenueReviews(venueID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch reviews"})
		return
	}
	c.JSON(http.StatusOK, reviews)
}

// -------------------------------------------------------
// UPDATE VENUE (OWNER OR ADMIN)
// -------------------------------------------------------
func UpdateVenueHandler(c *gin.Context) {
	venueID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid venue ID"})
		return
	}

	// 🔒 SECURITY CHECK
	userID := c.MustGet("userID").(int64)
	userRole := c.MustGet("userRole").(string)

	if userRole != "admin" {
		if err := VerifyVenueOwnership(venueID, userID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
	}

	var venue Venue
	if err := c.ShouldBindJSON(&venue); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err = ModifyVenue(venueID, &venue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update venue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Venue updated successfully"})
}


// venue/venue_handler.go

// ReplyReviewHandler handles POST /api/v1/reviews/:id/reply
func ReplyReviewHandler(c *gin.Context) {
	reviewID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	var req struct {
		Reply string `json:"reply" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reply content is required"})
		return
	}

	err = ReplyToReview(reviewID, req.Reply)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save reply"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reply posted successfully"})
}