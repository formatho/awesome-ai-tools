#!/bin/bash

# Agent Orchestrator - Staging Environment Startup Script
# Created: March 24, 2026
# Task ID: 5d3a5376-c365-4d8a-97ca-df63062003bf

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🏗️  Agent Orchestrator - Staging Environment${NC}"
echo "================================================"
echo ""

# Configuration
STAGING_PORT=8080
STAGING_DB="agent_orchestrator_staging.db"
ENV_FILE=".env.staging"
BACKEND_DIR="/Users/studio/sandbox/formatho/agent-orchestrator/backend"

# Check if .env.staging exists
if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}Error: $ENV_FILE not found!${NC}"
    exit 1
fi

# Navigate to backend directory
cd "$BACKEND_DIR"

# Check if staging database exists
if [ -f "$STAGING_DB" ]; then
    echo -e "${YELLOW}⚠️  Staging database exists: $STAGING_DB${NC}"
    read -p "Do you want to reset it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Deleting existing staging database..."
        rm -f "$STAGING_DB"
        echo -e "${GREEN}✓ Database reset${NC}"
    fi
else
    echo -e "${GREEN}✓ Fresh staging database will be created${NC}"
fi

# Check if staging server is already running
if lsof -Pi :$STAGING_PORT -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo -e "${RED}Error: Port $STAGING_PORT is already in use!${NC}"
    echo "Running processes on port $STAGING_PORT:"
    lsof -Pi :$STAGING_PORT -sTCP:LISTEN
    echo ""
    read -p "Do you want to kill the existing process? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "Killing process on port $STAGING_PORT..."
        lsof -ti:$STAGING_PORT | xargs kill -9 2>/dev/null || true
        sleep 2
    else
        echo "Aborting..."
        exit 1
    fi
fi

# Check if development server is running (port 18765)
if lsof -Pi :18765 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo -e "${GREEN}✓ Development server detected on port 18765${NC}"
    echo "  Staging will run on port $STAGING_PORT (no conflict)"
fi

echo ""
echo "Starting staging environment..."
echo "  Port: $STAGING_PORT"
echo "  Database: $STAGING_DB"
echo "  Config: $ENV_FILE"
echo ""

# Export environment variables from .env.staging
set -a
source "../$ENV_FILE"
set +a

# Override with staging-specific settings
export PORT=$STAGING_PORT
export DATABASE_URL="file:$STAGING_DB"
export ENV="staging"

echo -e "${GREEN}Environment variables loaded:${NC}"
echo "  ENV=$ENV"
echo "  PORT=$PORT"
echo "  DATABASE_URL=$DATABASE_URL"
echo "  LOG_LEVEL=$LOG_LEVEL"
echo ""

# Check if server binary exists
if [ ! -f "./server" ]; then
    echo -e "${YELLOW}⚠️  Server binary not found. Building...${NC}"
    go build -o server ./cmd/server
    echo -e "${GREEN}✓ Server binary built${NC}"
fi

# Start the server
echo -e "${GREEN}🚀 Starting staging server...${NC}"
echo ""

# Run the server
./server

# Note: The script will stay running while the server is active
# To stop: Press Ctrl+C
