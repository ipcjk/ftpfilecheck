#!/usr/bin/env bash

# X-compile everything ;-)
env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ftpfilecheck *.go
env GOOS=darwin GOARCH=amd64 go  build -ldflags="-s -w" -o ftpfilecheck.mac *.go

# pack our executables
upx --force ftpfilecheck*
