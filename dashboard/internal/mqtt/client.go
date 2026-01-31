package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"swadesh-dashboard/internal/database"
	"swadesh-dashboard/internal/models"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	DefaultBroker = "tcp://localhost:1883"
	Topic         = "swadesh/motor1/inference"
	ClientID      = "swadesh-dashboard"
)

// Client wraps the MQTT client with payload broadcasting
type Client struct {
	client     mqtt.Client
	broadcast  chan models.MLPayload
	lastPayload *models.MLPayload
}

// NewClient creates a new MQTT client
func NewClient(broker string, broadcast chan models.MLPayload) *Client {
	if broker == "" {
		broker = DefaultBroker
	}

	return &Client{
		broadcast: broadcast,
	}
}

// Connect establishes connection to the MQTT broker
func (c *Client) Connect(broker string) error {
	if broker == "" {
		broker = DefaultBroker
	}

	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(ClientID).
		SetAutoReconnect(true).
		SetConnectionLostHandler(func(client mqtt.Client, err error) {
			log.Printf("MQTT connection lost: %v", err)
		}).
		SetOnConnectHandler(func(client mqtt.Client) {
			log.Println("MQTT connected, subscribing to topic...")
			c.subscribe()
		})

	c.client = mqtt.NewClient(opts)

	token := c.client.Connect()
	token.Wait()

	if token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	return nil
}

// subscribe sets up the message handler for the inference topic
func (c *Client) subscribe() {
	token := c.client.Subscribe(Topic, 1, func(client mqtt.Client, msg mqtt.Message) {
		c.handleMessage(msg.Payload())
	})
	token.Wait()

	if token.Error() != nil {
		log.Printf("Failed to subscribe to %s: %v", Topic, token.Error())
	} else {
		log.Printf("Subscribed to topic: %s", Topic)
	}
}

// handleMessage parses the ML payload and broadcasts it
func (c *Client) handleMessage(payload []byte) {
	var mlPayload models.MLPayload
	if err := json.Unmarshal(payload, &mlPayload); err != nil {
		log.Printf("Failed to parse MQTT payload: %v", err)
		return
	}

	log.Printf("Received ML inference: %s (%.2f%% confidence)", 
		mlPayload.MLLabel, mlPayload.Confidence*100)

	// Store critical predictions
	if err := database.LogPrediction(&mlPayload); err != nil {
		log.Printf("Failed to log prediction: %v", err)
	}

	// Store last payload for new SSE connections
	c.lastPayload = &mlPayload

	// Broadcast to all SSE clients
	select {
	case c.broadcast <- mlPayload:
	default:
		log.Println("Broadcast channel full, dropping message")
	}
}

// GetLastPayload returns the most recent payload (for new SSE clients)
func (c *Client) GetLastPayload() *models.MLPayload {
	return c.lastPayload
}

// Disconnect gracefully closes the MQTT connection
func (c *Client) Disconnect() {
	if c.client != nil && c.client.IsConnected() {
		c.client.Disconnect(250)
		log.Println("MQTT disconnected")
	}
}

// StartMockPublisher simulates ESP32 data for testing (development only)
func StartMockPublisher(broadcast chan models.MLPayload) {
	labels := []string{"healthy", "unbalance", "bearing_fault"}
	labelIdx := 0

	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for range ticker.C {
			// Cycle through labels for demo
			label := labels[labelIdx]
			labelIdx = (labelIdx + 1) % len(labels)

			confidence := 0.85 + (float64(time.Now().UnixNano()%15) / 100.0)
			if confidence > 0.99 {
				confidence = 0.99
			}

			payload := models.MLPayload{
				MLLabel:      label,
				Confidence:   confidence,
				AnomalyScore: 0.1 + (float64(time.Now().UnixNano()%20) / 100.0),
				Telemetry: models.Telemetry{
					VibrationPeak: 300 + float64(time.Now().UnixNano()%200),
					CurrentAmps:   1.0 + (float64(time.Now().UnixNano()%50) / 100.0),
					TemperatureC:  45 + float64(time.Now().UnixNano()%30),
				},
			}

			log.Printf("[MOCK] Inference: %s (%.1f%% confidence)", payload.MLLabel, payload.Confidence*100)
			
			// Log critical predictions
			database.LogPrediction(&payload)
			
			broadcast <- payload
		}
	}()
}
