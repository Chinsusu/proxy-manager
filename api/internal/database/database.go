package database

import (
	"fmt"
	"log"
	"time"

	"github.com/Chinsusu/proxy-manager/api/internal/config"
	"github.com/Chinsusu/proxy-manager/api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DB struct {
	*gorm.DB
}

// Connect establishes database connection and runs migrations
func Connect(cfg *config.Config) (*DB, error) {
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}
	
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	// Run migrations
	if err := models.AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	dbWrapper := &DB{db}
	
	// Seed admin user
	if err := dbWrapper.SeedAdminUser(cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Printf("Warning: failed to seed admin user: %v", err)
	}

	return dbWrapper, nil
}

// SeedAdminUser creates admin user if not exists
func (db *DB) SeedAdminUser(email, password string) error {
	var existingUser models.User
	result := db.Where("email = ?", email).First(&existingUser)
	
	if result.Error == nil {
		// User already exists
		return nil
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	adminUser := models.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "admin",
	}

	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Printf("Admin user created with email: %s", email)
	return nil
}

// IncrementConfigVersion increments config version for a server
func (db *DB) IncrementConfigVersion(serverID uint) error {
	return db.Model(&models.Server{}).
		Where("id = ?", serverID).
		UpdateColumn("config_version", gorm.Expr("config_version + 1")).Error
}

// GetCurrentConfigVersion returns current config version for a server
func (db *DB) GetCurrentConfigVersion(serverID uint) (int, error) {
	var server models.Server
	err := db.Select("config_version").Where("id = ?", serverID).First(&server).Error
	if err != nil {
		return 0, err
	}
	return server.ConfigVersion, nil
}
