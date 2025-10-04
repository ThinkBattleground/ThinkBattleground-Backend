package controllers

import (
	"net/http"

	"github.com/ThinkBattleground/ThinkBattleground-Backend/database"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/models"
	"github.com/gin-gonic/gin"
)

type AdminController struct{}

// MakeUserAdmin promotes a user to admin status
func (ac *AdminController) MakeUserAdmin(c *gin.Context) {
	var input struct {
		UserID string `json:"userId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("firebase_uid = ?", input.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update user to admin
	user.IsAdmin = true
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User promoted to admin successfully",
		"user": gin.H{
			"id":      user.ID,
			"email":   user.Email,
			"isAdmin": user.IsAdmin,
		},
	})
}

// UpdateUserProfile updates additional user information
func (ac *AuthController) UpdateUserProfile(c *gin.Context) {
	var input struct {
		UserID      string  `json:"userId"`
		DisplayName *string `json:"displayName"`
		Phone       *string `json:"phone"`
		Country     *string `json:"country"`
		Bio         *string `json:"bio"`
	}
	userId, exists := c.Get("userId")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}
	input.UserID = userId.(string)

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := database.DB.Where("firebase_uid = ?", input.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Update only provided fields
	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}
	if input.Phone != nil {
		user.Phone = *input.Phone
	}
	if input.Country != nil {
		user.Country = *input.Country
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}

	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User profile updated successfully",
		"user": gin.H{
			"id":          user.ID,
			"email":       user.Email,
			"displayName": user.DisplayName,
			"phone":       user.Phone,
			"country":     user.Country,
			"bio":         user.Bio,
			"isAdmin":     user.IsAdmin,
		},
	})
}

// ListUsers returns a list of all users
func (ac *AdminController) ListUsers(c *gin.Context) {
	var users []models.User
	if err := database.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
	})
}
