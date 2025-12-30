#!/bin/bash

TOKEN=$(cat login_resp.json | jq -r '.access_token')

echo "=== Testing New API Endpoints ==="
echo ""

# Analytics Endpoints
echo "📊 ANALYTICS ENDPOINTS"
echo "====================="
echo ""

echo "1. Testing GET /api/analytics/metrics..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/analytics/metrics | jq .

echo ""
echo "2. Testing GET /api/analytics/trends?days=7..."
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3001/api/analytics/trends?days=7" | jq .

echo ""
echo "3. Testing GET /api/analytics/bottlenecks..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/analytics/bottlenecks | jq .

# Notification Endpoints
echo ""
echo ""
echo "🔔 NOTIFICATION ENDPOINTS"
echo "========================"
echo ""

echo "4. Testing GET /api/notifications..."
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3001/api/notifications?limit=10" | jq .

echo ""
echo "5. Testing GET /api/notifications/unread..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/notifications/unread | jq .

echo ""
echo "6. Testing GET /api/notifications/unread/count..."
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:3001/api/notifications/unread/count | jq .

# Audit Log Endpoints
echo ""
echo ""
echo "📝 AUDIT LOG ENDPOINTS"
echo "====================="
echo ""

echo "7. Testing GET /api/audit-logs/my..."
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3001/api/audit-logs/my?limit=10" | jq .

echo ""
echo "8. Testing GET /api/audit-logs..."
curl -s -H "Authorization: Bearer $TOKEN" "http://localhost:3001/api/audit-logs?limit=10" | jq .

echo ""
echo ""
echo "✅ All new endpoints tested!"
echo ""
echo "Summary:"
echo "  ✓ 3 Analytics endpoints"
echo "  ✓ 3 Notification endpoints"
echo "  ✓ 2 Audit Log endpoints"
echo "  = 8 new endpoints total"
