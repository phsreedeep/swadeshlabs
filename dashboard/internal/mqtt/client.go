package mqtt

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"swadesh-dashboard/internal/database"
	"swadesh-dashboard/internal/handlers"
)

// MLPayload matches the JSON sent by ESP32
type MLPayload struct {
	MLLabel      string    `json:"ml_label"`
	Confidence   float64   `json:"confidence"`
	AnomalyScore float64   `json:"anomaly_score"`
	Telemetry    Telemetry `json:"telemetry"`
}

type Telemetry struct {
	VibrationPeak float64 `json:"vibration_peak"` // micrometers
	TemperatureC  float64 `json:"temperature_c"`  // Celsius
	CurrentAmps   float64 `json:"current_amps"`   // Amps
}

// StartMQTTClient (Placeholder for real MQTT)
func StartMQTTClient() {
	// In real implementation: connect to broker, subscribe to topic
	// On message: Parse JSON -> Save to DB -> Broadcast SSE
}

// StartMockPublisher simulates the ESP32 sending data every 3 seconds
func StartMockPublisher() {
	states := []string{"healthy", "healthy", "healthy", "unbalance", "unbalance", "bearing_fault", "bearing_fault"}
	// Weighted to show faults for testing

	idx := 0

	for {
		// Cycle through states
		currentState := states[idx%len(states)]
		idx++

		// Generate realistic data based on state
		// Reduced frequency to avoid spamming while testing popup
		payload := generateMockPayload(currentState)

		// 1. Save to DB
		// Also capture the ID so we can send it to frontend
		alertID := database.SavePrediction(
			payload.MLLabel,
			payload.Confidence,
			payload.AnomalyScore,
			payload.Telemetry.VibrationPeak,
			payload.Telemetry.TemperatureC,
		)

		// 2. Broadcast via SSE
		// Allow adding the ID to the payload for the frontend to use
		jsonBytes, _ := json.Marshal(payload)

		// Broadcast
		handlers.BroadcastMessage(jsonBytes)

		fmt.Printf("[MOCK] Inference: %s (%.1f%% confidence) [ID: %d]\n", currentState, payload.Confidence*100, alertID)

		time.Sleep(3 * time.Second)
	}
}

func generateMockPayload(label string) MLPayload {
	t := Telemetry{}
	confidence := 0.0
	anomaly := 0.0

	switch label {
	case "healthy":
		t.VibrationPeak = 150 + rand.Float64()*50 // 150-200 um
		t.TemperatureC = 45 + rand.Float64()*5    // 45-50 C
		t.CurrentAmps = 1.2 + rand.Float64()*0.1
		confidence = 0.85 + rand.Float64()*0.14 // 85-99%
		anomaly = rand.Float64() * 0.2          // Low anomaly

	case "unbalance":
		t.VibrationPeak = 600 + rand.Float64()*200 // 600-800 um
		t.TemperatureC = 55 + rand.Float64()*10
		t.CurrentAmps = 1.8 + rand.Float64()*0.3
		confidence = 0.80 + rand.Float64()*0.15
		anomaly = 0.4 + rand.Float64()*0.3

	case "bearing_fault":
		t.VibrationPeak = 1200 + rand.Float64()*800 // High vibration
		t.TemperatureC = 65 + rand.Float64()*15     // High temp
		t.CurrentAmps = 2.5 + rand.Float64()*0.5
		confidence = 0.85 + rand.Float64()*0.14 // High confidence
		anomaly = 0.8 + rand.Float64()*0.2      // High anomaly
	}

	return MLPayload{
		MLLabel:      label,
		Confidence:   confidence,
		AnomalyScore: anomaly,
		Telemetry:    t,
	}
}
