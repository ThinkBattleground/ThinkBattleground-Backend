package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ThinkBattleground/ThinkBattleground-Backend/database"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/models"
	"github.com/gin-gonic/gin"
)

type PuzzleController struct{}

// CreatePuzzle creates a new puzzle (admin only)
func (pc *PuzzleController) CreatePuzzle(c *gin.Context) {
	var input struct {
		Topic      string                 `json:"topic" binding:"required"`
		Difficulty models.DifficultyLevel `json:"difficulty" binding:"required"`
		Data       json.RawMessage        `json:"data" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that Data is valid JSON
	if !json.Valid(input.Data) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON in data field"})
		return
	}

	// Get the admin user who is creating the puzzle
	adminUser, exists := c.Get("dbUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	admin := adminUser.(models.User)

	// Create the puzzle
	puzzle := models.Puzzle{
		Topic:      input.Topic,
		Difficulty: input.Difficulty,
		Data:       input.Data,
		CreatedBy:  admin.ID,
	}

	if err := database.DB.Create(&puzzle).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create puzzle"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Puzzle created successfully",
		"puzzle": gin.H{
			"id":         puzzle.ID,
			"topic":      puzzle.Topic,
			"difficulty": puzzle.Difficulty,
			"data":       puzzle.Data,
			"createdBy":  puzzle.CreatedBy,
			"createdAt":  puzzle.CreatedAt,
		},
	})
}

// GetPuzzle retrieves a puzzle by ID
func (pc *PuzzleController) GetPuzzle(c *gin.Context) {
	puzzleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid puzzle ID"})
		return
	}

	var puzzle models.Puzzle
	if err := database.DB.First(&puzzle, puzzleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Puzzle not found"})
		return
	}

	// Get creator information
	var creator models.User
	if err := database.DB.Select("id, email, display_name").First(&creator, puzzle.CreatedBy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get creator info"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"puzzle": gin.H{
			"id":         puzzle.ID,
			"topic":      puzzle.Topic,
			"difficulty": puzzle.Difficulty,
			"data":       puzzle.Data,
			"creator": gin.H{
				"id":          creator.ID,
				"email":       creator.Email,
				"displayName": creator.DisplayName,
			},
			"createdAt": puzzle.CreatedAt,
			"updatedAt": puzzle.UpdatedAt,
		},
	})
}

// ListPuzzles returns a list of all puzzles with optional filters
func (pc *PuzzleController) ListPuzzles(c *gin.Context) {
	var puzzles []models.Puzzle
	query := database.DB

	// Apply filters if provided
	if topic := c.Query("topic"); topic != "" {
		query = query.Where("topic = ?", topic)
	}
	if difficulty := c.Query("difficulty"); difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}

	// Support JSON field filtering
	if jsonField := c.Query("jsonField"); jsonField != "" {
		jsonValue := c.Query("jsonValue")
		query = query.Where("data->>? = ?", jsonField, jsonValue)
	}

	if err := query.Find(&puzzles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch puzzles"})
		return
	}

	var formattedPuzzles []gin.H
	for _, puzzle := range puzzles {
		var creator models.User
		if err := database.DB.Select("id, email, display_name").First(&creator, puzzle.CreatedBy).Error; err != nil {
			continue
		}

		formattedPuzzles = append(formattedPuzzles, gin.H{
			"id":         puzzle.ID,
			"topic":      puzzle.Topic,
			"difficulty": puzzle.Difficulty,
			"data":       puzzle.Data,
			"creator": gin.H{
				"id":          creator.ID,
				"email":       creator.Email,
				"displayName": creator.DisplayName,
			},
			"createdAt": puzzle.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"puzzles": formattedPuzzles,
	})
}
