# Design Document: Industrial AI Interface

### 1. Visual Logic
* **Zone A (Sidebar):**
    * **Confidence Meter:** A bar showing the AI's certainty (e.g., "98% Confidence").
    * **Raw Data:** Live Charts for Vibration & Temp.
* **Zone B (Center):**
    * **3D Motor:** Color is driven strictly by `ml_label`.
* **Zone C (Intelligence):**
    * **Status:** Displays the ML Label (e.g., "HEALTHY" or "BEARING FAULT").
    * **Explanation:** "AI detected spectral signature of Inner Race Defect."

### 2. Color System
* **Healthy:** `#00FF94`
* **Unbalance:** `#FFC107`
* **Bearing Fault:** `#FF003C`
