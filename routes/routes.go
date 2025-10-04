package routes

import (
	"context"
	"log"

	"github.com/ThinkBattleground/ThinkBattleground-Backend/controllers"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/middleware"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
)

func InitializeRoutes(router *gin.Engine, app *firebase.App) {
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
		return
	}

	// Initialize controllers
	authController := controllers.NewAuthController(authClient)

	// Public routes
	public := router.Group("/api/v1")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "OK",
			})
		})

		// Auth routes
		public.GET("/auth/methods", authController.GetAuthMethods)
		public.POST("/auth/signup", authController.SignUp)
		public.POST("/auth/signin", authController.SignIn)
		public.POST("/auth/forgot-password", authController.ForgotPassword)
		public.GET("/auth/google", authController.InitiateGoogleSignIn)
		public.GET("/auth/google/callback", authController.HandleGoogleCallback)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(middleware.AuthMiddleware(authClient))
	{
		protected.GET("/verify", authController.VerifyToken)
		protected.GET("/profile", authController.GetUserProfile)
		protected.PUT("/users/profile", authController.UpdateUserProfile)
	}

	// Admin routes
	adminController := &controllers.AdminController{}
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware(authClient))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.POST("/users/make-admin", adminController.MakeUserAdmin)
		admin.GET("/users", adminController.ListUsers)
	}
}
