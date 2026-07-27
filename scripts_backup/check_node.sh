#!/bin/bash
echo "=== node 环境 ==="
which node 2>/dev/null && node -v || echo "node 未安装"
which npm 2>/dev/null && npm -v || echo "npm 未安装"
find / -name "node" -type f 2>/dev/null | head -3
echo ""
echo "=== 前端源码 ==="
ls /www/wwwroot/fettle/frontend/web-admin/package.json 2>/dev/null || echo "web-admin 不存在"
ls /www/wwwroot/fettle/frontend/web-admin/dist/index.html 2>/dev/null || echo "web-admin dist 不存在"
