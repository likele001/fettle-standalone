#!/bin/bash
export PATH=/www/server/nodejs/v24.18.0/bin:$PATH

echo "=== 直接用 vite build 编译（跳过类型检查）==="
cd /www/wwwroot/fettle/frontend/web-admin
npx vite build 2>&1

echo ""
echo "=== 检查 dist ==="
ls -la dist/assets/ 2>/dev/null | head -10
