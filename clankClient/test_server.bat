@echo off
REM Test script for ClankClient Server
set SERVER_URL=http://localhost:8081

echo 🚀 Testing ClankClient Server at %SERVER_URL%
echo.

echo 1. Testing health endpoint...
curl -s "%SERVER_URL%/health"
echo.

echo 2. Testing info endpoint...
curl -s "%SERVER_URL%/api/info"
echo.

echo 3. Testing list tools...
curl -s "%SERVER_URL%/api/tools"
echo.

echo 4. Testing list resources...
curl -s "%SERVER_URL%/api/resources"
echo.

echo 5. Testing list prompts...
curl -s "%SERVER_URL%/api/prompts"
echo.

echo 6. Testing tool call (system info)...
curl -s -X POST "%SERVER_URL%/api/tools/call" -H "Content-Type: application/json" -d "{\"tool_name\": \"clank:system_info\", \"arguments\": {\"type\": \"cwd\"}}"
echo.

echo 7. Testing tool call (file operations)...
curl -s -X POST "%SERVER_URL%/api/tools/call" -H "Content-Type: application/json" -d "{\"tool_name\": \"clank:file_operations\", \"arguments\": {\"operation\": \"list\", \"path\": \".\"}}"
echo.

echo ✅ Testing completed!
pause
