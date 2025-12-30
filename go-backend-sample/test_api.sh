#!/bin/bash

echo "=== Testing API Authentication ==="

echo "1. Logging in..."
curl -s -X POST http://localhost:3001/api/auth/login \
  -H "Content-Type: application/json" \
  -d @test_login.json > login_resp.json

cat login_resp.json | jq .

TOKEN=$(cat login_resp.json | jq -r '.access_token')
echo ""
echo "2. Token extracted (length: ${#TOKEN})"
echo "Token: $TOKEN"

echo ""
echo "3. Testing /api/approvals/tasks endpoint..."
curl -s -H "Authorization: Bearer $TOKEN" \
  http://localhost:3001/api/approvals/tasks | jq .

echo ""
echo "4. Server logs:"
tail -10 server.log
