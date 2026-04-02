package seed


import (
	"log"

	"github.com/nduka-junior/movie-reservation/internal/models"
	"github.com/nduka-junior/movie-reservation/internal/utils"
	"gorm.io/gorm"
)

// SeedAdmin creates the initial admin user if it doesn't exist
func SeedAdmin(db *gorm.DB) {
	var count int64

	// Check if admin already exists
	db.Model(&models.User{}).Where("email = ?", "juniorduke@gmail.com").Count(&count)

	if count > 0 {
		log.Println("✅ Admin user already exists")
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword("admin123")
	if err != nil {
		log.Fatal("Failed to hash admin password:", err)
	}

	admin := models.User{
		Email:        "juniorduke@gmail.com",
		PasswordHash: hashedPassword,
		FullName:     "System Administrator",
		IsAdmin:      true,
	}

	if err := db.Create(&admin).Error; err != nil {
		log.Fatal("Failed to create admin user:", err)
	}

	log.Println("🚀 Initial Admin User Created Successfully!")
	log.Println("Email    : admin@example.com")
	log.Println("Password : admin123")
	log.Println("Please change the password after first login!")
}