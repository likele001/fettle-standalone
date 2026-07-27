#!/bin/bash

echo "=== extension 配置 ==="
ls /www/server/panel/vhost/nginx/extension/fettle.cenkor.cn/ 2>/dev/null
cat /www/server/panel/vhost/nginx/extension/fettle.cenkor.cn/*.conf 2>/dev/null

echo ""
echo "=== rewrite 配置 ==="
cat /www/server/panel/vhost/rewrite/fettle.cenkor.cn.conf 2>/dev/null
