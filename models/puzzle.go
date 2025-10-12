package models

import (
	"time"

	"gorm.io/gorm"
)

type DifficultyLevel string

const (
	Easy   DifficultyLevel = "easy"
	Medium DifficultyLevel = "medium"
	Hard   DifficultyLevel = "hard"
)

type Puzzle struct {
	ID         uint            `gorm:"primaryKey"`
	Topic      string          `gorm:"size:50;index;not null"`
	Difficulty DifficultyLevel `gorm:"size:20;not null"`
	Data       interface{}     `gorm:"type:jsonb"` // Store any JSON data
	CreatedBy  uint            `gorm:"not null"`   // Reference to User ID
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Puzzle model
func (Puzzle) TableName() string {
	return "puzzles"
}
