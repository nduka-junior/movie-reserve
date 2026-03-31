package models

import (
	"time"

	"gorm.io/gorm"
)


type Reservation struct {
	gorm.Model
	UserID      uint

	ShowtimeID  uint
	TotalPrice  float64
	Status      string    `gorm:"default:'pending'"` // pending, confirmed, cancelled, attended
	ReservedAt  time.Time `gorm:"autoCreateTime"`
	CancelledAt *time.Time

	Showtime Showtime `gorm:"foreignKey:ShowtimeID"`
	Seats []Seat `gorm:"many2many:reserved_seats;"`
}

type ReservedSeat struct {
	ReservationID uint `gorm:"primaryKey"`
	SeatID        uint `gorm:"primaryKey"`
}
type Seat struct {
	gorm.Model

	RowLetter  string // "A", "B", "C", etc.
	
	SeatNumber  uint   // 1, 2, 3...


}
