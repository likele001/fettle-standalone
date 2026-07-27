#!/bin/bash

TOKEN=$(curl -s -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)

echo "=== 支付配置 ==="
curl -s 'http://localhost:9100/api/v1/admin/payment/config' \
  -H "Authorization: Bearer $TOKEN" | python3 -m json.tool 2>/dev/null

echo ""
echo "=== 套餐列表 ==="
curl -s 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; [print(f"{p[\"id\"]} {p[\"name\"]} ¥{p[\"price\"]}") for p in json.load(sys.stdin)]' 2>/dev/null

echo ""
echo "=== 创建订单 ==="
PLAN_ID=$(curl -s 'http://localhost:9100/api/v1/billing/plans' \
  -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; plans=json.load(sys.stdin); print([p["id"] for p in plans if p.get("price",0)>0][0] if plans else "")' 2>/dev/null)
echo "Plan ID: $PLAN_ID"

curl -s -X POST 'http://localhost:9100/api/v1/billing/payment/create' \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'X-Frontend-URL: http://fettle.cenkor.cn' \
  -d "{\"plan_id\": \"$PLAN_ID\"}" | python3 -m json.tool 2>/dev/null
