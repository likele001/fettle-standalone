#!/bin/bash

TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/admin/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "=== monitor/health 原始返回 ==="
curl -s 'http://localhost:9100/api/v1/admin/monitor/health' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== monitor/stats 原始返回 ==="
curl -s 'http://localhost:9100/api/v1/admin/monitor/stats' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== tenants 原始返回 ==="
curl -s 'http://localhost:9100/api/v1/admin/tenants?page_size=2' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null
