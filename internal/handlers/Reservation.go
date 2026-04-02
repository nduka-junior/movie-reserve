package handlers

import (

	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nduka-junior/movie-reservation/internal/database"
	"github.com/nduka-junior/movie-reservation/internal/models"
)

type ReservationHandler struct {
	db *database.Database
}

func NewReservationHandler(db *database.Database) *ReservationHandler {
	return &ReservationHandler{db: db}
}

// =============================================
// 1. Get My Reservations (User)
// =============================================
func (h *ReservationHandler) GetMyReservations(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var reservations []models.Reservation

	err := h.db.DB.Preload("Showtime.Movie").
		Preload("Seats").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&reservations).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reservations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reservations": reservations,
	})
}

// =============================================
// 2. Create Reservation (Reserve Seats) - Main Feature
// =============================================
func (h *ReservationHandler) CreateReservation(c *gin.Context) {
userIDInterface, exists := c.Get("user_id")
	
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	// Safe conversion from float64 (from JWT) to uint
	var userID uint
	switch v := userIDInterface.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	case int:
		userID = uint(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}

	// ==================== CHANGE 1 ====================
	// Changed input from SeatIDs to human-friendly seats (Row + Number)
	var input struct {
		ShowtimeID uint `json:"showtime_id" binding:"required"`
		Seats []struct {
			RowLetter  string `json:"row_letter" binding:"required"`
			SeatNumber uint   `json:"seat_number" binding:"required"`
		} `json:"seats" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.db.DB.Transaction(func(tx *gorm.DB) error {

		// 1. Lock showtime to prevent race conditions
		var showtime models.Showtime
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&showtime, input.ShowtimeID).Error; err != nil {
			return err
		}

		// 2. Validate showtime
		if showtime.Status != "active" || showtime.StartTime.Before(time.Now()) {
			return gorm.ErrRecordNotFound
		}

		if uint(len(input.Seats)) > showtime.AvailableSeats {
			return gorm.ErrRecordNotFound
		}

		// ==================== CHANGE 2 ====================
		// Find or create seats based on RowLetter + SeatNumber
		var seatIDs []uint
		for _, s := range input.Seats {
			var seat models.Seat
			err := tx.Where("row_letter = ? AND seat_number = ?", s.RowLetter, s.SeatNumber).
				FirstOrCreate(&seat, models.Seat{
					RowLetter:  s.RowLetter,
					SeatNumber: s.SeatNumber,
				}).Error
			if err != nil {
				return err
			}
			seatIDs = append(seatIDs, seat.ID)
		}

		// 3. Check if any seat is already reserved
		var reservedCount int64
		tx.Model(&models.ReservedSeat{}).
			Where("seat_id IN ?", seatIDs).
			Count(&reservedCount)

		if reservedCount > 0 {
			return gorm.ErrDuplicatedKey
		}

		// 4. Create Reservation
		reservation := models.Reservation{
			UserID:     userID,
			ShowtimeID: input.ShowtimeID,
			Status:     "confirmed",
			TotalPrice: 0,
		}

		if err := tx.Create(&reservation).Error; err != nil {
			return err
		}

		// 5. Link seats to reservation
		var reservedSeats []models.ReservedSeat
		var totalPrice float64 = 0

		for _, seatID := range seatIDs {
			reservedSeats = append(reservedSeats, models.ReservedSeat{
				ReservationID: reservation.ID,
				SeatID:        seatID,
			})
			totalPrice += showtime.BasePrice
		}

		if err := tx.Create(&reservedSeats).Error; err != nil {
			return err
		}

		// 6. Update total price
		reservation.TotalPrice = totalPrice
		if err := tx.Save(&reservation).Error; err != nil {
			return err
		}

		// 7. Decrease available seats
		if err := tx.Model(&showtime).
			Update("available_seats", gorm.Expr("available_seats - ?", len(input.Seats))).
			Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		switch err {
case gorm.ErrRecordNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough seats or showtime not available"})
		case gorm.ErrDuplicatedKey:
			c.JSON(http.StatusBadRequest, gin.H{"error": "One or more seats are already taken"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reservation"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Reservation created successfully",
	})
}

// =============================================
// 3. Cancel Reservation (User - Only Upcoming)
// =============================================
func (h *ReservationHandler) CancelReservation(c *gin.Context) {
userIDInterface, exists := c.Get("user_id")
	
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	// Safe conversion from float64 (from JWT) to uint
	var userID uint
	switch v := userIDInterface.(type) {
	case float64:
		userID = uint(v)
	case uint:
		userID = v
	case int:
		userID = uint(v)
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID type"})
		return
	}
	reservationIDStr := c.Param("id")
	reservationID, err := strconv.ParseUint(reservationIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reservation ID"})
		return
	}

	err = h.db.DB.Transaction(func(tx *gorm.DB) error {
		var reservation models.Reservation
		if err := tx.Preload("Showtime").First(&reservation, reservationID).Error; err != nil {
			return err
		}

		if reservation.UserID != userID{
			return gorm.ErrRecordNotFound
		}

		if reservation.Showtime.StartTime.Before(time.Now()) {
			return gorm.ErrRecordNotFound
		}

		if reservation.Status == "cancelled" {
			return gorm.ErrRecordNotFound
		}

		now := time.Now()
		if err := tx.Model(&reservation).Updates(map[string]interface{}{
			"status":       "cancelled",
			"cancelled_at": now,
		}).Error; err != nil {
			return err
		}

		// Restore available seats
		seatCount := len(reservation.Seats)
		if err := tx.Model(&models.Showtime{}).
			Where("id = ?", reservation.ShowtimeID).
			Update("available_seats", gorm.Expr("available_seats + ?", seatCount)).
			Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot cancel this reservation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reservation cancelled successfully"})
}