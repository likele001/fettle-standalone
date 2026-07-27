#!/bin/bash
echo "=== 前端部署目录 ==="
ls -la /www/wwwroot/fettle/frontend/web-admin/ 2>/dev/null | head -10

echo ""
echo "=== dist 目录 ==="
ls -la /www/wwwroot/fettle/frontend/web-admin/dist/ 2>/dev/null | head -10

echo ""
echo "=== 是否有 node_modules ==="
ls /www/wwwroot/fettle/frontend/web-admin/node_modules/ 2>/dev/null | head -5 || echo "没有 node_modules"

echo ""
echo "=== 检查是否有全局 node ==="
find /usr/local /usr/bin /opt -name "node" -type f 2>/dev/null | head -3
find /root -name "node" -type f 2>/dev/null | head -3

echo ""
echo "=== 检查 nvm ==="
ls -la /root/.nvm/versions/node/ 2>/dev/null || echo "没有 nvm"

echo ""
echo "=== 检查宝塔面板 node 管理 ==="
ls /www/server/nodejs/ 2>/dev/null || echo "没有宝塔 node"
ls /www/server/nvm/ 2>/dev/null || echo "没有宝塔 nvm"
