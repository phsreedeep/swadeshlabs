#include <stdio.h>
#include <string.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_system.h"
#include "nvs_flash.h"
#include "esp_adc/adc_continuous.h"
#include "esp_timer.h"
#include "esp_log.h"
#include "driver/uart.h"
#include "driver/gpio.h"

// C++ compatibility for the DS18B20 C library
extern "C" {
    #include "ds18b20.h"
}

// Edge Impulse Header
#include "edge-impulse-sdk/classifier/ei_run_classifier.h"

static const char *TAG = "EI_S3_MOTOR_MONITOR";

/* ================= CONFIG ================= */
#define DS_PIN          14
#define ADC_RATE_HZ     20000
#define MAX_SAMPLES     1000 
#define ADC_MAX         4095.0
#define VREF            3.3
#define ZERO_G_V        1.708
#define MV_PER_G        0.300
#define CT_OFFSET       1300

/* ================= DATA STRUCTURES ================= */
typedef struct {
    float audio_v;
    float accel_g;
    float ct_val;
    float temp_c;
} sample_t;

sample_t sample_buf[MAX_SAMPLES];
uint16_t sample_idx = 0;
float current_global_temp = 0.0f; // Shared variable for temperature
adc_continuous_handle_t adc_handle;

/* ================= TEMPERATURE TASK ================= */
// Non-blocking background task for DS18B20

void ds18b20_task(void *pvParameters) {
    // Use the exact names from your ds18b20.h (lines 37 and 52)
    ds18b20_init(DS_PIN); 

    while (1) {
        // High-level function defined on line 52
        current_global_temp = ds18b20_get_temp(); 
        
        if (current_global_temp == DEVICE_DISCONNECTED_C) {
            printf("DS18B20: Sensor disconnected\n");
        }

        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}
/* ================= INFERENCE BRIDGE ================= */
int get_signal_data(size_t offset, size_t length, float *out_ptr) {
    for (size_t i = 0; i < length; i++) {
        out_ptr[i * 4 + 0] = sample_buf[offset + i].audio_v;
        out_ptr[i * 4 + 1] = sample_buf[offset + i].accel_g;
        out_ptr[i * 4 + 2] = sample_buf[offset + i].ct_val;
        out_ptr[i * 4 + 3] = sample_buf[offset + i].temp_c;
    }
    return 0;
}

void run_inference() {
    ei_impulse_result_t result = { 0 };
    ei_signal_t signal = { .get_data = &get_signal_data, .total_length = MAX_SAMPLES };

    if (run_classifier(&signal, &result, false) == EI_IMPULSE_OK) {
        ESP_LOGI(TAG, "Temp: %.1f C | Motor Status:", current_global_temp);
        for (size_t ix = 0; ix < EI_CLASSIFIER_LABEL_COUNT; ix++) {
            printf("  %s: %.2f ", result.classification[ix].label, result.classification[ix].value);
        }
        printf("\n");
    }
}

/* ================= ADC INIT ================= */
void init_adc() {
    adc_continuous_handle_cfg_t hc = { .max_store_buf_size = 8192, .conv_frame_size = 256 };
    adc_continuous_new_handle(&hc, &adc_handle);

    adc_digi_pattern_config_t pattern[3] = {
        { .atten = ADC_ATTEN_DB_12, .channel = ADC_CHANNEL_2, .unit = ADC_UNIT_1, .bit_width = ADC_BITWIDTH_12 },
        { .atten = ADC_ATTEN_DB_12, .channel = ADC_CHANNEL_3, .unit = ADC_UNIT_1, .bit_width = ADC_BITWIDTH_12 },
        { .atten = ADC_ATTEN_DB_12, .channel = ADC_CHANNEL_4, .unit = ADC_UNIT_1, .bit_width = ADC_BITWIDTH_12 }
    };

    adc_continuous_config_t cfg = {
        .pattern_num = 3, .adc_pattern = pattern, .sample_freq_hz = ADC_RATE_HZ,
        .conv_mode = ADC_CONV_SINGLE_UNIT_1, .format = ADC_DIGI_OUTPUT_FORMAT_TYPE2
    };

    adc_continuous_config(adc_handle, &cfg);
    adc_continuous_start(adc_handle);
}

/* ================= MAIN ENTRY ================= */
extern "C" void app_main(void) {
    nvs_flash_init();
    init_adc();

    // Start DS18B20 Task on Core 0
    xTaskCreatePinnedToCore(&ds18b20_task, "ds18b20_task", 2048, NULL, 5, NULL, 0);

    uint8_t result_buf[1024];
    uint32_t bytes_read = 0;

    while (1) {
        if (adc_continuous_read(adc_handle, result_buf, 1024, &bytes_read, pdMS_TO_TICKS(10)) == ESP_OK) {
            int n = bytes_read / sizeof(adc_digi_output_data_t);
            for (int i = 0; i < n; i++) {
                adc_digi_output_data_t *d = (adc_digi_output_data_t*)&result_buf[i * sizeof(adc_digi_output_data_t)];
                uint16_t val = d->type2.data & 0x0FFF;
                float v = (val / ADC_MAX) * VREF;

                if (d->type2.channel == ADC_CHANNEL_2) {
                    sample_buf[sample_idx].audio_v = v;
                } else if (d->type2.channel == ADC_CHANNEL_3) {
                    sample_buf[sample_idx].accel_g = (v - ZERO_G_V) / MV_PER_G;
                } else if (d->type2.channel == ADC_CHANNEL_4) {
                    sample_buf[sample_idx].ct_val = (val > CT_OFFSET) ? (float)(val - CT_OFFSET) : 0.0f;
                    sample_buf[sample_idx].temp_c = current_global_temp; // Using data from task
                    sample_idx++;
                }

                if (sample_idx >= MAX_SAMPLES) {
                    run_inference();
                    sample_idx = 0;
                }
            }
        }
        vTaskDelay(1);
    }
}
