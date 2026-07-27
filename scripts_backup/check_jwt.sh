#!/bin/bash

echo "=== .env 中的 JWT_SECRET ==="
grep JWT_SECRET /www/wwwroot/fettle/.env

echo ""
echo "=== 登录获取 token ==="
RESP=$(curl -s -X POST 'http://localhost:9100/api/v1/admin/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}')
TOKEN=$(echo $RESP | python3 -c 'import sys,json; print(json.load(sys.stdin)["data"]["tokens"]["access_token"])' 2>/dev/null)
echo "Token: ${TOKEN:0:50}..."

echo ""
echo "=== 用 python 验证 token ==="
python3 -c "
import jwt, json
token = '$TOKEN'
secret = 'fettle_ai_platform_jwt_secret_key_2026'
try:
    decoded = jwt.decode(token, secret, algorithms=['HS256'])
    print('Decode OK:', json.dumps(decoded, indent=2))
except Exception as e:
    print('Decode FAILED:', e)
"

echo ""
echo "=== 通过 gateway 测 ==="
curl -s 'http://localhost:9100/api/v1/admin/tenants?page_size=1' \
  -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("gateway:", d.get("code"), d.get("message"))' 2>/dev/null

echo ""
echo "=== 直接测 user-service ==="
curl -s 'http://localhost:9200/admin/tenants?page_size=1' \
  -H "Authorization: Bearer $TOKEN" | python3 -c 'import sys,json; d=json.load(sys.stdin); print("user-service:", d.get("code"), d.get("message"))' 2>/dev/null

echo ""
echo "=== gateway 进程环境变量 ==="
cat /proc/$(pgrep -f 'build/gateway' | head -1)/environ 2>/dev/null | tr '\0' '\n' | grep JWT
