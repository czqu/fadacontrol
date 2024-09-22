#!/bin/sh

build_exe() {
  local build_version=$1
  local build_authors=$2
  local build_edition=$3
  local build_exe_os=$4
  local build_exe_arch=$5
  local build_exe_cc=$6
  local build_exe_cxx=$7
  local build_exe_output=$8

  CGO_ENABLED=1 CC=$build_exe_cc CXX=$build_exe_cxx GOOS=$build_exe_os GOARCH=$build_exe_arch go build -x -buildmode=pie -trimpath -ldflags "-X 'fadacontrol/internal/base/version.AuthorEmail=$build_authors' -X 'fadacontrol/internal/base/version._VersionName=$build_version' -X 'fadacontrol/internal/base/version.GitCommit=$(git log --pretty=format:'%h' -1)' -X 'fadacontrol/internal/base/version.Edition=$build_edition' -X 'fadacontrol/internal/base/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -s -w -linkmode=external -extldflags '-flto -O2 -Wl,--gc-sections'" -o $build_exe_output
}
if [ "$#" -ne 4 ]; then
  echo "Usage: $0 <author_email> <build_version> <build_edition> <build_os>"
  exit 1
fi
author_email=$1
build_version=$2
build_edition=$3
build_os=$4
# x64
build_exe $build_version $author_email $build_edition $build_os amd64 x86_64-w64-mingw32-gcc x86_64-w64-mingw32-g++ ./out/x64/core-service.exe

# arm64
build_exe $build_version $author_email  $build_edition $build_os arm64 aarch64-w64-mingw32-gcc aarch64-w64-mingw32-g++ ./out/arm64/core-service.exe

# x86
build_exe $build_version $author_email  $build_edition $build_os 386 i686-w64-mingw32-gcc i686-w64-mingw32-g++ ./out/x86/core-service.exe
