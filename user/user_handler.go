package user

import (
	"context" // <-- Add this
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"github.com/cloudinary/cloudinary-go/v2" // <-- Add this
	"github.com/cloudinary/cloudinary-go/v2/api/uploader" // <-- Add this
)

// Global Cloudinary instance for User package
var cld *cloudinary.Cloudinary

func SetCloudinary(cloudinaryClient *cloudinary.Cloudinary) {
	cld = cloudinaryClient
}

func RegisterUserHandler(c *gin.Context) {
	var user User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := RegisterNewUser(&user)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully!"})
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginUserHandler now returns a token
func LoginUserHandler(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// 1. Get both token and role from the service
	tokenString, userRole, err := LoginUser(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 2. Send them both to the frontend
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful!",
		"token":   tokenString,
		"role":    userRole, // <-- ADD THIS LINE
	})
}
// GetProfileHandler handles fetching the logged-in user's profile
// user/user_handler.go
// ... (keep existing functions)

// GetProfileHandler handles GET /api/v1/profile/me
func GetProfileHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	user, err := GetUserProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfileHandler handles PATCH /api/v1/profile/me
func UpdateProfileHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

	var userUpdates User
	if err := c.ShouldBindJSON(&userUpdates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := UpdateUserProfile(userID, &userUpdates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// UploadProfilePicHandler handles uploading a user avatar
func UploadProfilePicHandler(c *gin.Context) {
	userID := c.MustGet("userID").(int64)

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

	// Upload to Cloudinary folder 'playarena_users'
	ctx := context.Background()
	params := uploader.UploadParams{Folder: "playarena_users"}

	result, err := cld.Upload.Upload(ctx, src, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image"})
		return
	}

	imageURL := result.SecureURL

	// Save to DB
	err = UpdateUserAvatar(userID, imageURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save avatar URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Avatar updated", "url": imageURL})
}

// GetAllUsersHandler handles GET /api/v1/admin/users
func GetAllUsersHandler(c *gin.Context) {
	users, err := GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// DeleteUserHandler handles DELETE /api/v1/admin/users/:id
func DeleteUserHandler(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = RemoveUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// UpdateUserRoleHandler handles PATCH /api/v1/admin/users/:id/role
func UpdateUserRoleHandler(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role is required"})
		return
	}

	err = ChangeUserRole(userID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "User role updated"})
}