#!/bin/bash

echo "=== 检查现有表和数据 ==="
psql -U ai_platform -d ai_platform -c "\dt" 2>/dev/null | grep -E "plan|subscri|billing"

echo ""
echo "=== subscriptions 表结构 ==="
psql -U ai_platform -d ai_platform -c "\d subscriptions" 2>/dev/null

echo ""
echo "=== subscriptions 数据量 ==="
psql -U ai_platform -d ai_platform -c "SELECT count(*) FROM subscriptions;" 2>/dev/null

echo ""
echo "=== plans 数据 ==="
psql -U ai_platform -d ai_platform -c "SELECT * FROM plans;" 2>/dev/null

echo ""
echo "=== 删除旧的 subscriptions 表（空表，重建） ==="
psql -U ai_platform -d ai_platform -c "DROP TABLE IF EXISTS subscriptions CASCADE;" 2>/dev/null
echo "已删除"

echo ""
echo "=== 删除旧的 billing_records 表 ==="
psql -U ai_platform -d ai_platform -c "DROP TABLE IF EXISTS billing_records CASCADE;" 2>/dev/null
echo "已删除"

echo ""
echo "=== 重新编译并启动 billing-service ==="
export PATH=$PATH:/usr/local/go/bin
pkill billing-service || true
sleep 1

cd /www/wwwroot/fettle/backend/billing-service
go build -o /www/wwwroot/fettle/build/billing-service . 2>&1
if [ $? -ne 0 ]; then echo "编译失败"; exit 1; fi
echo "编译成功"

set -a
source /www/wwwroot/fettle/.env
set +a
nohup /www/wwwroot/fettle/build/billing-service > /www/wwwroot/fettle/logs/billing-service.log 2>&1 &
sleep 3

echo ""
echo "=== 进程 ==="
ps aux | grep billing-service | grep -v grep

echo ""
echo "=== 日志 ==="
tail -5 /www/wwwroot/fettle/logs/billing-service.log

echo ""
echo "=== 插入默认套餐 ==="
psql -U ai_platform -d ai_platform << 'SQL'
INSERT INTO plans (id, name, description, price, period, max_agents, max_messages, features, is_active) VALUES
('11111111-1111-1111-1111-111111111111', '免费版', '适合个人体验', 0, 'month', 1, 100, '["1个智能体","100条消息/月","基础功能"]', true),
('22222222-2222-2222-2222-222222222222', '专业版', '适合小型团队', 299, 'month', 5, 10000, '["5个智能体","10000条消息/月","高级功能","优先支持","API接入"]', true),
('33333333-3333-3333-3333-333333333333', '企业版', '适合大型企业', 999, 'month', 999, 999999, '["无限智能体","无限消息","全部功能","专属支持","私有部署","SLA保障"]', true)
ON CONFLICT (id) DO NOTHING;
SQL

echo ""
echo "=== 验证 plans ==="
psql -U ai_platform -d ai_platform -c "SELECT id, name, price, max_agents, max_messages FROM plans;" 2>/dev/null

echo ""
echo "=== 通过 gateway 测试 ==="
TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "--- /billing/plans ---"
curl -s 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "--- /billing/subscriptions ---"
curl -s 'http://localhost:9100/api/v1/billing/subscriptions' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "--- /billing/billing/quota ---"
curl -s 'http://localhost:9100/api/v1/billing/billing/quota' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null
