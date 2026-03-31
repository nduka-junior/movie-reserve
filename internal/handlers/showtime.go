package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nduka-junior/movie-reservation/internal/database"
	"github.com/nduka-junior/movie-reservation/internal/models"
)

type ShowtimeHandler struct {
	db *database.Database
}

func NewShowtimeHandler(db *database.Database) *ShowtimeHandler {
	return &ShowtimeHandler{db: db}
}

// AddShowtime – ADMIN ONLY
// POST /api/v1/movies/:movieId/showtimes
func (h *ShowtimeHandler) AddShowtime(c *gin.Context) {
	// Get movieId from URL
	movieIDStr := c.Param("movieId")
	movieID, err := strconv.ParseUint(movieIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	var input struct {
		StartTime string  `json:"start_time" binding:"required"` // e.g. "2025-06-20T18:30:00Z"
		BasePrice float64 `json:"base_price" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify movie exists
	var movie models.Movie
	if err := h.db.DB.First(&movie, movieID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	// Parse start time
	startTime, err := time.Parse(time.RFC3339, input.StartTime)
	if err != nil {
		// Try alternative format
		startTime, err = time.Parse("2006-01-02 15:04:05", input.StartTime)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_time format. Use ISO format (e.g. 2025-06-20T18:30:00Z)"})
			return
		}
	}

	// Calculate EndTime based on movie duration
	endTime := startTime.Add(time.Duration(movie.DurationMinutes) * time.Minute)

	showtime := models.Showtime{
		MovieID:        uint(movieID),
		StartTime:      startTime,
		EndTime:        endTime,
		BasePrice:      input.BasePrice,
		AvailableSeats: 0, // Will be updated after seats are created
		Status:         "active",
	}

	// Create showtime inside transaction
	err = h.db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&showtime).Error; err != nil {
			return err
		}

		// Optional: Set initial AvailableSeats (you can improve this later)
		// For now, we'll set a default value. You can change this logic.
		return tx.Model(&showtime).Update("available_seats", 100).Error // Default 100 seats
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create showtime"})
		return
	}

	// Load movie info for response
	h.db.DB.Preload("Movie").First(&showtime, showtime.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Showtime created successfully",
		"showtime": showtime,
	})
}