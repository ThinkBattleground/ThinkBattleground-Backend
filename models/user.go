package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          uint   `gorm:"primaryKey"`
	FirebaseUID string `gorm:"unique;not null"`
	Email       string `gorm:"unique;not null"`
	IsAdmin     bool   `gorm:"default:false"`
	DisplayName string `gorm:"size:255"`
	PhotoURL    string `gorm:"size:512"`
	Phone       string `gorm:"size:20"`
	Country     string `gorm:"size:100"`
	Bio         string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

// TableName specifies the table name for User model
func (User) TableName() string {
	return "users"
}
