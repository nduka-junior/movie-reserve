package models

import (
	"time"

	"gorm.io/gorm"
)


type Genre struct {
	gorm.Model
	Name   string `gorm:"unique;not null"`
	Movies []Movie `gorm:"many2many:movie_genres;"`
}

type Movie struct {
	gorm.Model
	Title            string
	Description      string
	PosterURL        string
	ReleaseDate      *time.Time
	Status           string `gorm:"default:'coming-soon'"` // coming-soon, now-showing, ended
	Genres           []Genre `gorm:"many2many:movie_genres;"`
	DurationMinutes  uint


}

