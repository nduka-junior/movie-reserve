package models

import (
	"errors"
	"regexp"


	"gorm.io/gorm"
)

// User represents our database user
type User struct {
	gorm.Model
	Email        string `gorm:"unique;not null;index"`
	PasswordHash string `gorm:"not null"`
	FullName     string
	IsAdmin      bool `gorm:"default:false"`
}
// type User struct {
// 		gorm.Model
// 	    ID           int       `json:"id"`
// 	Email   	 string `gorm:"unique;not null;index"`
// 	PasswordHash string `gorm:"not null"`
// 	FullName     string
// 	Roles        []Role `gorm:"many2many:user_roles;"`
// 	Reservations []Reservation

// }

// Role is a simple enum-like table (admin, user)
// In practice many projects use just a string field on User, but separate table allows future expansion
// type Role struct {
// 	gorm.Model
// 	Name  string `gorm:"unique;not null"` // "admin", "user"
// 	Users []User `gorm:"many2many:user_roles;"`
// }

// UserRole join table (if using many-to-many)
// type UserRole struct {
// 	UserID uint `gorm:"primaryKey"`
// 	RoleID uint `gorm:"primaryKey"`
// }

// UserLogin represents login request data
type UserLogin struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// UserRegister represents registration request data
type UserRegister struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

// Validate checks if email format is valid
func (u *UserRegister) Validate() error {
    emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
    if !emailRegex.MatchString(u.Email) {
        return errors.New("invalid email format")
    }
    return nil
}