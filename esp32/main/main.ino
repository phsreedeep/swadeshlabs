#include "esp_adc/adc_continuous.h"
#include "driver/adc.h"
#include <OneWire.h>
#include <DallasTemperature.h>

/* ================= CONFIG ================= */
#define SAMPLE_RATE_HZ   40000
#define DMA_BUF_SIZE    4096
#define FRAME_SIZE      1024

#define MIC_CH   ADC_CHANNEL_2   // ADC1_CH2
#define ACC_CH   ADC_CHANNEL_3   // ADC1_CH3

#define BAUD_RATE 2000000

#define SYNC_BYTE_MIC  0xA5
#define SYNC_BYTE_ACC  0xA6
#define TEMP_SYNC_BYTE 0xB6
#define DIAG_SYNC_BYTE 0xC7

#define ONE_WIRE_BUS 14
#define TEMP_READ_INTERVAL_MS 1000
#define DIAG_INTERVAL_MS 5000

/* ================= ADC HANDLE ================= */
adc_continuous_handle_t adc_handle = NULL;

/* ================= DS18B20 ================= */
OneWire oneWire(ONE_WIRE_BUS);
DallasTemperature tempSensor(&oneWire);

unsigned long lastTempRead = 0;
unsigned long lastDiagSend = 0;
float currentTemperature = -127.0f;
bool temp_sensor_available = false;

/* ================= DC BLOCKERS ================= */
float mic_y = 0.0f, mic_x_prev = 0.0f;
float acc_y = 0.0f, acc_x_prev = 0.0f;

/* exact HPF coefficients */
constexpr float MIC_ALPHA = expf(-2.0f * 3.14159265f * 10.0f / 20000.0f);
constexpr float ACC_ALPHA = expf(-2.0f * 3.14159265f * 2.0f / 20000.0f);

inline float dc_block(float x, float &y, float &x_prev, float alpha) {
  y = x - x_prev + alpha * y;
  x_prev = x;
  return y;
}

inline int16_t float_to_int16(float v) {
  if (v > 32767.0f) return 32767;
  if (v < -32768.0f) return -32768;
  return (int16_t)v;
}

/* ================= BINARY FRAMES ================= */
struct __attribute__((packed)) MicFrame {
  uint8_t  sync;
  uint32_t timestamp;
  int16_t  mic;
};

struct __attribute__((packed)) AccFrame {
  uint8_t  sync;
  uint32_t timestamp;
  int16_t  acc;
};

struct __attribute__((packed)) TempFrame {
  uint8_t  sync;
  uint32_t timestamp;
  float    temperature;
};

struct __attribute__((packed)) DiagFrame {
  uint8_t  sync;
  uint32_t timestamp;
  uint32_t invalid_channel_count;
  uint32_t adc_read_errors;
  uint32_t serial_overflow_count;
  uint32_t mic_frames_sent;
  uint32_t acc_frames_sent;
};

MicFrame  mic_frame{SYNC_BYTE_MIC, 0, 0};
AccFrame  acc_frame{SYNC_BYTE_ACC, 0, 0};
TempFrame temp_frame{TEMP_SYNC_BYTE, 0, 0.0f};
DiagFrame diag_frame{DIAG_SYNC_BYTE, 0, 0, 0, 0, 0, 0};

/* ================= COUNTERS ================= */
uint32_t invalid_channel_count = 0;
uint32_t adc_read_errors = 0;
uint32_t serial_overflow_count = 0;
uint32_t mic_frames_sent = 0;
uint32_t acc_frames_sent = 0;

/* ================= TIMING ================= */
constexpr uint32_t MIC_PERIOD_US = 1000000UL / 20000UL;
constexpr uint32_t ACC_PERIOD_US = 1000000UL / 5000UL;

uint32_t mic_ts = 0;
uint32_t acc_ts = 0;
uint32_t acc_decim = 0;

/* ================= SERIAL ================= */
inline bool safe_serial_write(const uint8_t *data, size_t len) {
  if (Serial.availableForWrite() < (int)len) {
    serial_overflow_count++;
    return false;
  }
  Serial.write(data, len);
  return true;
}

