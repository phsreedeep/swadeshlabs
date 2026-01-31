# Tech Stack (Go + HTMX + Edge AI Edition)

### 1. Backend (Go)
* **Framework:** Echo (v4).
* **Database:** SQLite (GORM).
* **MQTT:** `paho.mqtt.golang`.
* **Real-Time:** SSE (Server-Sent Events).

### 2. Frontend (Hypermedia)
* **HTMX:** Handles Status Card swaps and Modal popups.
* **JavaScript:** Updates Charts and Spline 3D model based on ML confidence scores.

### 3. The ML Data Payload (JSON)
The ESP32-S3 will send this exact structure. The Backend must parse it:
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
