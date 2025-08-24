@echo off
cd /d "C:\Users\Celot\Documents\Projects\corruptionTracker\clankClient"
echo Testing build after initial fix...
go build .\cmd\clankClient\main.go
if %errorlevel% equ 0 (
    echo Build successful!
    echo Running the application...
    go run .\cmd\clankClient\main.go
) else (
    echo Build failed with errors above.
)
