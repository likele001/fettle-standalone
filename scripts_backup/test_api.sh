#!/bin/bash

echo "=== 登录拿 token ==="
RESP=$(curl -s -X POST 'http://localhost:9100/api/v1/admin/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}')
echo "Login response: $(echo $RESP | head -c 200)"

TOKEN=$(echo $RESP | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["data"]["tokens"]["access_token"])' 2>/dev/null)
if [ -z "$TOKEN" ]; then
  echo "获取 token 失败"
  exit 1
fi
echo "Token: ${TOKEN:0:30}..."
echo ""

echo "=== revenue/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/revenue/stats' \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "=== monitor/stats ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/monitor/stats' \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "=== tenants?page_size=1 ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/tenants?page_size=1' \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "=== tenants/?page_size=1 ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/admin/tenants/?page_size=1' \
  -H "Authorization: Bearer $TOKEN"
echo ""

echo "=== auth/refresh ==="
REFRESH=$(echo $RESP | python3 -c 'import sys,json; d=json.load(sys.stdin); print(d["data"]["tokens"]["refresh_token"])' 2>/dev/null)
curl -s -w '\nHTTP %{http_code}\n' -X POST 'http://localhost:9100/api/v1/admin/auth/refresh' \
  -H 'Content-Type: application/json' \
  -d "{\"refresh_token\":\"$REFRESH\"}"
