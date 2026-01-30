package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"swadesh-dashboard/internal/database"
	"swadesh-dashboard/internal/handlers"
	"swadesh-dashboard/internal/models"
	"swadesh-dashboard/internal/mqtt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Server port")
	broker := flag.String("broker", "tcp://localhost:1883", "MQTT broker address")
	dbPath := flag.String("db", "predictions.db", "SQLite database path")
	mockMode := flag.Bool("mock", true, "Enable mock data publisher (for testing)")
	flag.Parse()

	log.Println("==============================================")
	log.Println("  SWADESH LABS - Predictive Maintenance System")
	log.Println("==============================================")

	// Initialize database
	if err := database.Initialize(*dbPath); err != nil {
		log.Fatalf("[ERROR] Database initialization failed: %v", err)
	}

	// Create broadcast channel for SSE
	broadcast := make(chan models.MLPayload, 100)

	// Initialize SSE Hub
	sseHub := handlers.NewSSEHub(broadcast)
	go sseHub.Run()

	// Initialize MQTT client (or mock mode)
	if *mockMode {
		log.Println("[INFO] Running in MOCK mode - simulating ESP32 data")
		mqtt.StartMockPublisher(broadcast)
	} else {
		mqttClient := mqtt.NewClient(*broker, broadcast)
		if err := mqttClient.Connect(*broker); err != nil {
			log.Printf("[WARN] MQTT connection failed: %v", err)
			log.Println("[INFO] Falling back to MOCK mode")
			mqtt.StartMockPublisher(broadcast)
		} else {
			defer mqttClient.Disconnect()
		}
	}

	// Initialize Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Template renderer
	e.Renderer = handlers.NewTemplateRenderer("views/*.html")

	// Setup routes
	handlers.SetupRoutes(e, sseHub)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("\n[INFO] Shutting down gracefully...")
		e.Close()
	}()

	// Start server
	log.Printf("[INFO] Server listening on http://localhost:%s", *port)
	log.Printf("[INFO] Dashboard URL: http://localhost:%s", *port)
	log.Printf("[INFO] SSE endpoint: http://localhost:%s/events", *port)

	if err := e.Start(":" + *port); err != nil {
		log.Printf("[INFO] Server stopped: %v", err)
	}
}
