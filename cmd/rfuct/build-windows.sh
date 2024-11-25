#!/bin/bash

build_exe() {
  local build_exe_os=$1
  local build_exe_arch=$2
  local build_exe_cc=$3
  local build_exe_cxx=$4
  local build_exe_output=$5
  local build_go_args=$6
  echo "go_args: $build_go_args"
  echo "build_data: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

  CGO_CFLAGS="-O2" CGO_ENABLED=1 CC=$build_exe_cc CXX=$build_exe_cxx GOOS=$build_exe_os GOARCH=$build_exe_arch go build $build_go_args -buildmode=pie -trimpath -ldflags "-s -w -linkmode=external -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o $build_exe_output
}

go_args=""

# x64
build_exe windows   amd64 x86_64-w64-mingw32-gcc  x86_64-w64-mingw32-g++ ./out/rfuct-x64.exe $go_args
cp ./out/rfuct-x64.exe ./out/rfuct-arm64.exe #arm64 is not ready yet, so copy x64 for now
# arm64
#build_exe  windows arm64 aarch64-w64-mingw32-gcc-15.0.0 aarch64-w64-mingw32-g++ ./out/core-service-arm64.exe $go_args

# x86
build_exe  windows  386 i686-w64-mingw32-gcc i686-w64-mingw32-g++ ./out/rfuct-x86.exe $go_args