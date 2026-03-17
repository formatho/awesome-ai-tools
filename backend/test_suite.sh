#!/bin/bash

# Test Suite Runner for Agent Orchestrator
# Runs all agent-related tests

set -e

echo "======================================"
echo "Agent Orchestrator Test Suite"
echo "======================================"
echo ""

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TOTAL=0
PASSED=0
FAILED=0

# Function to run a test
run_test() {
    local test_name=$1
    local test_command=$2

    echo -e "${YELLOW}Running: $test_name${NC}"
    TOTAL=$((TOTAL + 1))

    if eval "$test_command"; then
        echo -e "${GREEN}✓ PASSED: $test_name${NC}"
        PASSED=$((PASSED + 1))
    else
        echo -e "${RED}✗ FAILED: $test_name${NC}"
        FAILED=$((FAILED + 1))
    fi
    echo ""
}

# Change to backend directory
cd "$(dirname "$0")/backend" || exit 1

echo "======================================"
echo "1. Service Layer Tests"
echo "======================================"
echo ""

run_test "Agent Service Tests" \
    "go test -v ./internal/services -run TestAgentService"

run_test "Agent LLM Integration Tests" \
    "go test -v ./internal/services -run TestAgentLLMIntegration"

echo "======================================"
echo "2. Handler Tests"
echo "======================================"
echo ""

run_test "Agent Handler Tests" \
    "go test -v ./internal/api/handlers -run TestAgentHandler"

echo "======================================"
echo "Test Summary"
echo "======================================"
echo "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}All tests passed! 🎉${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed. Please review the output above.${NC}"
    exit 1
fi
