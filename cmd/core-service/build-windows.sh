#!/bin/sh

build_exe() {
  local build_exe_os=$1
  local build_exe_arch=$2
  local build_exe_cc=$3
  local build_exe_cxx=$4
  local build_exe_output=$5

  CGO_ENABLED=1 CC=$build_exe_cc CXX=$build_exe_cxx GOOS=$build_exe_os GOARCH=$build_exe_arch go build -x -buildmode=pie -trimpath -ldflags "-X 'fadacontrol/internal/base/version.Edition=canary' -X 'fadacontrol/internal/base/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -s -w -linkmode=external -extldflags '-flto -O2 -Wl,--gc-sections'" -o $build_exe_output
}

# x64
build_exe windows amd64 x86_64-w64-mingw32-gcc x86_64-w64-mingw32-g++ ./out/x64/core-service.exe

# arm64
build_exe windows arm64 aarch64-w64-mingw32-gcc aarch64-w64-mingw32-g++ ./out/arm64/core-service.exe

# x86
build_exe windows 386 i686-w64-mingw32-gcc i686-w64-mingw32-g++ ./out/x86/core-service.exe
