#!/bin/bash
cd /www/wwwroot/fettle

export APP_ENV=production
export JWT_SECRET=fettle_ai_platform_jwt_secret_key_2026
export JWT_EXPIRE_HOURS=72
export DB_HOST=127.0.0.1
export DB_PORT=5432
export DB_USER=ai_platform
export DB_PASSWORD=ai_platform_secret
export DB_NAME=ai_platform
export DB_SSL_MODE=disable
export REDIS_ADDR=127.0.0.1:6379
export USER_SERVICE_PORT=9200
export CHAT_SERVICE_PORT=9400
export AI_ENGINE_PORT=9700
export AI_ENGINE_URL=http://localhost:9700

nohup ./build/user-service > logs/user-service.log 2>&1 &
echo "user-service started"

nohup ./build/chat-service > logs/chat-service.log 2>&1 &
echo "chat-service started"

sleep 2
ps aux | grep -E 'user-service|chat-service' | grep -v grep