#!/bin/bash

echo "=== 登录拿 token ==="
RESP=$(curl -s -X POST 'http://localhost:9100/api/v1/admin/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}')
echo "Login: $(echo $RESP | head -c 100)"

TOKEN=$(echo $RESP | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)
echo ""
echo "=== Token payload (base64 decode) ==="
echo $TOKEN | cut -d. -f2 | base64 -d 2>/dev/null
echo ""
echo ""

echo "=== 用 token 测 admin/tenants ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/tenants?page_size=1' \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "=== 用 token 测 admin/revenue/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/revenue/stats' \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "=== 直接测 user-service 的 admin/tenants (绕过 gateway) ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9200/admin/tenants?page_size=1' \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "=== gateway 日志 ==="
tail -5 /www/wwwroot/fettle/logs/gateway.log
