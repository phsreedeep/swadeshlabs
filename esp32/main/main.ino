#include "esp_adc/adc_continuous.h"
#include <math.h>

/* ================= CONFIG ================= */

#define SAMPLE_RATE_HZ 10000   
#define FFT_SIZE 256
#define BUF_SIZE 1024

#define ADC_MAX 4095.0f
#define VREF 3.3f

#define ZERO_G_V 1.708f
#define MV_PER_G 0.300f

/* ================= THRESHOLDS FROM CSV ================= */

#define T_LOW  0.0294f
#define T_MID  0.3223f
#define T_HIGH 0.8105f

/* ================= BUFFERS ================= */

static float samples[FFT_SIZE];
static size_t sample_idx = 0;

/* ================= ADC ================= */

adc_continuous_handle_t adc_handle;

/* ================= SETUP ================= */

void setup() {
    Serial.begin(1000000);

    adc_continuous_handle_cfg_t hc = {
        .max_store_buf_size = 8192,
        .conv_frame_size = 256
    };
    adc_continuous_new_handle(&hc, &adc_handle);

    static adc_digi_pattern_config_t pattern[] = {
        {ADC_ATTEN_DB_11, ADC_CHANNEL_3, ADC_UNIT_1, ADC_BITWIDTH_12}
    };

    adc_continuous_config_t cfg = {
        .pattern_num = 1,
        .adc_pattern = pattern,
        .sample_freq_hz = SAMPLE_RATE_HZ,
        .conv_mode = ADC_CONV_SINGLE_UNIT_1,
        .format = ADC_DIGI_OUTPUT_FORMAT_TYPE2
    };

    adc_continuous_config(adc_handle, &cfg);
    adc_continuous_start(adc_handle);
}

/* ================= DFT ================= */

void compute_band_energy() {

    float low = 0, mid = 0, high = 0;

    for (int k = 1; k < FFT_SIZE / 2; k++) {

        float real = 0;
        float imag = 0;

        for (int n = 0; n < FFT_SIZE; n++) {
            float angle = 2.0f * PI * k * n / FFT_SIZE;
            real += samples[n] * cos(angle);
            imag -= samples[n] * sin(angle);
        }

        float mag2 = real * real + imag * imag;
        float freq = (k * SAMPLE_RATE_HZ) / FFT_SIZE;  // adjusted for 10 kHz

        if (freq <= 200.0f)
            low += mag2;
        else if (freq <= 800.0f)
            mid += mag2;
        else
            high += mag2;
    }

    float total = low + mid + high + 1e-9f;

    float R_low  = low / total;
    float R_mid  = mid / total;
    float R_high = high / total;

    // Print ratios
    Serial.print("RATIO | LOW=");
    Serial.print(R_low, 3);
    Serial.print(" MID=");
    Serial.print(R_mid, 3);
    Serial.print(" HIGH=");
    Serial.println(R_high, 3);

    // Decision based on thresholds from CSV
    if (R_high > T_HIGH) {
        Serial.println("FAULT DETECTED: HIGH-BAND ENERGY ELEVATED");
    } 
    else if (R_mid > T_MID) {
        Serial.println("FAULT DETECTED: MID-BAND ENERGY ELEVATED");
    } 
    else if (R_low > T_LOW) {
        Serial.println("NOTICE: LOW-BAND ENERGY ELEVATED");
    } 
    else {
        Serial.println("HEALTHY VIBRATION PROFILE");
    }
}

/* ================= LOOP ================= */

void loop() {

    uint8_t buf[BUF_SIZE];
    uint32_t bytes_read;

    if (adc_continuous_read(adc_handle, buf, sizeof(buf), &bytes_read, 0) != ESP_OK)
        return;

    int n = bytes_read / sizeof(adc_digi_output_data_t);

    for (int i = 0; i < n; i++) {

        auto *d = (adc_digi_output_data_t*)&buf[i * sizeof(adc_digi_output_data_t)];
        uint16_t val = d->type2.data & 0x0FFF;

        float v = (val / ADC_MAX) * VREF;
        float accel_g = (v - ZERO_G_V) / MV_PER_G;

        samples[sample_idx++] = accel_g;

        if (sample_idx >= FFT_SIZE) {
            compute_band_energy();
            sample_idx = 0;
        }
    }
}
