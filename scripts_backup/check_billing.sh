#!/bin/bash

echo "=== billing-service 进程 ==="
ps aux | grep billing | grep -v grep

echo ""
echo "=== 直接测 billing-service ==="
curl -s http://localhost:9600/health 2>/dev/null || echo "billing-service 未运行"

echo ""
echo "=== 检查 plans 表数据 ==="
psql -U ai_platform -d ai_platform -c "SELECT id, name, price, max_agents, max_messages FROM plans;" 2>/dev/null || echo "无法连接数据库"

echo ""
echo "=== 检查 subscriptions 表 ==="
psql -U ai_platform -d ai_platform -c "SELECT * FROM subscriptions LIMIT 5;" 2>/dev/null || echo "无法连接数据库"

echo ""
echo "=== 通过 gateway 测 /billing/plans ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

curl -s 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== 通过 gateway 测 /billing/subscriptions ==="
curl -s 'http://localhost:9100/api/v1/billing/subscriptions' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== 通过 gateway 测 /billing/billing/quota ==="
curl -s 'http://localhost:9100/api/v1/billing/billing/quota' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== 通过 gateway 测 /billing/billing/records ==="
curl -s 'http://localhost:9100/api/v1/billing/billing/records' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null
