package middleware

import (
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(auth *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		idToken := strings.Replace(authHeader, "Bearer ", "", 1)

		if idToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No token provided",
			})
			c.Abort()
			return
		}

		// Verify the ID token
		token, err := auth.VerifyIDToken(c, idToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// Add the user ID to the context
		c.Set("userId", token.UID)
		c.Next()
	}
}
