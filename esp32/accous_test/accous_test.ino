#include "esp_adc/adc_continuous.h"

// Audio Sampling Parameters
#define SAMPLE_RATE     10000    // 10 kHz sample rate (max freq ~5 kHz)
#define BURST_SAMPLES   10000
#define BUF_SIZE        512

// MAX4466 / ADC Parameters
#define ADC_MAX     4095.0
#define VREF        3.3          // Supply voltage

adc_continuous_handle_t adc_handle;

void setup() {
    Serial.begin(1000000); // Fast baud rate for high-speed logging
    delay(2000);

    // 1. Configure the ADC Handle
    adc_continuous_handle_cfg_t hc = {
        .max_store_buf_size = 4096,
        .conv_frame_size = 256,
    };
    adc_continuous_new_handle(&hc, &adc_handle);

    // 2. Configure the Channel (GPIO3 = ADC1_CHANNEL_2)
    adc_digi_pattern_config_t p = {
        .atten = ADC_ATTEN_DB_12,    // Range: 0V - 3.3V (Using DB_12 for newer IDF versions, or DB_11)
        .channel = ADC_CHANNEL_2,    // CORRECT FOR GPIO3 on ESP32-S3
        .unit = ADC_UNIT_1,
        .bit_width = ADC_BITWIDTH_12
    };

    // 3. Configure the Continuous Mode
    adc_continuous_config_t c = {
        .pattern_num = 1,
        .adc_pattern = &p,
        .sample_freq_hz = SAMPLE_RATE,
        .conv_mode = ADC_CONV_SINGLE_UNIT_1,
        .format = ADC_DIGI_OUTPUT_FORMAT_TYPE2,
    };

    adc_continuous_config(adc_handle, &c);
    adc_continuous_start(adc_handle);

    delay(1000);
}

void loop() {
    uint8_t buf[BUF_SIZE];
    uint32_t bytes_read;
    uint32_t burst_count = 0;

    Serial.println("adc,voltage");

    // -------- 1 second burst --------
    while (burst_count < BURST_SAMPLES) {
        // Read data from the ADC DMA buffer
        adc_continuous_read(adc_handle, buf, sizeof(buf), &bytes_read, portMAX_DELAY);

        // Process the buffer
        int n = bytes_read / sizeof(adc_digi_output_data_t);
        for (int i = 0; i < n && burst_count < BURST_SAMPLES; i++) {
            adc_digi_output_data_t *d =
                (adc_digi_output_data_t *)&buf[i * sizeof(adc_digi_output_data_t)];

            // Extract 12-bit Raw Data
            uint16_t adc = d->type2.data & 0x0FFF;
            
            // Convert to Voltage
            float voltage = (adc / ADC_MAX) * VREF;

            // Output CSV format: Raw Value, Voltage
            Serial.print(adc);
            Serial.print(",");
            Serial.println(voltage, 4);

            burst_count++;
        }
    }

    // -------- burst report --------
    Serial.print("BURST_SAMPLES_CAPTURED=");
    Serial.println(burst_count);

    delay(4000);
}