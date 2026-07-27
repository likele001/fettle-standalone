#!/bin/bash
export PATH=/www/server/nodejs/v24.18.0/bin:$PATH

echo "=== 编译 web-platform 前端 ==="
cd /www/wwwroot/fettle/frontend/web-platform
npx vite build 2>&1

echo ""
echo "=== 检查 dist ==="
ls -la dist/assets/ 2>/dev/null | head -10
