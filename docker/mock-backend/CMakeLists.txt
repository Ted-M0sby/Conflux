cmake_minimum_required(VERSION 3.21)
project(mock_backend LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

include(FetchContent)
FetchContent_Declare(cpp_httplib
  GIT_REPOSITORY https://github.com/yhirose/cpp-httplib.git
  GIT_TAG v0.15.3)
FetchContent_MakeAvailable(cpp_httplib)

add_executable(mock-backend main.cpp)
target_include_directories(mock-backend PRIVATE ${cpp_httplib_SOURCE_DIR})
