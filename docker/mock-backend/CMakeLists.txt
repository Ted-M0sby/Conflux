cmake_minimum_required(VERSION 3.21)
project(mock_backend LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

add_executable(mock-backend main.cpp)
target_include_directories(mock-backend PRIVATE ../../third_party)
