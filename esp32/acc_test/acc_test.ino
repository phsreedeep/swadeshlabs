#include "esp_adc/adc_continuous.h"

#define SAMPLE_RATE     10000
#define BURST_SAMPLES   10000
#define BUF_SIZE        512

// ADXL335 parameters
#define ADC_MAX     4095.0
#define VREF        3.3
#define ZERO_G_V    1.65
#define MV_PER_G    0.300   // ADXL335

adc_continuous_handle_t adc_handle;

void setup() {
    Serial.begin(1000000);
    delay(2000);

    adc_continuous_handle_cfg_t hc = {
        .max_store_buf_size = 4096,
        .conv_frame_size = 256,
    };
    adc_continuous_new_handle(&hc, &adc_handle);

    adc_digi_pattern_config_t p = {
        .atten = ADC_ATTEN_DB_11,
        .channel = ADC_CHANNEL_3,   // GPIO4
        .unit = ADC_UNIT_1,
        .bit_width = ADC_BITWIDTH_12
    };

    adc_continuous_config_t c = {
        .pattern_num = 1,
        .adc_pattern = &p,
        .sample_freq_hz = SAMPLE_RATE,
        .conv_mode = ADC_CONV_SINGLE_UNIT_1,
        .format = ADC_DIGI_OUTPUT_FORMAT_TYPE2,
    };

    adc_continuous_config(adc_handle, &c);
    adc_continuous_start(adc_handle);

    delay(5000);
}

void loop() {
    uint8_t buf[BUF_SIZE];
    uint32_t bytes_read;
    uint32_t burst_count = 0;

    Serial.println("adc,g");

    // -------- 1 second burst --------
    while (burst_count < BURST_SAMPLES) {
        adc_continuous_read(adc_handle, buf, sizeof(buf), &bytes_read, portMAX_DELAY);

        int n = bytes_read / sizeof(adc_digi_output_data_t);
        for (int i = 0; i < n && burst_count < BURST_SAMPLES; i++) {
            adc_digi_output_data_t *d =
                (adc_digi_output_data_t *)&buf[i * sizeof(adc_digi_output_data_t)];

            uint16_t adc = d->type2.data & 0x0FFF;
            float voltage = (adc / ADC_MAX) * VREF;
            float g = (voltage - ZERO_G_V) / MV_PER_G;

            Serial.print(adc);
            Serial.print(",");
            Serial.println(g, 4);

            burst_count++;
        }
    }

    // -------- burst report --------
    Serial.print("BURST_SAMPLES_CAPTURED=");
    Serial.println(burst_count);

    delay(4000);
}
