cmake_minimum_required(VERSION 3.22)

include($ENV{IDF_PATH}/tools/cmake/project.cmake)
project(edgemodel)
add_compile_options("-D_GLIBCXX_USE_C99_MATH_TR1")
add_compile_options("-include" "cmath")
# Do not add target_compile_options here
