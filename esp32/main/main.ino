#include "esp_adc/adc_continuous.h"
#include <OneWire.h>

/* ================= CONFIG ================= */

#define SAMPLE_RATE     3000
#define BUF_SIZE        512
#define ADC_MAX         4095.0
#define VREF            3.3

#define SAMPLE_WINDOW_MS 1000
#define REST_WINDOW_MS   4000

// ADXL335
#define ZERO_G_V        1.708
#define MV_PER_G        0.300

// OneWire
#define ONEWIRE_PIN 14
byte ds_addr[8] = {0x28,0x41,0x0E,0x03,0x00,0x02,0x24,0xA5};

/* ================= DATA STRUCT ================= */

typedef struct {
    uint32_t ts_us;
    float audio_v;
    float accel_g;
    uint16_t ct_adc;
    float temp_c;
} sample_t;

#define MAX_SAMPLES 2048
sample_t sample_buf[MAX_SAMPLES];
volatile uint16_t sample_idx = 0;

/* ================= GLOBALS ================= */

adc_continuous_handle_t adc_handle;
OneWire ds(ONEWIRE_PIN);

float temp_c = NAN;
uint32_t last_temp_ms = 0;

bool sampling_active = false;
uint32_t window_start_ms = 0;

/* ================= DS18B20 ================= */

float read_ds18b20() {
    byte data[9];

    ds.reset();
    ds.select(ds_addr);
    ds.write(0x44);
    delay(750);

    ds.reset();
    ds.select(ds_addr);
    ds.write(0xBE);

    for (int i = 0; i < 9; i++) data[i] = ds.read();

    int16_t raw = (data[1] << 8) | data[0];
    return raw / 16.0;
}

/* ================= SETUP ================= */

void setup() {
    Serial.begin(1000000);
    delay(2000);

    adc_continuous_handle_cfg_t hc = {
        .max_store_buf_size = 4096,
        .conv_frame_size = 256,
    };
    adc_continuous_new_handle(&hc, &adc_handle);

    static adc_digi_pattern_config_t pattern[3];

    pattern[0] = { ADC_ATTEN_DB_12, ADC_CHANNEL_2, ADC_UNIT_1, ADC_BITWIDTH_12 };
    pattern[1] = { ADC_ATTEN_DB_11, ADC_CHANNEL_3, ADC_UNIT_1, ADC_BITWIDTH_12 };
    pattern[2] = { ADC_ATTEN_DB_11, ADC_CHANNEL_4, ADC_UNIT_1, ADC_BITWIDTH_12 };

    adc_continuous_config_t cfg = {
        .pattern_num = 3,
        .adc_pattern = pattern,
        .sample_freq_hz = SAMPLE_RATE,
        .conv_mode = ADC_CONV_SINGLE_UNIT_1,
        .format = ADC_DIGI_OUTPUT_FORMAT_TYPE2,
    };

    adc_continuous_config(adc_handle, &cfg);

    sampling_active = true;
    window_start_ms = millis();
    adc_continuous_start(adc_handle);
}

/* ================= LOOP ================= */

void loop() {
    uint32_t now_ms = millis();

    if (sampling_active && (now_ms - window_start_ms >= SAMPLE_WINDOW_MS)) {
        adc_continuous_stop(adc_handle);
        sampling_active = false;
        window_start_ms = now_ms;
    }

    if (!sampling_active && (now_ms - window_start_ms >= REST_WINDOW_MS)) {
        sample_idx = 0;
        sampling_active = true;
        window_start_ms = now_ms;
        adc_continuous_start(adc_handle);
    }

    if (!sampling_active) return;

    if (now_ms - last_temp_ms > 2000) {
        temp_c = read_ds18b20();
        last_temp_ms = now_ms;
    }

    uint8_t buf[BUF_SIZE];
    uint32_t bytes_read;

    adc_continuous_read(adc_handle, buf, sizeof(buf), &bytes_read, 0);

    int n = bytes_read / sizeof(adc_digi_output_data_t);

    uint16_t audio_adc = 0, accel_adc = 0, ct_adc = 0;
    bool got_audio = false, got_accel = false, got_ct = false;

    for (int i = 0; i < n; i++) {
        adc_digi_output_data_t *d =
            (adc_digi_output_data_t *)&buf[i * sizeof(adc_digi_output_data_t)];

        uint16_t val = d->type2.data & 0x0FFF;

        switch (d->type2.channel) {
            case ADC_CHANNEL_2: audio_adc = val; got_audio = true; break;
            case ADC_CHANNEL_3: accel_adc = val; got_accel = true; break;
            case ADC_CHANNEL_4: ct_adc    = val; got_ct    = true; break;
        }

        if (got_audio && got_accel && got_ct && sample_idx < MAX_SAMPLES) {
            float audio_v = (audio_adc / ADC_MAX) * VREF;
            float accel_v = (accel_adc / ADC_MAX) * VREF;
            float accel_g = (accel_v - ZERO_G_V) / MV_PER_G;

            sample_buf[sample_idx++] = {
                micros(),
                audio_v,
                accel_g,
                ct_adc,
                temp_c
            };

            got_audio = got_accel = got_ct = false;
        }
    }
}
