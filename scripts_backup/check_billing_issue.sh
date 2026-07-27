#!/bin/bash

echo "=== Nginx fettle.cenkor.cn 配置 ==="
cat /www/server/panel/vhost/nginx/fettle.cenkor.cn.conf 2>/dev/null | grep -A5 -B2 "billing\|api\|proxy_pass"

echo ""
echo "=== 测试 subscriptions ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/billing/subscriptions' \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "=== 测试 quota ==="
curl -s -w '\nHTTP %{http_code}\n' 'http://localhost:9100/api/v1/billing/quota' \
  -H "Authorization: Bearer $TOKEN"

echo ""
echo "=== billing-service 日志 ==="
tail -10 /www/wwwroot/fettle/logs/billing.log 2>/dev/null
