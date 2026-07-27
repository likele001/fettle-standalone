#!/bin/bash

CONF="/www/server/panel/vhost/nginx/fettle.cenkor.cn.conf"

# 先备份
cp $CONF ${CONF}.bak.$(date +%Y%m%d%H%M%S)

# 在 access_log 行之前插入 API 代理和 SPA 回退配置
# 用 sed 在 "access_log" 行前面插入
sed -i '/access_log.*fettle.cenkor.cn.log/i\
    # API 代理到 gateway\
    location /api/ {\
        proxy_pass http://127.0.0.1:9100;\
        proxy_set_header Host $host;\
        proxy_set_header X-Real-IP $remote_addr;\
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\
        proxy_set_header X-Forwarded-Proto $scheme;\
        proxy_connect_timeout 60s;\
        proxy_read_timeout 120s;\
        proxy_send_timeout 60s;\
    }\
\
    # SPA 路由回退\
    location / {\
        try_files $uri $uri/ /index.html;\
    }\
' $CONF

echo "=== 修改后的配置 ==="
cat $CONF

echo ""
echo "=== 测试 Nginx 配置 ==="
nginx -t 2>&1

echo ""
echo "=== 重载 Nginx ==="
nginx -s reload 2>&1
echo "重载完成"
