const int pinZ = 6; // Z-axis usually shows vertical motor vibration best
const int SAMPLES = 256;
int buffer[SAMPLES];

void setup() {
  Serial.begin(115200);
  while (!Serial);
  analogReadResolution(12);
}

void loop() {
  // Capture a high-speed burst
  for (int i = 0; i < SAMPLES; i++) {
    buffer[i] = analogRead(pinZ);
    delayMicroseconds(200); // Sample at approx 5kHz
  }

  // Find the peak-to-peak amplitude (intensity of vibration)
  int maxVal = 0;
  int minVal = 4095;
  for (int i = 0; i < SAMPLES; i++) {
    if (buffer[i] > maxVal) maxVal = buffer[i];
    if (buffer[i] < minVal) minVal = buffer[i];
  }

  int intensity = maxVal - minVal;
  Serial.print("Vibration Intensity: ");
  Serial.println(intensity);

  delay(100); // Gap between bursts
}