/* ================= SETUP ================= */
void setup() {
  Serial.begin(BAUD_RATE);
  Serial.setTxBufferSize(4096);
  delay(300);

  tempSensor.begin();
  if (tempSensor.getDeviceCount() > 0) {
    tempSensor.setResolution(12);
    tempSensor.setWaitForConversion(false);
    tempSensor.requestTemperatures();
    temp_sensor_available = true;
  }

  adc_continuous_handle_cfg_t handle_cfg = {
    .max_store_buf_size = DMA_BUF_SIZE,
    .conv_frame_size   = FRAME_SIZE,
  };

  if (adc_continuous_new_handle(&handle_cfg, &adc_handle) != ESP_OK) {
    adc_handle = NULL;
    adc_read_errors++;
    return;
  }

  adc_digi_pattern_config_t pattern[2] = {};

  pattern[0].atten = ADC_ATTEN_DB_11;
  pattern[0].channel = MIC_CH;
  pattern[0].unit = ADC_UNIT_1;
  pattern[0].bit_width = ADC_BITWIDTH_12;

  pattern[1].atten = ADC_ATTEN_DB_11;
  pattern[1].channel = ACC_CH;
  pattern[1].unit = ADC_UNIT_1;
  pattern[1].bit_width = ADC_BITWIDTH_12;

  adc_continuous_config_t cfg = {
    .sample_freq_hz = SAMPLE_RATE_HZ,
    .conv_mode      = ADC_CONV_SINGLE_UNIT_1,
    .format         = ADC_DIGI_OUTPUT_FORMAT_TYPE1,
    .pattern_num    = 2,
    .adc_pattern    = pattern,
  };

  if (adc_continuous_config(adc_handle, &cfg) != ESP_OK ||
      adc_continuous_start(adc_handle) != ESP_OK) {
    adc_read_errors++;
    adc_handle = NULL;
    return;
  }

  mic_ts = micros();
  acc_ts = mic_ts;
}

/* ================= LOOP ================= */
void loop() {
  static uint8_t dma_buf[FRAME_SIZE];
  uint32_t out_len = 0;
  unsigned long now_ms = millis();

  if (temp_sensor_available && now_ms - lastTempRead >= TEMP_READ_INTERVAL_MS) {
    lastTempRead = now_ms;
    currentTemperature = tempSensor.getTempCByIndex(0);

    if (currentTemperature > -100.0f && currentTemperature < 125.0f) {
      temp_frame.timestamp = micros();
      temp_frame.temperature = currentTemperature;
      safe_serial_write((uint8_t *)&temp_frame, sizeof(temp_frame));
    }
    tempSensor.requestTemperatures();
  }

  if (now_ms - lastDiagSend >= DIAG_INTERVAL_MS) {
    lastDiagSend = now_ms;
    diag_frame.timestamp = micros();
    diag_frame.invalid_channel_count = invalid_channel_count;
    diag_frame.adc_read_errors = adc_read_errors;
    diag_frame.serial_overflow_count = serial_overflow_count;
    diag_frame.mic_frames_sent = mic_frames_sent;
    diag_frame.acc_frames_sent = acc_frames_sent;
    safe_serial_write((uint8_t *)&diag_frame, sizeof(diag_frame));
  }

  if (!adc_handle) return;

  esp_err_t ret = adc_continuous_read(adc_handle, dma_buf, FRAME_SIZE, &out_len, 1000);
  if (ret != ESP_OK || out_len == 0) return;

  for (uint32_t i = 0; i < out_len; i += sizeof(adc_digi_output_data_t)) {
    adc_digi_output_data_t *d = (adc_digi_output_data_t *)&dma_buf[i];
    uint8_t ch = d->type1.channel;
    uint16_t raw = d->type1.data & 0x0FFF;
    float v = (float)raw;

    if (ch == MIC_CH) {
      float y = dc_block(v, mic_y, mic_x_prev, MIC_ALPHA);
      mic_frame.timestamp = mic_ts;
      mic_frame.mic = float_to_int16(y);
      if (safe_serial_write((uint8_t *)&mic_frame, sizeof(mic_frame))) {
        mic_frames_sent++;
      }
      mic_ts += MIC_PERIOD_US;
    }
    else if (ch == ACC_CH) {
      acc_decim++;
      if ((acc_decim & 0x03) == 0) {
        float y = dc_block(v, acc_y, acc_x_prev, ACC_ALPHA);
        acc_frame.timestamp = acc_ts;
        acc_frame.acc = float_to_int16(y);
        if (safe_serial_write((uint8_t *)&acc_frame, sizeof(acc_frame))) {
          acc_frames_sent++;
        }
        acc_ts += ACC_PERIOD_US;
      }
    }
    else {
      invalid_channel_count++;
    }
  }
}

