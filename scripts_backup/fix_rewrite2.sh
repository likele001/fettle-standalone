#!/bin/bash

REWRITE="/www/server/panel/vhost/rewrite/fettle.cenkor.cn.conf"

# rewrite 需要把 /api/xxx 重写为 /api/v1/xxx 再代理到 gateway
cat > $REWRITE << 'EOF'
location /api/ {
    rewrite ^/api/(.*)$ /api/v1/$1 break;
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
