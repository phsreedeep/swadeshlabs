idf_component_register(SRCS "main.cpp"
                       INCLUDE_DIRS "."
                       REQUIRES edge_impulse esp_adc esp_timer esp_driver_uart esp_driver_gpio ds18b20 nvs_flash esp-dsp)

target_compile_options(${COMPONENT_LIB} PRIVATE -w)
