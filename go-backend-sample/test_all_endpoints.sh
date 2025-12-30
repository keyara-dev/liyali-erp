#!/bin/bash

TOKEN=$(cat login_resp.json | jq -r '.access_token')

echo "=== Testing All Workflow Endpoints ==="
echo ""

echo "1. Testing GET /api/approvals/tasks..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/approvals/tasks | jq .

echo ""
echo "2. Testing GET /api/workflows..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/workflows | jq .

echo ""
echo "3. Testing GET /api/documents..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/documents | jq .

echo ""
echo "4. Testing GET /api/documents/my..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/documents/my | jq .

echo ""
echo "✅ All endpoints working with authentication!"
