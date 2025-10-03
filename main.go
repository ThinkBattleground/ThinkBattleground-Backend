package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/config"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/routes"
)

func main() {
	// Load environment variables
	if err := config.LoadEnv(); err != nil {
		log.Fatal("Failed to load .env file")
	}

	// Initialize Gin router
	router := gin.Default()

	// Initialize Firebase App
	firebaseApp := config.InitializeFirebase()
	if firebaseApp == nil {
		log.Fatal("Error initializing Firebase")
	}

	// Initialize routes
	routes.InitializeRoutes(router, firebaseApp)

	// Start server
	if err := router.Run(":8080"); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
