#!/bin/bash
export PATH=$PATH:/usr/local/go/bin

echo "=== 编译 billing-service ==="
cd /www/wwwroot/fettle/backend/billing-service
go build -o /www/wwwroot/fettle/build/billing-service . 2>&1
if [ $? -ne 0 ]; then echo "编译失败"; exit 1; fi
echo "编译成功"

echo ""
echo "=== 重启 billing-service ==="
pkill billing-service || true
sleep 1
set -a
source /www/wwwroot/fettle/.env
set +a
nohup /www/wwwroot/fettle/build/billing-service > /www/wwwroot/fettle/logs/billing.log 2>&1 &
sleep 3

echo "=== 进程 ==="
ps aux | grep billing-service | grep -v grep

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
