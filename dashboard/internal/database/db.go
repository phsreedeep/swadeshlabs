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

// UpdateDismissReason updates the dismiss_reason field for a given prediction log
func UpdateDismissReason(id string, reason string) error {
	result := DB.Model(&models.PredictionLog{}).Where("id = ?", id).Update("dismiss_reason", reason)
	if result.Error != nil {
		log.Printf("Failed to update dismiss reason: %v", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		log.Printf("No prediction log found with ID: %s", id)
		return gorm.ErrRecordNotFound
	}
	log.Printf("Updated dismiss reason for alert #%s: %s", id, reason)
	return nil
}
