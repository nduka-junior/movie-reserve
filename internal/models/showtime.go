package models

import (
	"time"

	"gorm.io/gorm"
)


type Showtime struct {
	gorm.Model
	MovieID        uint
	StartTime      time.Time `gorm:"index"`
	EndTime        time.Time // can be computed: StartTime + Movie.Duration

	BasePrice      float64
	AvailableSeats uint   `gorm:"default:0"` // denormalized count – update transactionally
	Status         string `gorm:"default:'active'"`
	Movie Movie `gorm:"foreignKey:MovieID"`
	Reservations []Reservation
}
