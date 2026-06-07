#!/bin/bash

echo "Testing users endpoint..."

# Start server in background
echo "Starting server..."
go run cmd/main.go &
SERVER_PID=$!

# Wait for server to start
sleep 5

# Test the endpoint
echo "Testing GET /api/v1/users"
curl -X GET "http://localhost:8100/api/v1/users?limit=10&offset=0" \
     -H "Content-Type: application/json" \
     -w "\nHTTP Status: %{http_code}\n"

# Kill server
echo "Stopping server..."
kill $SERVER_PID

echo "Test completed."

