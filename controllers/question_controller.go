package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ThinkBattleground/ThinkBattleground-Backend/config"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/database"
	"github.com/ThinkBattleground/ThinkBattleground-Backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type QuestionController struct{}

// CreateQuestionWithGemini generates and creates a new math question using Gemini API
func (qc *QuestionController) CreateQuestionWithGemini(c *gin.Context) {
	var input models.CreateQuestionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate difficulty level
	validDifficulties := map[string]bool{
		"beginner":     true,
		"intermediate": true,
		"advanced":     true,
		"expert":       true,
	}
	if !validDifficulties[input.Difficulty] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid difficulty level. Must be: beginner, intermediate, advanced, or expert"})
		return
	}

	// Get the admin user who is creating the question
	adminUser, exists := c.Get("dbUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	admin := adminUser.(models.User)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Generate question using Gemini API
	generatedQuestion, err := config.GenerateQuestion(ctx, input.Category, input.Difficulty)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to generate question: %v", err)})
		return
	}

	// Parse the generated question into our model
	question, err := parseGeneratedQuestion(generatedQuestion, admin.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to parse generated question: %v", err)})
		return
	}

	// Save to database
	if err := database.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question to database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Question created successfully",
		"question": question,
	})
}

// ListQuestions retrieves all questions
func (qc *QuestionController) ListQuestions(c *gin.Context) {
	var questions []models.Question

	if err := database.DB.Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":     len(questions),
		"questions": questions,
	})
}

// GetQuestion retrieves a single question by ID
func (qc *QuestionController) GetQuestion(c *gin.Context) {
	id := c.Param("id")

	var question models.Question
	if err := database.DB.Where("question_id = ?", id).First(&question).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Question not found"})
		return
	}

	c.JSON(http.StatusOK, question)
}

// GetQuestionsByCategory retrieves questions filtered by category
func (qc *QuestionController) GetQuestionsByCategory(c *gin.Context) {
	category := c.Param("category")

	var questions []models.Question
	if err := database.DB.Where("category = ?", category).Find(&questions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve questions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"category":  category,
		"count":     len(questions),
		"questions": questions,
	})
}

// parseGeneratedQuestion converts Gemini's response into a Question model
func parseGeneratedQuestion(data map[string]interface{}, userID uint) (models.Question, error) {
	question := models.Question{}

	// Extract and set fields from the generated data
	if id, ok := data["id"].(string); ok && id != "" {
		question.QuestionID = id
	} else {
		question.QuestionID = uuid.New().String()
	}

	if title, ok := data["title"].(string); ok {
		question.Title = title
	}

	if questionText, ok := data["question"].(string); ok {
		question.Question = questionText
	}

	// Extract solution
	if solutionData, ok := data["solution"].(map[string]interface{}); ok {
		if answer, ok := solutionData["answer"].(string); ok {
			question.Answer = answer
		}
		if explanation, ok := solutionData["explanation"].(string); ok {
			question.Explanation = explanation
		}
	}

	// Extract hints
	if hintsData, ok := data["hints"].([]interface{}); ok {
		hints := make([]string, len(hintsData))
		for i, h := range hintsData {
			if hint, ok := h.(string); ok {
				hints[i] = hint
			}
		}
		question.Hints = pq.StringArray(hints)
	}

	if difficulty, ok := data["difficulty"].(string); ok {
		question.Difficulty = models.DifficultyLevel(difficulty)
	}

	if expectedTime, ok := data["expectedTime"].(float64); ok {
		question.ExpectedTime = int(expectedTime)
	}

	if points, ok := data["points"].(float64); ok {
		question.Points = int(points)
	}

	if category, ok := data["category"].(string); ok {
		question.Category = category
	}

	if subcategory, ok := data["subcategory"].(string); ok {
		question.SubCategory = subcategory
	}

	// Extract tags
	if tagsData, ok := data["tags"].([]interface{}); ok {
		tags := make([]string, len(tagsData))
		for i, t := range tagsData {
			if tag, ok := t.(string); ok {
				tags[i] = tag
			}
		}
		question.Tags = pq.StringArray(tags)
	}

	// Extract requirements
	if reqsData, ok := data["requirements"].([]interface{}); ok {
		reqs := make([]string, len(reqsData))
		for i, r := range reqsData {
			if req, ok := r.(string); ok {
				reqs[i] = req
			}
		}
		question.Requirements = pq.StringArray(reqs)
	}

	if imageUrl, ok := data["imageUrl"].(string); ok {
		question.ImageUrl = imageUrl
	}

	question.CreatedBy = userID

	// Validate required fields
	if question.Title == "" || question.Question == "" || question.Category == "" {
		return question, fmt.Errorf("generated question missing required fields")
	}

	return question, nil
}

// CreateManualQuestion creates a question manually without AI
func (qc *QuestionController) CreateManualQuestion(c *gin.Context) {
	var input models.CreateManualQuestionRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate difficulty level
	validDifficulties := map[string]bool{
		"beginner":     true,
		"intermediate": true,
		"advanced":     true,
		"expert":       true,
	}
	if !validDifficulties[input.Difficulty] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid difficulty level. Must be: beginner, intermediate, advanced, or expert"})
		return
	}

	// Get the admin user who is creating the question
	adminUser, exists := c.Get("dbUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	admin := adminUser.(models.User)

	// Set default values if not provided
	expectedTime := input.ExpectedTime
	if expectedTime == 0 {
		expectedTime = 10
	}

	points := input.Points
	if points == 0 {
		points = 50
	}

	// Create the question
	question := models.Question{
		QuestionID:   uuid.New().String(),
		Title:        input.Title,
		Question:     input.Question,
		Answer:       input.Answer,
		Explanation:  input.Explanation,
		Hints:        pq.StringArray(input.Hints),
		Difficulty:   models.DifficultyLevel(input.Difficulty),
		ExpectedTime: expectedTime,
		Points:       points,
		Category:     input.Category,
		SubCategory:  input.SubCategory,
		Tags:         pq.StringArray(input.Tags),
		Requirements: pq.StringArray(input.Requirements),
		ImageUrl:     input.ImageUrl,
		CreatedBy:    admin.ID,
	}

	// Save to database
	if err := database.DB.Create(&question).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save question to database"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Question created successfully",
		"question": question,
	})
}
