package database

import (
	"fmt"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// PredictionLog stores ML inference results
type PredictionLog struct {
	gorm.Model
	Label        string    `json:"label" gorm:"index"`
	Confidence   float64   `json:"confidence"`
	Timestamp    time.Time `json:"timestamp" gorm:"index"`
	AnomalyScore float64   `json:"anomaly_score"`

	// Telemetry Snapshot
	VibrationPeak float64 `json:"vibration_peak"`
	Temperature   float64 `json:"temperature"`

	// Alert Management
	IsAlert       bool       `json:"is_alert" gorm:"default:false"`
	Dismissed     bool       `json:"dismissed" gorm:"default:false"`
	DismissReason string     `json:"dismiss_reason"`
	DismissedAt   *time.Time `json:"dismissed_at"`
}

func InitDB(dbPath string) error {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// Auto Migrate
	return DB.AutoMigrate(&PredictionLog{})
}

func SavePrediction(label string, confidence float64, anomalyScore float64, vib float64, temp float64) uint {
	isAlert := label == "bearing_fault" && confidence > 0.8

	log := PredictionLog{
		Label:         label,
		Confidence:    confidence,
		Timestamp:     time.Now(),
		AnomalyScore:  anomalyScore,
		VibrationPeak: vib,
		Temperature:   temp,
		IsAlert:       isAlert,
	}

	result := DB.Create(&log)
	if result.Error != nil {
		fmt.Println("Error saving prediction:", result.Error)
		return 0
	}

	if isAlert {
		fmt.Printf("Logged critical prediction: %s (%.2f%%)\n", label, confidence*100)
	}

	return log.ID
}

func GetRecentPredictions(limit int) ([]PredictionLog, error) {
	var logs []PredictionLog
	result := DB.Order("timestamp desc").Limit(limit).Find(&logs)
	return logs, result.Error
}

// DismissAlert marks a prediction as dismissed
func DismissAlert(id string, reason string) error {
	return DB.Model(&PredictionLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"dismissed":      true,
		"dismiss_reason": reason,
		"dismissed_at":   time.Now(),
	}).Error
}
