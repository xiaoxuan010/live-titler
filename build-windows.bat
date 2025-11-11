@echo off
echo Building live-titler for Windows...
set GOOS=windows
set GOARCH=amd64
go build -o live-titler.exe main.go
echo Build complete: live-titler.exe
