@echo off

REM Build script for clankClient
echo Building clankClient...

REM Clean previous builds
if exist bin rmdir /s /q bin

REM Create bin directory
mkdir bin

REM Build for current platform
echo Building for current platform...
go build -o bin\clankClient.exe .\cmd\clankClient

REM Build for multiple platforms if requested
if "%1"=="all" (
    echo Building for multiple platforms...
    
    REM Windows
    set GOOS=windows
    set GOARCH=amd64
    go build -o bin\clankClient-windows-amd64.exe .\cmd\clankClient
    
    REM macOS
    set GOOS=darwin
    set GOARCH=amd64
    go build -o bin\clankClient-darwin-amd64 .\cmd\clankClient
    set GOARCH=arm64
    go build -o bin\clankClient-darwin-arm64 .\cmd\clankClient
    
    REM Linux
    set GOOS=linux
    set GOARCH=amd64
    go build -o bin\clankClient-linux-amd64 .\cmd\clankClient
    set GOARCH=arm64
    go build -o bin\clankClient-linux-arm64 .\cmd\clankClient
    
    echo Cross-platform builds completed!
)

echo Build completed! Binaries are in the 'bin' directory.
pause
