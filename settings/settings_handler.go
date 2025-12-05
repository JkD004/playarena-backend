package settings

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// GetTermsHandler (Public)
func GetTermsHandler(c *gin.Context) {
	terms, err := GetSetting("terms")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not fetch terms"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": terms})
}

// UpdateTermsHandler (Admin Only)
func UpdateTermsHandler(c *gin.Context) {
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content is required"})
		return
	}

	err := UpdateSetting("terms", req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update terms"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Terms updated successfully"})
}