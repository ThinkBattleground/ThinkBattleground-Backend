package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

type DifficultyLevel string

const (
	Easy   DifficultyLevel = "easy"
	Medium DifficultyLevel = "medium"
	Hard   DifficultyLevel = "hard"
)

type Question struct {
	ID           uint            `gorm:"primaryKey"`
	QuestionID   string          `gorm:"uniqueIndex;not null"` // Unique identifier from Gemini
	Title        string          `gorm:"size:255;not null"`
	Question     string          `gorm:"type:text;not null"`
	Answer       string          `gorm:"type:text;not null"`
	Explanation  string          `gorm:"type:text;not null"`
	Hints        pq.StringArray  `gorm:"type:text[]"`
	Difficulty   DifficultyLevel `gorm:"size:20;not null"`
	ExpectedTime int             `gorm:"default:10"` // in minutes
	Points       int             `gorm:"default:10"`
	Category     string          `gorm:"size:100;index;not null"`
	SubCategory  string          `gorm:"size:100"`
	Tags         pq.StringArray  `gorm:"type:text[]"`
	Requirements pq.StringArray  `gorm:"type:text[]"`
	ImageUrl     string          `gorm:"type:text"`
	CreatedBy    uint            `gorm:"not null"` // Reference to User ID
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Question model
func (Question) TableName() string {
	return "questions"
}

// CreateQuestionRequest represents the request body for creating a question
type CreateQuestionRequest struct {
	Category   string `json:"category" binding:"required"`
	Difficulty string `json:"difficulty" binding:"required,oneof=beginner intermediate advanced expert"`
}

// CreateManualQuestionRequest represents the request body for manually creating a question
type CreateManualQuestionRequest struct {
	Title        string   `json:"title" binding:"required"`
	Question     string   `json:"question" binding:"required"`
	Answer       string   `json:"answer" binding:"required"`
	Explanation  string   `json:"explanation" binding:"required"`
	Hints        []string `json:"hints"`
	Difficulty   string   `json:"difficulty" binding:"required,oneof=beginner intermediate advanced expert"`
	ExpectedTime int      `json:"expectedTime"`
	Points       int      `json:"points"`
	Category     string   `json:"category" binding:"required"`
	SubCategory  string   `json:"subcategory"`
	Tags         []string `json:"tags"`
	Requirements []string `json:"requirements"`
	ImageUrl     string   `json:"imageUrl"`
}
