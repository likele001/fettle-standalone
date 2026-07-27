#!/bin/bash
export PATH=/www/server/nodejs/v24.18.0/bin:$PATH

echo "=== node 版本 ==="
node -v
npm -v

echo ""
echo "=== 编译前端 ==="
cd /www/wwwroot/fettle/frontend/web-admin
npm run build 2>&1

echo ""
echo "=== 检查 dist ==="
ls -la dist/assets/ 2>/dev/null | head -10
