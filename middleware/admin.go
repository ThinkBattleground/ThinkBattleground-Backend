package middleware

import (
	"net/http"

	"github.com/ThinkBattleground/ThinkBattleground-Backend/database"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/models"
	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the user ID from the context (set by AuthMiddleware)
		firebaseUID, exists := c.Get("userId")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		// Find the user in the database
		var user models.User
		if err := database.DB.Where("firebase_uid = ?", firebaseUID).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in database"})
			c.Abort()
			return
		}

		// Check if the user is an admin
		if !user.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "User is not an admin"})
			c.Abort()
			return
		}

		// Add user to context for further use
		c.Set("dbUser", user)
		c.Next()
	}
}
