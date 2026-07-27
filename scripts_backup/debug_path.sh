#!/bin/bash

echo "=== 直接测 billing-service ==="
echo "--- /plans ---"
curl -s http://localhost:9600/plans
echo ""
echo "--- /billing/plans ---"
curl -s http://localhost:9600/billing/plans
echo ""
echo "--- /api/v1/plans ---"
curl -s http://localhost:9600/api/v1/plans
echo ""

echo "=== 通过 gateway 测 ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "--- /api/v1/billing/plans (via gateway) ---"
curl -sv 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN" 2>&1 | grep -E '< HTTP|^\{|^\['

echo ""
echo "=== gateway 日志 ==="
tail -3 /www/wwwroot/fettle/logs/gateway.log
