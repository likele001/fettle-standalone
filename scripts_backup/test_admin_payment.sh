#!/bin/bash
export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin

TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "=== 测试 /admin/payment/config ==="
curl -s "http://localhost:9100/api/v1/admin/payment/config" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== 测试 /admin/payment/orders ==="
curl -s "http://localhost:9100/api/v1/admin/payment/orders?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null
