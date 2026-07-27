#!/bin/bash

REWRITE="/www/server/panel/vhost/rewrite/fettle.cenkor.cn.conf"
CONF="/www/server/panel/vhost/nginx/fettle.cenkor.cn.conf"

# 1. 修复 rewrite 配置：/api/v1/ -> /api/
cat > $REWRITE << 'EOF'
location /api/ {
    proxy_pass http://127.0.0.1:9100;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_connect_timeout 60s;
    proxy_read_timeout 120s;
    proxy_send_timeout 60s;
}

location / {
    try_files $uri $uri/ /index.html;
}
EOF

# 2. 去掉主配置里重复的 location 块（恢复备份）
cp ${CONF}.bak.* $CONF 2>/dev/null || true

echo "=== rewrite 配置 ==="
cat $REWRITE

echo ""
echo "=== 测试 Nginx ==="
nginx -t 2>&1

echo ""
echo "=== 重载 Nginx ==="
nginx -s reload 2>&1

echo ""
echo "=== 测试登录 ==="
curl -s -w '\nHTTP %{http_code}\n' -X POST 'https://fettle.cenkor.cn/api/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"phone":"13800000000","password":"Admin@2026"}' | head -3
