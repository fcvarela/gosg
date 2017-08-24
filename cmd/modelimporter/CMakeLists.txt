cmake_minimum_required(VERSION 3.6)
project(modelimporter)

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11")

find_package(PkgConfig REQUIRED)
pkg_search_module(ASSIMP REQUIRED assimp)
pkg_search_module(PROTOBUF REQUIRED protobuf)

set(SOURCE_FILES main.cpp model.pb.cc)

include_directories(${ASSIMP_INCLUDE_DIRS})
include_directories(${PROTOBUF_INCLUDE_DIRS})

link_directories(${PROTOBUF_LIBRARY_DIRS})
link_directories(${ASSIMP_LIBRARY_DIRS})

add_executable(modelimporter ${SOURCE_FILES})
target_link_libraries(modelimporter ${ASSIMP_LIBRARIES} ${PROTOBUF_LIBRARIES})