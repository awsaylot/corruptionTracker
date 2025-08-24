@echo off
cd /d "C:\Users\Celot\Documents\Projects\corruptionTracker\clankClient"
echo Building to see all errors...
go build .\cmd\clankClient\main.go 2>&1
