#include <OneWire.h>

// Sensor Address: 28 41 0E 03 00 02 24 A5
byte addr[8] = {0x28, 0x41, 0x0E, 0x03, 0x00, 0x02, 0x24, 0xA5};
OneWire ds(14);

void setup() {
  Serial.begin(115200);
  while (!Serial); // Wait for Serial Monitor
  Serial.println("--- Raw OneWire Temperature Ping ---");
}

void loop() {
  byte data[12];

  // Step 1: Start Temperature Conversion
  ds.reset();
  ds.select(addr);
  ds.write(0x44);      // Start conversion

  delay(1000);         // Wait for conversion (750ms needed for 12-bit)

  // Step 2: Read Scratchpad
  byte present = ds.reset();
  ds.select(addr);    
  ds.write(0xBE);      // Read Scratchpad

  for (int i = 0; i < 9; i++) {
    data[i] = ds.read();
  }

  // Step 3: Convert to Celsius
  int16_t raw = (data[1] << 8) | data[0];
  float celsius = (float)raw / 16.0;

  if (present) {
    Serial.print("Sensor Address: ");
    for(int i=0; i<8; i++) { Serial.print(addr[i], HEX); Serial.print(" "); }
    Serial.print("| Temp: ");
    Serial.print(celsius);
    Serial.println("°C");
  } else {
    Serial.println("Sensor not responding to ping.");
  }

  delay(2000);
}