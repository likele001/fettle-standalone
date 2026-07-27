#!/bin/bash

echo "=== 测试登录 ==="
LOGIN_RESP=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}')
echo "$LOGIN_RESP" | python3 -m json.tool 2>/dev/null

TOKEN=$(echo "$LOGIN_RESP" | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)
echo ""
echo "Token: ${TOKEN:0:30}..."

echo ""
echo "=== 测试 dashboard/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/dashboard/stats' \
  -H "Authorization: Bearer $TOKEN" | head -5

echo ""
echo "=== 测试 tenants ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/tenants' \
  -H "Authorization: Bearer $TOKEN" | head -3

echo ""
echo "=== 测试 billing/plans ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN" | head -3

echo ""
echo "=== 测试 revenue/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/revenue/stats' \
  -H "Authorization: Bearer $TOKEN" | head -3

echo ""
echo "=== 测试 monitor/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/monitor/stats' \
  -H "Authorization: Bearer $TOKEN" | head -3

echo ""
echo "=== gateway 日志 ==="
tail -20 /www/wwwroot/fettle/logs/gateway.log
