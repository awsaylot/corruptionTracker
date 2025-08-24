@echo off
echo 🚀 Building and starting ClankClient Server...
echo.

echo Building...
go build -o clankClient.exe ./cmd/clankClient

if %ERRORLEVEL% neq 0 (
    echo ❌ Build failed!
    pause
    exit /b 1
)

echo ✅ Build successful!
echo.
echo Starting server on port 8081...
echo Press Ctrl+C to stop the server
echo.

clankClient.exe serve
