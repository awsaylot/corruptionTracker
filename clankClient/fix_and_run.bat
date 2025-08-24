@echo off
cd /d "C:\Users\Celot\Documents\Projects\corruptionTracker\clankClient"
echo Running go mod tidy...
go mod tidy
echo.
echo Attempting to build...
go run .\cmd\clankClient\main.go
