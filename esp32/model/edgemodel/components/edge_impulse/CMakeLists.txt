# Get all source files, including .cc files used by TensorFlow
file(GLOB_RECURSE SOURCES 
    "edge-impulse-sdk/*.cpp" 
    "edge-impulse-sdk/*.c"
    "edge-impulse-sdk/*.cc"
    "model-parameters/*.cpp"
    "tflite-model/*.cpp"
)

# Register the component
idf_component_register(
    SRCS ${SOURCES}
    INCLUDE_DIRS 
        "."
        "edge-impulse-sdk"
        "model-parameters"
        "tflite-model"
        "edge-impulse-sdk/classifier"
        "edge-impulse-sdk/dsp"
        "edge-impulse-sdk/porting/espressif/esp-dsp/modules/common/include"
        "edge-impulse-sdk/porting/espressif/esp-dsp/modules/fft/include"
        "edge-impulse-sdk/dsp/kissfft"
        "edge-impulse-sdk/porting"
        "edge-impulse-sdk/tensorflow"
        "edge-impulse-sdk/third_party/flatbuffers/include"
        "edge-impulse-sdk/third_party/gemmlowp"
        "edge-impulse-sdk/third_party/ruy"
    PRIV_REQUIRES esp_timer esp-dsp esp_driver_gpio
)

# Crucial for TFLM on ESP32
target_compile_definitions(${COMPONENT_LIB} PRIVATE TF_LITE_STATIC_MEMORY)

# Disable warnings
target_compile_options(${COMPONENT_LIB} PRIVATE -w)
