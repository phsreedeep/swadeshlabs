#include "esp_adc/adc_continuous.h"
#include <OneWire.h>

/* ================= CONFIG ================= */

#define ADC_RATE_HZ     20000
#define BUF_SIZE        1024
#define MAX_SAMPLES     1000 

#define ADC_MAX         4095.0
#define VREF            3.3

#define ZERO_G_V        1.708
#define MV_PER_G        0.300

#define CT_OFFSET       1300  // Tares 1290-1310 noise floor to 0

#define ONEWIRE_PIN     14
OneWire ds(ONEWIRE_PIN);

/* ================= DATA FRAME ================= */

#pragma pack(push, 1)
typedef struct {
    uint32_t ts_us;    // Microseconds since program start
    float audio_v;
    float accel_g;
    uint16_t ct_adc;
    float temp_c;
} sample_t;
#pragma pack(pop)

/* ================= GLOBALS ================= */

adc_continuous_handle_t adc_handle;
sample_t sample_buf[MAX_SAMPLES];
uint16_t sample_idx = 0;

float current_temp = NAN;
uint32_t last_temp_ms = 0;
uint32_t start_time_us = 0; // Reference for manageable timestamps

/* ================= NON-BLOCKING TEMP ================= */

void update_temp_async() {
    static enum { IDLE, CONVERTING } state = IDLE;
    static uint32_t conversion_start_ms = 0;

    if (state == IDLE) {
        if (millis() - last_temp_ms > 2000) {
            if (ds.reset()) {
                ds.skip();
                ds.write(0x44); 
                conversion_start_ms = millis();
                state = CONVERTING;
            }
        }
    } 
    else if (state == CONVERTING) {
        if (millis() - conversion_start_ms >= 750) {
            if (ds.reset()) {
                ds.skip();
                ds.write(0xBE);
                byte data[9];
                for (int i = 0; i < 9; i++) data[i] = ds.read();
                int16_t raw = (data[1] << 8) | data[0];
                current_temp = raw / 16.0;
            }
            last_temp_ms = millis();
            state = IDLE;
        }
    }
}

/* ================= SETUP ================= */

void setup() {
    Serial.begin(1000000);
    start_time_us = micros(); // Mark the start to keep ts_us manageable
    
    adc_continuous_handle_cfg_t hc = {
        .max_store_buf_size = 8192,
        .conv_frame_size = 256,
    };
    adc_continuous_new_handle(&hc, &adc_handle);

    static adc_digi_pattern_config_t pattern[3];
    pattern[0] = {ADC_ATTEN_DB_11, ADC_CHANNEL_2, ADC_UNIT_1, ADC_BITWIDTH_12};
    pattern[1] = {ADC_ATTEN_DB_11, ADC_CHANNEL_3, ADC_UNIT_1, ADC_BITWIDTH_12};
    pattern[2] = {ADC_ATTEN_DB_11, ADC_CHANNEL_4, ADC_UNIT_1, ADC_BITWIDTH_12};

    adc_continuous_config_t cfg = {
        .pattern_num = 3,
        .adc_pattern = pattern,
        .sample_freq_hz = ADC_RATE_HZ,
        .conv_mode = ADC_CONV_SINGLE_UNIT_1,
        .format = ADC_DIGI_OUTPUT_FORMAT_TYPE2,
    };

    adc_continuous_config(adc_handle, &cfg);
    adc_continuous_start(adc_handle);
}

/* ================= LOOP ================= */

void loop() {
    uint8_t buf[BUF_SIZE];
    uint32_t bytes_read;

    // Handle DS18B20 without stopping the ADC processing
    update_temp_async();

    esp_err_t ret = adc_continuous_read(adc_handle, buf, sizeof(buf), &bytes_read, 0);

    if (ret == ESP_OK && bytes_read > 0) {
        int n = bytes_read / sizeof(adc_digi_output_data_t);

        for (int i = 0; i < n; i++) {
            auto *d = (adc_digi_output_data_t*)&buf[i * sizeof(adc_digi_output_data_t)];
            uint16_t val = d->type2.data & 0x0FFF;
            float v = (val / ADC_MAX) * VREF;

            if (d->type2.channel == ADC_CHANNEL_2) {
                // Time relative to start_time_us to keep numbers small
                sample_buf[sample_idx].ts_us = micros() - start_time_us;
                sample_buf[sample_idx].audio_v = v;
                sample_buf[sample_idx].temp_c = current_temp;
                
                sample_idx++;
                if (sample_idx >= MAX_SAMPLES) {
                    uint8_t marker = 0xAA;
                    uint16_t count = sample_idx;
                    Serial.write(&marker, 1);
                    Serial.write((uint8_t*)&count, sizeof(count));
                    Serial.write((uint8_t*)sample_buf, count * sizeof(sample_t));
                    Serial.flush();
                    sample_idx = 0;
                }
            }
            else if (d->type2.channel == ADC_CHANNEL_3) {
                sample_buf[sample_idx].accel_g = (v - ZERO_G_V) / MV_PER_G;
            }
            else if (d->type2.channel == ADC_CHANNEL_4) {
                // Apply CT Offset and Floor at 0
                if (val > CT_OFFSET) {
                    sample_buf[sample_idx].ct_adc = val - CT_OFFSET;
                } else {
                    sample_buf[sample_idx].ct_adc = 0;
                }
            }
        }
    }
}