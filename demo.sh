#!/bin/bash

# Drifter Platform RPG Demo Script
# This script demonstrates the basic functionality of the RPG platform

echo "🎮 Drifter Platform RPG Demo"
echo "=============================="
echo ""

# Check if binaries exist
if [ ! -f "./build/drifter-server" ]; then
    echo "❌ Server binary not found. Building..."
    make build-server
fi

if [ ! -f "./build/drifter-client" ]; then
    echo "❌ Client binary not found. Building..."
    make build-client
fi

echo "✅ Binaries ready"
echo ""

# Start server in background
echo "🚀 Starting server on port 8080..."
./build/drifter-server &
SERVER_PID=$!

# Wait for server to start
sleep 2

echo "✅ Server started (PID: $SERVER_PID)"
echo ""

# Test server health
echo "🔍 Testing server health..."
if curl -s http://localhost:8080/ws > /dev/null; then
    echo "✅ Server is responding"
else
    echo "❌ Server is not responding"
    kill $SERVER_PID
    exit 1
fi

echo ""
echo "📋 Available commands:"
echo "  - /help     - Show help"
echo "  - /move forward - Move character forward"
echo "  - /attack enemy - Attack an enemy"
echo "  - /chat Hello! - Send chat message"
echo "  - /quit     - Exit client"
echo ""
echo "🎯 Starting client (type /quit to exit)..."
echo ""

# Start client
./build/drifter-client --server ws://localhost:8080/ws

# Cleanup
echo ""
echo "🧹 Cleaning up..."
kill $SERVER_PID
echo "✅ Demo completed"
