#!/bin/bash

echo "=== Testing Users Endpoint ==="

# Kill any existing server processes
pkill -f "go run" 2>/dev/null || true

echo "Starting server..."
cd "$(git rev-parse --show-toplevel)"
go run cmd/main.go &
SERVER_PID=$!

echo "Waiting for server to start..."
sleep 5

echo "Testing GET /api/v1/users endpoint..."
curl -X GET "http://localhost:8100/api/v1/users?limit=10&offset=0" \
     -H "Content-Type: application/json" \
     -w "\nHTTP Status: %{http_code}\n" \
     -s

echo ""
echo "Stopping server..."
kill $SERVER_PID 2>/dev/null || true

echo "Test completed."
