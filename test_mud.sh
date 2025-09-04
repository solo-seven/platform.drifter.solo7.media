#!/bin/bash

# Test script for MUD client and server

echo "Starting MUD Test..."

# Build the project
echo "Building project..."
go build -o bin/server ./cmd/server
go build -o bin/client ./cmd/client

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"

# Start server in background
echo "Starting server..."
./bin/server -port 8080 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test server is running
if ! curl -s http://localhost:8080/ws > /dev/null; then
    echo "Server failed to start!"
    kill $SERVER_PID
    exit 1
fi

echo "Server started successfully!"

# Start client
echo "Starting client..."
echo "You can now interact with the MUD. Try commands like:"
echo "  look"
echo "  move north"
echo "  get rusty sword"
echo "  inventory"
echo "  say hello"
echo "  quit"
echo ""

# Run client (this will block)
./bin/client --server localhost:8081

# Cleanup
echo "Shutting down server..."
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

echo "Test completed!"
