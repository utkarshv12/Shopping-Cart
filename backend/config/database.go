package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"shopping-cart/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ConnectDatabase() {
	// Allow full DSN override via MYSQL_DSN env var
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		user := getenv("DB_USER", "root")
		pass := os.Getenv("DB_PASS") // do NOT provide a default here - require it for safety
		host := getenv("DB_HOST", "127.0.0.1")
		port := getenv("DB_PORT", "3306")
		name := getenv("DB_NAME", "shopping_cart")

		if strings.TrimSpace(pass) == "" {
			log.Fatal("No database credentials found. Set MYSQL_DSN or set DB_PASS in environment (see .env.example). Aborting for safety.")
		}

		dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, name)
		// Mask password when logging
		masked := strings.Replace(dsn, ":"+pass+"@", ":****@", 1)
		log.Println("Using DB DSN:", masked)
	} else {
		// For full DSN provided, do not print the password; mask any @... preceding DB name if possible
		masked := dsn
		// attempt to hide credentials between start and @
		if at := strings.Index(dsn, "@"); at != -1 && strings.Contains(dsn, ":") {
			// keep scheme after first '@' and mask credentials portion
			masked = "****@" + dsn[at+1:]
		}
		log.Println("Using DB DSN (provided via MYSQL_DSN):", masked)
	}

	var err error

	// Connect to MySQL
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to MySQL database:", err)
	}

	// Run AutoMigrate on all models
	err = DB.AutoMigrate(
		&models.User{},
		&models.Item{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
		&models.OrderItem{},
		&models.Session{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("MySQL database connected and migrated successfully")
}
