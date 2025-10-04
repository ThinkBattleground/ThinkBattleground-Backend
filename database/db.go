package database

import (
	"fmt"
	"log"

	"github.com/ThinkBattleground/ThinkBattleground-Backend/config"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	dbHost := config.GetEnv("POSTGRES_HOST", "localhost")
	dbUser := config.GetEnv("POSTGRES_USER", "postgres")
	dbPassword := config.GetEnv("POSTGRES_PASSWORD", "")
	dbName := config.GetEnv("POSTGRES_DB", "thinkbattleground")
	dbPort := config.GetEnv("POSTGRES_PORT", "5432")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbHost, dbUser, dbPassword, dbName, dbPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto-migrate the models in the correct order
	// First, migrate independent models (User and PuzzleTag)
	if err = db.AutoMigrate(&models.User{}, &models.PuzzleTag{}); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// Then migrate Puzzle which depends on User and has a many-to-many with PuzzleTag
	if err = db.AutoMigrate(&models.Puzzle{}); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	// Finally migrate Solution which depends on both User and Puzzle
	if err = db.AutoMigrate(&models.Solution{}); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	DB = db
	log.Println("Database migration completed successfully")

	DB = db
	log.Println("Database connected successfully")
	return nil
}
