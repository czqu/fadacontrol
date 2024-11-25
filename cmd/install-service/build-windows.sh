#!/bin/bash
go install github.com/akavel/rsrc@v0.10.2
rsrc  -arch amd64 -manifest ./manifest.xml  -o app.syso
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -buildmode=pie -trimpath -ldflags "-s -w -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o ./out/install-service-x64.exe
rsrc  -arch 386 -manifest ./manifest.xml   -o app.syso
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -buildmode=pie -trimpath -ldflags "-s -w  -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o ./out/install-service-x86.exe
rsrc  -arch arm64 -manifest ./manifest.xml   -o app.syso
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -buildmode=pie -trimpath -ldflags "-s -w  -extldflags '-static -flto -O2 -Wl,--gc-sections'" -o ./out/install-service-arm64.exe