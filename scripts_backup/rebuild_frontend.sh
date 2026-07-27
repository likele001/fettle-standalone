#!/bin/bash

echo "=== 重新编译 web-platform ==="
cd /www/wwwroot/fettle/frontend/web-platform
npm run build 2>&1 | tail -5

echo ""
echo "=== 验证 dist ==="
ls -la dist/assets/ | head -5
echo ""
grep 'src=' dist/index.html
