#!/bin/bash

echo "=== 直接测试 billing-service (绕过 gateway) ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "--- 直接调 billing-service :9400 ---"
curl -s -w '\nHTTP %{http_code} time: %{time_total}s\n' 'http://localhost:9400/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "--- 通过 gateway 调 billing ---"
curl -s -w '\nHTTP %{http_code} time: %{time_total}s\n' 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "--- 通过 gateway 调 billing/subscriptions ---"
curl -s -w '\nHTTP %{http_code} time: %{time_total}s\n' 'http://localhost:9100/api/v1/billing/subscriptions' \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "--- 通过 gateway 调 billing/quota ---"
curl -s -w '\nHTTP %{http_code} time: %{time_total}s\n' 'http://localhost:9100/api/v1/billing/quota' \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "--- 通过 gateway 调 billing/records ---"
curl -s -w '\nHTTP %{http_code} time: %{time_total}s\n' 'http://localhost:9100/api/v1/billing/records?page=1&page_size=10' \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "=== gateway 日志 ==="
tail -20 /www/wwwroot/fettle/logs/gateway.log
