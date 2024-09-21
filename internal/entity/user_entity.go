package entity

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username string `gorm:"unique,not null,column:username"`
	Password string `gorm:"not null"`
	Salt     string `gorm:"not null"`
}
