#!/bin/bash

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -buildmode=pie -trimpath -ldflags "-s -w -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o ./out/install-service-x64.exe
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -buildmode=pie -trimpath -ldflags "-s -w  -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o ./out/install-service-x86.exe
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -buildmode=pie -trimpath -ldflags "-s -w  -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o ./out/install-service-arm64.exe