package database

import (
	"fmt"
	"log"

	"github.com/game-ops/ai-alert-system/internal/config"
	"github.com/game-ops/ai-alert-system/internal/models"
	"github.com/sethvargo/go-password/password"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	// Migrate tables one by one
	migrator := db.Migrator()
	tables := []interface{}{
		&models.User{},
		&models.Alert{},
		&models.DataSource{},
		&models.Channel{},
		&models.RouteRule{},
		&models.SilenceRule{},
		&models.OnDuty{},
		&models.AILog{},
		&models.SilenceRecommendation{},
	}

	for _, table := range tables {
		if !migrator.HasTable(table) {
			if err := migrator.CreateTable(table); err != nil {
				log.Printf("Warning: failed to create table: %v", err)
			}
		} else {
			if err := migrator.AutoMigrate(table); err != nil {
				log.Printf("Warning: failed to migrate table: %v", err)
			}
		}
	}

	// Create default admin user if not exists
	createDefaultAdminUser(db)

	log.Println("Database connection established and migrated")
	return db, nil
}

func createDefaultAdminUser(db *gorm.DB) {
	var count int64
	db.Model(&models.User{}).Count(&count)

	if count == 0 {
		pwd, err := password.Generate(64, 10, 10, false, false)
		if err != nil {
			log.Fatalf("Failed to generate admin password: %v", err)
		}
		admin := models.User{
			Username: "admin",
			Name:     "管理员",
			Email:    "admin@example.com",
			Role:     "admin",
		}
		if err := admin.SetPassword(pwd); err != nil {
			log.Printf("Warning: failed to set admin password: %v", err)
			return
		}
		if err := db.Create(&admin).Error; err != nil {
			log.Printf("Warning: failed to create default admin user: %v", err)
			return
		}
		log.Printf("Default admin user created: admin / %s", pwd)
		log.Printf("You must change password when you sign in for the first time.")
	}
}
