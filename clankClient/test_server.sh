#!/bin/bash

# Test script for ClankClient Server
SERVER_URL="http://localhost:8081"

echo "🚀 Testing ClankClient Server at $SERVER_URL"
echo

# Test health endpoint
echo "1. Testing health endpoint..."
curl -s "$SERVER_URL/health" | jq . || echo "Health check failed"
echo

# Test info endpoint
echo "2. Testing info endpoint..."
curl -s "$SERVER_URL/api/info" | jq . || echo "Info failed"
echo

# Test list tools
echo "3. Testing list tools..."
curl -s "$SERVER_URL/api/tools" | jq . || echo "List tools failed"
echo

# Test list resources
echo "4. Testing list resources..."
curl -s "$SERVER_URL/api/resources" | jq . || echo "List resources failed"
echo

# Test list prompts
echo "5. Testing list prompts..."
curl -s "$SERVER_URL/api/prompts" | jq . || echo "List prompts failed"
echo

# Test tool call - system info
echo "6. Testing tool call (system info)..."
curl -s -X POST "$SERVER_URL/api/tools/call" \
  -H "Content-Type: application/json" \
  -d '{"tool_name": "clank:system_info", "arguments": {"type": "cwd"}}' | jq . || echo "Tool call failed"
echo

# Test tool call - file operations
echo "7. Testing tool call (file operations)..."
curl -s -X POST "$SERVER_URL/api/tools/call" \
  -H "Content-Type: application/json" \
  -d '{"tool_name": "clank:file_operations", "arguments": {"operation": "list", "path": "."}}' | jq . || echo "File operations failed"
echo

echo "✅ Testing completed!"
