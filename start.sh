#!/bin/bash

# Agent Orchestrator Startup Script
# Usage: ./start.sh [backend|frontend|all|build]

set -e

# Get the script's directory and resolve to absolute path
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$SCRIPT_DIR"
BACKEND_DIR="$PROJECT_ROOT/backend"
FRONTEND_DIR="$PROJECT_ROOT/electron-app"

build_backend() {
    echo "📦 Building Backend (all platforms)..."
    cd "$BACKEND_DIR"

    # Detect current platform
    CURRENT_OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    CURRENT_ARCH=$(uname -m)
    if [ "$CURRENT_ARCH" = "arm64" ]; then
        CURRENT_ARCH="arm64"
    else
        CURRENT_ARCH="amd64"
    fi

    # Build for current platform (fast, for development)
    echo "  Building for $CURRENT_OS-$CURRENT_ARCH..."
    go build -o bin/server ./cmd/server

    # Build all platforms for Electron packaging
    echo "  Building for darwin-arm64..."
    GOOS=darwin GOARCH=arm64 go build -o bin/agent-orchestrator-server-darwin-arm64 ./cmd/server

    echo "  Building for darwin-amd64..."
    GOOS=darwin GOARCH=amd64 go build -o bin/agent-orchestrator-server-darwin-amd64 ./cmd/server

    echo "  Building for linux-arm64..."
    GOOS=linux GOARCH=arm64 go build -o bin/agent-orchestrator-server-linux-arm64 ./cmd/server

    echo "  Building for linux-amd64..."
    GOOS=linux GOARCH=amd64 go build -o bin/agent-orchestrator-server-linux-amd64 ./cmd/server

    # Update packaged Electron apps if they exist
    if [ -d "$FRONTEND_DIR/release/mac-arm64" ]; then
        echo "  Updating packaged mac-arm64 app..."
        cp bin/agent-orchestrator-server-darwin-arm64 "$FRONTEND_DIR/release/mac-arm64/Agent Orchestrator.app/Contents/Resources/backend/server"
    fi
    if [ -d "$FRONTEND_DIR/release/mac" ]; then
        echo "  Updating packaged mac-amd64 app..."
        cp bin/agent-orchestrator-server-darwin-amd64 "$FRONTEND_DIR/release/mac/Agent Orchestrator.app/Contents/Resources/backend/server"
    fi

    echo "✅ Backend build complete"
}

build_frontend() {
    echo "📦 Building Frontend..."
    cd "$FRONTEND_DIR"

    # Install dependencies if needed
    if [ ! -d "node_modules" ]; then
        echo "  Installing dependencies..."
        npm install
    fi

    echo "  Building React app..."
    npm run build

    echo "✅ Frontend build complete"
}

build_all() {
    echo "🔨 Building Everything..."
    echo ""
    build_backend
    echo ""
    build_frontend
    echo ""
    echo "✅ All builds complete!"
}

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
    build)
        build_all
        ;;
    all)
        start_all
        ;;
    *)
        echo "Usage: $0 [backend|frontend|all|build]"
        echo ""
        echo "  backend  - Start only the Go API server"
        echo "  frontend - Start only the Electron app"
        echo "  all      - Start both (default)"
        echo "  build    - Build backend (all platforms) and frontend"
        exit 1
        ;;
esac
