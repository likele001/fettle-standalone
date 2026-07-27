#!/bin/bash

echo "=== fettle.cenkor.cn Nginx 配置 ==="
grep -A 30 'fettle.cenkor.cn' /www/server/panel/vhost/nginx/fettle.cenkor.cn.conf 2>/dev/null || \
grep -A 30 'fettle.cenkor.cn' /www/server/panel/vhost/nginx/*.conf 2>/dev/null | head -50

echo ""
echo "=== 所有 Nginx 站点配置 ==="
ls /www/server/panel/vhost/nginx/ 2>/dev/null

echo ""
echo "=== 测试直接请求 gateway ==="
curl -s -w '\nHTTP %{http_code}\n' -X POST 'http://localhost:9100/api/v1/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | head -5
