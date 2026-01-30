package database

import (
	"log"

	"swadesh-dashboard/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize sets up SQLite database with GORM
func Initialize(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}

	// Auto-migrate the schema
	if err := DB.AutoMigrate(&models.PredictionLog{}); err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// LogPrediction saves a critical ML prediction to the database
func LogPrediction(payload *models.MLPayload) error {
	if payload.IsCritical() {
		predictionLog := payload.ToPredictionLog()
		result := DB.Create(predictionLog)
		if result.Error != nil {
			log.Printf("Failed to log prediction: %v", result.Error)
			return result.Error
		}
		log.Printf("Logged critical prediction: %s (%.2f%%)", payload.MLLabel, payload.Confidence*100)
	}
	return nil
}

// GetRecentPredictions fetches the last N predictions
func GetRecentPredictions(limit int) ([]models.PredictionLog, error) {
	var logs []models.PredictionLog
	result := DB.Order("created_at desc").Limit(limit).Find(&logs)
	return logs, result.Error
}
