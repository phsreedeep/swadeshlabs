# Product Requirements Document (PRD)
## Project: Swadesh Labs - AI Predictive Maintenance Dashboard
## Goal: Win BeachHack Season 7 (Problem Statement 6)

### 1. Product Summary
A "Digital Twin" dashboard that visualizes **Edge AI inference results** from an industrial motor. It uses a quantized ML model running on an ESP32-S3 to predict failure modes before they happen.

### 2. Core Features
* **Edge ML Inference:** The dashboard does *not* calculate faults. It displays the *decision* made by the ESP32 (e.g., "Prediction: Bearing Fault (98%)").
* **Real-Time Telemetry:** Visualizes the raw data (Vibration/Current/Temp) alongside the AI prediction.
* **3D Digital Twin:**
    * **Green:** AI Prediction = "Healthy"
    * **Yellow:** AI Prediction = "Unbalance"
    * **Red:** AI Prediction = "Bearing_Fault" OR Temp > 80°C (Safety Override).
* **Automated Workflow:**
    * Trigger: When `ml_label` == "Bearing_Fault" AND `confidence` > 0.85.
    * Action: Auto-open "Work Order #101" modal with specific repair instructions.

### 3. Data Requirements
* **Source:** ESP32-S3 N16R8 running Edge Impulse library.
* **Payload:** JSON via MQTT (`swadesh/motor1/inference`).
* **Storage:** Log every "Critical" ML prediction to SQLite.
