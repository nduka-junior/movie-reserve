package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nduka-junior/movie-reservation/internal/config"
	"github.com/nduka-junior/movie-reservation/internal/database"
	"github.com/nduka-junior/movie-reservation/internal/handlers"
	"github.com/nduka-junior/movie-reservation/internal/middleware"
	"github.com/nduka-junior/movie-reservation/internal/models"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }

    // Initialize database
    db, err := database.NewDatabase(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
defer func() {
    sqlDB, _ := db.DB.DB()
    sqlDB.Close()
}()


	if cfg.Environment == "development" {
    // Only migrate in local dev (fast on local Postgres)
    db.DB.AutoMigrate(&models.User{},
	&models.Movie{},
	&models.Genre{},
	&models.Showtime{},
	&models.Seat{},
	&models.Reservation{},
	&models.ReservedSeat{}, )
    log.Println("Auto-migrated schema (development only)")
} else {
    log.Println("Skipping AutoMigrate (production / Neon)")
}
    // Set Gin mode
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }


    // Initialize router with middleware
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(gin.Logger())

    // CORS middleware
    r.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })

    // Initialize handlers with JWT configuration
    authHandler := handlers.NewAuthHandler(db, []byte(cfg.JWT.Secret))
	movieHandler := handlers.NewMovieHandler(db)
	showtimeHandler := handlers.NewShowtimeHandler(db)	

    // Public routes
    public := r.Group("/api/v1")
    {
        public.POST("/register", authHandler.Register)
        public.POST("/login", authHandler.Login)
    }

    // Protected routes with JWT middleware
    protected := r.Group("/api/v1")
    protected.Use(middleware.AuthMiddleware([]byte(cfg.JWT.Secret)))
    {
		// Admin only routes
	admin := protected.Group("")
	admin.Use(middleware.RequireAdmin())
	{
		admin.POST("/movies", movieHandler.AddMovie)    
		admin.PUT("/movies/:id", movieHandler.UpdateMovie)
		admin.DELETE("/movies/:id", movieHandler.DeleteMovie)
		admin.POST("/movies/:movieId/showtimes", showtimeHandler.AddShowtime)

		    // ← Only Admin
		// You can add more admin routes here later:
		// admin.POST("/showtimes", ...)
		// admin.PUT("/movies/:id", ...)
	}
        protected.POST("/refresh-token", authHandler.RefreshToken)
        protected.POST("/logout", authHandler.Logout)
        protected.GET("/profile", getUserProfile)
    }

    // Start server with configured host and port
    serverAddr := cfg.Server.Host + ":" + cfg.Server.Port
    log.Printf("Server starting on %s", serverAddr)

    srv := &http.Server{
        Addr:         serverAddr,
        Handler:      r,
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }

    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal("Server failed to start:", err)
    }
}

func getUserProfile(c *gin.Context) {
    userID, _ := c.Get("user_id")
    email, _ := c.Get("email")
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "email":   email,
    })
}