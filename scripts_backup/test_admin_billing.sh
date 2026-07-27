#!/bin/bash
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin

TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "=== GET /billing/plans ==="
curl -s "http://localhost:9100/api/v1/billing/plans" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== GET /billing/subscriptions ==="
curl -s "http://localhost:9100/api/v1/billing/subscriptions" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== GET /billing/quota ==="
curl -s "http://localhost:9100/api/v1/billing/quota" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== GET /billing/records ==="
curl -s "http://localhost:9100/api/v1/billing/records?page=1&page_size=10" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null
