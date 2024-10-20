#!/bin/bash

build_exe() {
  local build_version=$1
  local build_authors=$2
  local build_edition=$3
  local build_exe_os=$4

  local build_exe_arch=$5
  local build_exe_cc=$6
  local build_exe_cxx=$7
  local build_exe_output=$8
  local build_go_args=$9
  echo "go_args: $build_go_args"
  echo "build_data: $(date -u +%Y-%m-%dT%H:%M:%SZ)"

  CGO_ENABLED=1 CC=$build_exe_cc CXX=$build_exe_cxx GOOS=$build_exe_os GOARCH=$build_exe_arch go build $build_go_args -buildmode=pie -trimpath -ldflags "-H=windowsgui -X 'fadacontrol/internal/base/version.AuthorEmail=$build_authors' -X 'fadacontrol/internal/base/version._VersionName=$build_version' -X 'fadacontrol/internal/base/version.GitCommit=$(git log --pretty=format:'%h' -1)' -X 'fadacontrol/internal/base/version.Edition=$build_edition' -X 'fadacontrol/internal/base/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)' -s -w -linkmode=external -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o $build_exe_output
}
if [ "$#" -ne 3 ]; then
  echo "Usage: $0 <build_version> <author_email>  <build_os>"
  exit 1
fi
build_version=$1
author_email=$2
build_os=$3
last_digit=${build_version: -1}
build_edition="nightly"
go_args=""
case $last_digit in
    0)
        build_edition="release"
        ;;
    3)
        build_edition="beta"
        ;;
    5)
        build_edition="dev"
        ;;
    7)
        build_edition="canary"
        ;;
    9)

        build_edition="nightly"
        go_args='-tags=swag'
        ;;
    *)
        build_edition="nightly"
        ;;
esac

# x64
build_exe $build_version $author_email $build_edition $build_os  amd64 x86_64-w64-mingw32-gcc x86_64-w64-mingw32-g++ ./out/core-service-x64.exe $go_args

# arm64
build_exe $build_version $author_email  $build_edition $build_os  arm64 aarch64-w64-mingw32-gcc aarch64-w64-mingw32-g++ ./out/core-service-arm64.exe $go_args

# x86
build_exe $build_version $author_email  $build_edition $build_os  386 i686-w64-mingw32-gcc i686-w64-mingw32-g++ ./out/core-service-x86.exe $go_args