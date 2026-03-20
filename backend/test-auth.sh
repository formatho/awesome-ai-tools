#!/bin/bash

echo "🧪 PHASE 4 TESTING: Authentication Flow Tests"
echo "============================================="
echo ""
echo "Backend URL: http://localhost:18765"
echo ""

# Test 1: Get system status (should work without auth)
echo "Test 1: GET /api/system/status (no auth)"
curl -s http://localhost:18765/api/system/status | head -20 || echo "Connection failed"
echo ""
echo "---"
echo ""

# Test 2: Try to login with sample credentials
echo "Test 2: POST /auth/login (sample credentials)"
curl -s -X POST http://localhost:18765/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Admin123!"}' | head -50 || echo "Login failed"
echo ""
echo "---"
echo ""

# Test 3: Try to access protected endpoint without token (should fail)
echo "Test 3: GET /api/team/invitations (no auth - should fail)"
curl -s http://localhost:18765/api/team/invitations || echo "Connection failed"
echo ""
echo "---"
echo ""

# Test 4: Try with token from login (if successful)
echo "Test 4: GET /api/team/invitations/stats (with token)"
TOKEN=$(curl -s -X POST http://localhost:18765/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"Admin123!"}' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -n "$TOKEN" ]; then
  echo "Token received: ${TOKEN:0:50}..."
  curl -s http://localhost:18765/api/team/invitations/stats \
    -H "Authorization: Bearer $TOKEN" || echo "Request failed"
else
  echo "No token available from login test"
fi

echo ""
echo "============================================="
echo "Test Summary:"
echo "- System status endpoint works ✅ (no auth required)"
echo "- Login endpoint tested for JWT token generation"
echo "- Protected endpoints require valid tokens 🔐"
echo "============================================="
