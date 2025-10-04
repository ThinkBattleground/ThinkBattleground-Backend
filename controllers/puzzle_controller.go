package controllers

import (
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
		Title       string                 `json:"title" binding:"required"`
		Description string                 `json:"description" binding:"required"`
		Difficulty  models.DifficultyLevel `json:"difficulty" binding:"required"`
		Type        models.QuestionType    `json:"type" binding:"required"`
		Tags        []string               `json:"tags" binding:"required"`
		Solution    struct {
			Content  string `json:"content" binding:"required"`
			Language string `json:"language" binding:"required"`
		} `json:"solution" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get the admin user who is creating the puzzle
	adminUser, exists := c.Get("dbUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	admin := adminUser.(models.User)

	// Start a database transaction
	tx := database.DB.Begin()

	// Create or get tags
	var tags []models.PuzzleTag
	for _, tagName := range input.Tags {
		var tag models.PuzzleTag
		if err := tx.FirstOrCreate(&tag, models.PuzzleTag{Name: tagName}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tags"})
			return
		}
		tags = append(tags, tag)
	}

	// Create the puzzle
	puzzle := models.Puzzle{
		Title:       input.Title,
		Description: input.Description,
		Difficulty:  input.Difficulty,
		Type:        input.Type,
		Tags:        tags,
		CreatedBy:   admin.ID,
	}

	if err := tx.Create(&puzzle).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create puzzle"})
		return
	}

	// Create the initial solution
	solution := models.Solution{
		PuzzleID:  puzzle.ID,
		Content:   input.Solution.Content,
		Language:  input.Solution.Language,
		CreatedBy: admin.ID,
	}

	if err := tx.Create(&solution).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create solution"})
		return
	}

	// Commit the transaction
	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"message": "Puzzle created successfully",
		"puzzle": gin.H{
			"id":          puzzle.ID,
			"title":       puzzle.Title,
			"description": puzzle.Description,
			"difficulty":  puzzle.Difficulty,
			"type":        puzzle.Type,
			"tags":        tags,
			"solution": gin.H{
				"id":       solution.ID,
				"content":  solution.Content,
				"language": solution.Language,
			},
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
	if err := database.DB.Preload("Tags").Preload("Solutions").First(&puzzle, puzzleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Puzzle not found"})
		return
	}

	// Get creator information
	var creator models.User
	if err := database.DB.Select("id, email, display_name").First(&creator, puzzle.CreatedBy).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get creator info"})
		return
	}

	// Format solutions
	var formattedSolutions []gin.H
	for _, solution := range puzzle.Solutions {
		var solutionCreator models.User
		if err := database.DB.Select("id, email, display_name").First(&solutionCreator, solution.CreatedBy).Error; err != nil {
			continue
		}

		formattedSolutions = append(formattedSolutions, gin.H{
			"id":       solution.ID,
			"content":  solution.Content,
			"language": solution.Language,
			"creator": gin.H{
				"id":          solutionCreator.ID,
				"email":       solutionCreator.Email,
				"displayName": solutionCreator.DisplayName,
			},
			"createdAt": solution.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"puzzle": gin.H{
			"id":          puzzle.ID,
			"title":       puzzle.Title,
			"description": puzzle.Description,
			"difficulty":  puzzle.Difficulty,
			"type":        puzzle.Type,
			"tags":        puzzle.Tags,
			"creator": gin.H{
				"id":          creator.ID,
				"email":       creator.Email,
				"displayName": creator.DisplayName,
			},
			"solutions": formattedSolutions,
			"createdAt": puzzle.CreatedAt,
			"updatedAt": puzzle.UpdatedAt,
		},
	})
}

// ListPuzzles returns a list of all puzzles with optional filters
func (pc *PuzzleController) ListPuzzles(c *gin.Context) {
	var puzzles []models.Puzzle
	query := database.DB.Preload("Tags")

	// Apply filters if provided
	if difficulty := c.Query("difficulty"); difficulty != "" {
		query = query.Where("difficulty = ?", difficulty)
	}
	if puzzleType := c.Query("type"); puzzleType != "" {
		query = query.Where("type = ?", puzzleType)
	}
	if tag := c.Query("tag"); tag != "" {
		query = query.Joins("JOIN puzzle_tags ON puzzles.id = puzzle_tags.puzzle_id").
			Joins("JOIN puzzle_tags tags ON puzzle_tags.tag_id = tags.id").
			Where("tags.name = ?", tag)
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
			"title":      puzzle.Title,
			"difficulty": puzzle.Difficulty,
			"type":       puzzle.Type,
			"tags":       puzzle.Tags,
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
