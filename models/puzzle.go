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

type QuestionType string

const (
	Algorithm     QuestionType = "algorithm"
	DataStructure QuestionType = "data_structure"
	SystemDesign  QuestionType = "system_design"
	Database      QuestionType = "database"
)

type Puzzle struct {
	ID          uint            `gorm:"primaryKey"`
	Title       string          `gorm:"size:255;not null"`
	Description string          `gorm:"type:text;not null"`
	Difficulty  DifficultyLevel `gorm:"size:20;not null"`
	Type        QuestionType    `gorm:"size:50;not null"`
	Tags        []PuzzleTag     `gorm:"many2many:puzzle_tag_map;"`
	Solutions   []Solution      `gorm:"foreignKey:PuzzleID"`
	CreatedBy   uint            `gorm:"not null"` // Reference to User ID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type PuzzleTag struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"size:50;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Solution struct {
	ID        uint   `gorm:"primaryKey"`
	PuzzleID  uint   `gorm:"not null"`
	Content   string `gorm:"type:text;not null"`
	Language  string `gorm:"size:50;not null"` // e.g., python, javascript, java
	CreatedBy uint   `gorm:"not null"`         // Reference to User ID
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for Puzzle model
func (Puzzle) TableName() string {
	return "puzzles"
}

// TableName specifies the table name for PuzzleTag model
func (PuzzleTag) TableName() string {
	return "puzzle_tags"
}

// TableName specifies the table name for Solution model
func (Solution) TableName() string {
	return "solutions"
}
