@echo off
mkdir build 2>nul

REM Build for Windows
SET GOOS=windows
SET GOARCH=amd64
go build -o build/ced.exe main.go

REM Build for Linux
SET GOOS=linux
SET GOARCH=amd64
go build -o build/ced-linux main.go

REM Build for macOS
SET GOOS=darwin
SET GOARCH=amd64
go build -o build/ced-macos main.go

echo [+] build completed successfully!