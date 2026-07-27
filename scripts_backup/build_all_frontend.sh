#!/bin/bash
export PATH=/www/server/nodejs/v24.18.0/bin:$PATH

echo "=== 编译 web-admin ==="
cd /www/wwwroot/fettle/frontend/web-admin
npx vite build 2>&1
if [ $? -ne 0 ]; then echo "❌ web-admin 编译失败"; exit 1; fi
echo "✅ web-admin 编译成功"

echo ""
echo "=== 编译 web-platform ==="
cd /www/wwwroot/fettle/frontend/web-platform
npx vite build 2>&1
if [ $? -ne 0 ]; then echo "❌ web-platform 编译失败"; exit 1; fi
echo "✅ web-platform 编译成功"

echo ""
echo "=== 完成 ==="
