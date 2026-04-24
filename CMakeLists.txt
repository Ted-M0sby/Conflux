cmake_minimum_required(VERSION 3.21)
project(conflux_cpp LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_executable(conflux
  src/main.cpp
  src/config.cpp
  src/router.cpp
)

target_include_directories(conflux PRIVATE
  src
  third_party
)

if (WIN32)
  target_link_libraries(conflux PRIVATE ws2_32)
endif()
