package models

import (
	"time"

	"gorm.io/gorm"
)

// PredictionLog stores ML inference results from the ESP32
type PredictionLog struct {
	gorm.Model
	Label         string    `json:"label" gorm:"index"`
	Confidence    float64   `json:"confidence"`
	Timestamp     time.Time `json:"timestamp" gorm:"index"`
	DismissReason string    `json:"dismiss_reason"` // Track why alert was dismissed
}

// MLPayload represents the JSON structure from ESP32-S3
type MLPayload struct {
	MLLabel      string    `json:"ml_label"`
	Confidence   float64   `json:"confidence"`
	AnomalyScore float64   `json:"anomaly_score"`
	Telemetry    Telemetry `json:"telemetry"`
}

// Telemetry contains raw sensor data
type Telemetry struct {
	VibrationPeak float64 `json:"vibration_peak"`
	CurrentAmps   float64 `json:"current_amps"`
	TemperatureC  float64 `json:"temperature_c"`
}

// IsCritical determines if this prediction should be logged
func (p *MLPayload) IsCritical() bool {
	return p.MLLabel == "bearing_fault" && p.Confidence > 0.85
}

// ToPredictionLog converts MLPayload to a PredictionLog for storage
func (p *MLPayload) ToPredictionLog() *PredictionLog {
	return &PredictionLog{
		Label:      p.MLLabel,
		Confidence: p.Confidence,
		Timestamp:  time.Now(),
	}
}

// GetFlaggedSensors analyzes telemetry data and returns sensors that crossed thresholds
func (p *MLPayload) GetFlaggedSensors() []string {
	sensors := []string{}
	if p.Telemetry.VibrationPeak > 2000 {
		sensors = append(sensors, "Vibration Sensor")
	}
	if p.Telemetry.TemperatureC > 60 {
		sensors = append(sensors, "Thermal Sensor")
	}
	if p.Telemetry.CurrentAmps > 2.0 {
		sensors = append(sensors, "Current Sensor")
	}
	return sensors
}
