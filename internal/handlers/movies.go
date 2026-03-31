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

type MovieHandler struct {
	db *database.Database
}

func NewMovieHandler(db *database.Database) *MovieHandler {
	return &MovieHandler{db: db}
}

// AddMovie – ADMIN ONLY
// POST /api/v1/movies
// AddMovie – ADMIN ONLY
// POST /api/v1/movies
func (h *MovieHandler) AddMovie(c *gin.Context) {

	var input struct {
		Title           string   `json:"title" binding:"required"`
		Description     string   `json:"description"`
		PosterURL       string   `json:"poster_url"`
		DurationMinutes uint     `json:"duration_minutes" binding:"required"` // Added - important for showtimes
		ReleaseDate     string   `json:"release_date"` // "2025-06-15"
		GenreIDs        []uint   `json:"genre_ids"`
		Status          string   `json:"status" binding:"omitempty,oneof=coming-soon now-showing ended"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse Release Date (optional)
	var releaseDate *time.Time
	if input.ReleaseDate != "" {
		parsed, err := time.Parse("2006-01-02", input.ReleaseDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release_date format. Use YYYY-MM-DD"})
			return
		}
		releaseDate = &parsed
	}

	// Default status
	if input.Status == "" {
		input.Status = "coming-soon"
	}

	// Create movie struct matching your model
	movie := models.Movie{
		Title:           input.Title,
		Description:     input.Description,
		PosterURL:       input.PosterURL, // ← Added
		ReleaseDate:     releaseDate,
		Status:          input.Status,
	}

	// Handle genres if provided
	if len(input.GenreIDs) > 0 {
		var genres []models.Genre
		if err := h.db.DB.Where("id IN ?", input.GenreIDs).Find(&genres).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "database error while fetching genres"})
			return
		}

		if len(genres) != len(input.GenreIDs) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "one or more genre IDs are invalid"})
			return
		}

		movie.Genres = genres
	}

	// Save to database
	if err := h.db.DB.Create(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create movie"})
		return
	}

	// Reload movie with Genres preloaded
	h.db.DB.Preload("Genres").First(&movie, movie.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Movie created successfully",
		"movie":   movie,
	})
}

// GetMovies – PUBLIC
// GET /api/v1/movies
// Supports ?status=now-showing&genre=action&limit=20&page=1
func (h *MovieHandler) GetMovies(c *gin.Context) {
	var movies []models.Movie

	query := h.db.DB.Model(&models.Movie{}).Preload("Genres")

	// Filters
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if genre := c.Query("genre"); genre != "" {
		query = query.Joins("JOIN movie_genres ON movie_genres.movie_id = movies.id").
			Joins("JOIN genres ON genres.id = movie_genres.genre_id").
			Where("genres.name = ?", genre)
	}

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	var total int64
	h.db.DB.Model(&models.Movie{}).Count(&total)

	query = query.Limit(limit).Offset(offset).Order("release_date DESC, title ASC")

	if err := query.Find(&movies).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movies"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"movies": movies,
		"total":  total,
		"page":   page,
		"limit":  limit,
	})
}

// GetMovie – PUBLIC
// GET /api/v1/movies/:id
func (h *MovieHandler) GetMovie(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	var movie models.Movie
	if err := h.db.DB.Preload("Genres").First(&movie, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

// UpdateMovie – ADMIN ONLY
// PUT /api/v1/movies/:id
func (h *MovieHandler) UpdateMovie(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	var input struct {
		Title           *string `json:"title"`
		Description     *string `json:"description"`
		PosterURL       *string `json:"poster_url"`
		ReleaseDate     *string `json:"release_date"`
		Status          *string `json:"status"`
		GenreIDs        []uint  `json:"genre_ids,omitempty"`
	}


	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var movie models.Movie
	if err := h.db.DB.First(&movie, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		}
		return
	}

	// Update only provided fields
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Description != nil {
		movie.Description = *input.Description
	}
	if input.PosterURL != nil {
		movie.PosterURL = *input.PosterURL
	}
	if input.Status != nil {
		movie.Status = *input.Status
	}
	if input.ReleaseDate != nil {
		if *input.ReleaseDate == "" {
			movie.ReleaseDate = nil
		} else {
			parsed, err := time.Parse("2006-01-02", *input.ReleaseDate)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid release_date format"})
				return
			}
			movie.ReleaseDate = &parsed
		}
	}

	// Update genres if provided (replace all)
	if input.GenreIDs != nil {
		var genres []models.Genre
		if len(input.GenreIDs) > 0 {
			if err := h.db.DB.Where("id IN ?", input.GenreIDs).Find(&genres).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid genre IDs"})
				return
			}
		}
		movie.Genres = genres
	}

	if err := h.db.DB.Save(&movie).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update movie"})
		return
	}

	h.db.DB.Preload("Genres").First(&movie, movie.ID)

	c.JSON(http.StatusOK, gin.H{"movie": movie})
}

// DeleteMovie – ADMIN ONLY
// DELETE /api/v1/movies/:id
func (h *MovieHandler) DeleteMovie(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie ID"})
		return
	}

	result := h.db.DB.Delete(&models.Movie{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete movie"})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Movie deleted successfully"})
}