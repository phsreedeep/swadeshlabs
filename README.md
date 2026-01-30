# Swadesh Labs - AI Predictive Maintenance Dashboard

A "Digital Twin" dashboard that visualizes Edge AI inference results from an industrial motor running on ESP32-S3.

![Dashboard Preview](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)
![Echo](https://img.shields.io/badge/Echo-v4-00ADD8?style=flat)
![HTMX](https://img.shields.io/badge/HTMX-1.9-3366cc?style=flat)

## Features

- **Real-Time ML Inference Display**: Visualizes predictions from ESP32-S3 (Healthy, Unbalance, Bearing Fault)
- **3D Digital Twin**: Interactive Spline 3D motor with color-coded status
- **Live Telemetry**: Charts for Vibration, Temperature, and Current
- **SSE (Server-Sent Events)**: Real-time data streaming to frontend
- **MQTT Integration**: Listens to `swadesh/motor1/inference` topic
- **Automated Work Orders**: Auto-triggers when critical faults are detected
- **SQLite Logging**: Stores critical ML predictions

## Project Structure

```
swadesh-dashboard/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── database/
│   │   └── db.go            # SQLite/GORM setup
│   ├── handlers/
│   │   ├── routes.go        # HTTP routes & templates
│   │   └── sse.go           # SSE hub implementation
│   ├── models/
│   │   └── prediction.go    # Data models (PredictionLog, MLPayload)
│   └── mqtt/
│       └── client.go        # MQTT client & mock publisher
├── public/
│   ├── css/
│   │   └── style.css        # Dashboard styles
│   └── js/
│       └── app.js           # Frontend logic (SSE, Charts, UI)
├── views/
│   ├── index.html           # Main dashboard template
│   ├── status_card.html     # HTMX partial for AI status
│   └── work_order_modal.html # Work order modal template
├── go.mod
└── README.md
```

## Quick Start

### Prerequisites

- Go 1.21 or later
- (Optional) MQTT Broker (Mosquitto) for real hardware

### Installation

```bash
# Clone or navigate to project directory
cd swadeshlabs

# Download dependencies
go mod tidy

# Run the server (mock mode by default)
go run cmd/server/main.go
```

### Access the Dashboard

Open your browser and navigate to:
- **Dashboard**: http://localhost:8080
- **SSE Stream**: http://localhost:8080/events

## Configuration

### Command Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-port` | `8080` | Server port |
| `-broker` | `tcp://localhost:1883` | MQTT broker address |
| `-db` | `predictions.db` | SQLite database path |
| `-mock` | `true` | Enable mock data publisher |

### Examples

```bash
# Run with custom port
go run cmd/server/main.go -port 3000

# Connect to real MQTT broker
go run cmd/server/main.go -mock=false -broker tcp://192.168.1.100:1883

# Production build
go build -o swadesh-dashboard cmd/server/main.go
./swadesh-dashboard -mock=false
```

## ML Data Payload Format

The ESP32-S3 sends JSON payloads via MQTT to `swadesh/motor1/inference`:

```json
{
  "ml_label": "bearing_fault",
  "confidence": 0.96,
  "anomaly_score": 0.2,
  "telemetry": {
    "vibration_peak": 450,
    "current_amps": 1.2,
    "temperature_c": 55.0
  }
}
```

### ML Labels

| Label | Color | Description |
|-------|-------|-------------|
| `healthy` | Green (#00FF94) | Motor operating normally |
| `unbalance` | Yellow (#FFC107) | Rotational imbalance detected |
| `bearing_fault` | Red (#FF003C) | Inner race defect detected |

## Work Order Trigger

A work order modal automatically appears when:
- `ml_label` == `"bearing_fault"`
- `confidence` > `0.85` (85%)

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Main dashboard page |
| GET | `/events` | SSE endpoint for real-time data |
| GET | `/api/predictions` | Recent prediction logs (JSON) |
| GET | `/partials/status-card` | HTMX partial for status card |
| GET | `/partials/work-order` | HTMX partial for work order modal |

## Technology Stack

- **Backend**: Go, Echo v4, GORM, SQLite, paho.mqtt.golang
- **Frontend**: HTMX, Chart.js, Spline 3D Viewer
- **Real-Time**: Server-Sent Events (SSE)
- **Hardware**: ESP32-S3 N16R8 with Edge Impulse ML

## Development

### Mock Mode

By default, the server runs in mock mode which simulates ESP32 data every 3 seconds, cycling through:
1. `healthy` (85-99% confidence)
2. `unbalance` (85-99% confidence)
3. `bearing_fault` (85-99% confidence)

This allows frontend development without hardware.

### Building for Production

```bash
# Build binary
go build -ldflags="-s -w" -o swadesh-dashboard cmd/server/main.go

# Run in production
./swadesh-dashboard -mock=false -broker tcp://your-mqtt-broker:1883
```

## License

MIT License - Swadesh Labs (BeachHack Season 7)
