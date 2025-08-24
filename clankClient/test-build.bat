@echo off
echo Testing clankClient build...

cd /d %~dp0

echo Current directory: %CD%

echo.
echo Running go mod tidy...
go mod tidy

echo.
echo Running go build...
go build -o bin\clankClient.exe .\cmd\clankClient

if errorlevel 1 (
    echo Build failed!
    exit /b 1
) else (
    echo Build successful!
)

echo.
echo Running tests...
go test ./...

echo.
echo Done!
