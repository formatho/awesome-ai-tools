#!/bin/bash

# Agent Orchestrator Startup Script
# Usage: ./start.sh [backend|frontend|all]

set -e

PROJECT_ROOT="/Users/studio/sandbox/formatho/agent-orchestrator"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/electron-app"

start_backend() {
    echo "🔧 Starting Backend API Server..."
    cd "$BACKEND_DIR"

    # Build if binary doesn't exist or source changed
    if [ ! -f bin/server ] || [ "$(find . -name '*.go' -newer bin/server | wc -l)" -gt 0 ]; then
        echo "📦 Building backend..."
        go build -o bin/server ./cmd/server
    fi

    echo "🚀 Backend running on http://localhost:18765"
    echo "📊 API: http://localhost:18765/api"
    echo "🔌 WebSocket: ws://localhost:18765/ws"
    echo ""

    ./bin/server
}

start_frontend() {
    echo "⚛️  Starting Electron App..."
    cd "$FRONTEND_DIR"

    # Check if node_modules exists
    if [ ! -d "node_modules" ]; then
        echo "📦 Installing dependencies..."
        npm install
    fi

    echo "🚀 Starting Electron..."
    npm run dev
}

start_all() {
    echo "🚀 Starting Agent Orchestrator (Full Stack)..."
    echo ""

    # Start backend in background
    echo "🔧 Starting Backend..."
    cd "$BACKEND_DIR"
    if [ ! -f bin/server ] || [ "$(find . -name '*.go' -newer bin/server | wc -l)" -gt 0 ]; then
        go build -o bin/server ./cmd/server
    fi
    ./bin/server &
    BACKEND_PID=$!
    echo "✅ Backend started (PID: $BACKEND_PID)"
    sleep 2

    # Start frontend
    echo ""
    echo "⚛️  Starting Electron..."
    cd "$FRONTEND_DIR"
    if [ ! -d "node_modules" ]; then
        npm install
    fi
    npm run dev

    # Cleanup on exit
    trap "kill $BACKEND_PID 2>/dev/null" EXIT
}

# Main logic
case "${1:-all}" in
    backend)
        start_backend
        ;;
    frontend)
        start_frontend
        ;;
    all)
        start_all
        ;;
    *)
        echo "Usage: $0 [backend|frontend|all]"
        echo ""
        echo "  backend  - Start only the Go API server"
        echo "  frontend - Start only the Electron app"
        echo "  all      - Start both (default)"
        exit 1
        ;;
esac
