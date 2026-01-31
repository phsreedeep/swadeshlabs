package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"swadesh-dashboard/internal/database"
	"swadesh-dashboard/internal/handlers"
	"swadesh-dashboard/internal/mqtt"
)

func main() {
	// Parse command line flags
	port := flag.String("port", "8080", "Port to listen on")
	mock := flag.Bool("mock", true, "Run in mock mode with simulated data")
	flag.Parse()

	// Initialize Database
	if err := database.InitDB("predictions.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database initialized successfully")

	// Initialize Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// Static Files
	// Serve "public" folder at "/public" URL path
	e.Static("/public", "public")

	// Serve HTML views
	e.File("/", "views/index.html")

	// API Routes
	api := e.Group("/api")
	api.GET("/predictions", handlers.GetPredictions)
	api.GET("/stats", handlers.GetStats)

	// Add Alert Management Routes
	api.GET("/alerts", handlers.GetAlerts)                 // Get all active alerts
	api.POST("/alerts/:id/dismiss", handlers.DismissAlert) // Dismiss specific alert
	api.POST("/work-orders", handlers.CreateWorkOrder)     // Create ticket from alert

	// SSE Endpoint
	e.GET("/events", handlers.SSEHandler)

	// Start MQTT Client or Mock Generator
	if *mock {
		fmt.Println("[INFO] Running in MOCK mode - simulating ESP32 data")
		go mqtt.StartMockPublisher()
	} else {
		fmt.Println("[INFO] Connecting to MQTT Broker...")
		go mqtt.StartMQTTClient()
	}

	// Graceful Shutdown
	go func() {
		address := fmt.Sprintf(":%s", *port)
		fmt.Printf("[INFO] Server listening on http://localhost%s\n", address)
		fmt.Printf("[INFO] Dashboard URL: http://localhost%s\n", address)
		fmt.Printf("[INFO] SSE endpoint: http://localhost%s/events\n", address)

		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	fmt.Println("\n[INFO] Shutting down server...")
	// Cleanup
}
