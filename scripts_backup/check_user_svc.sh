#!/bin/bash

echo "=== user-service 进程环境变量 ==="
US_PID=$(pgrep -f 'user-service\|build/user' | head -1)
if [ -n "$US_PID" ]; then
  cat /proc/$US_PID/environ 2>/dev/null | tr '\0' '\n' | grep JWT
else
  echo "找不到 user-service 进程"
  ps aux | grep -i user | grep -v grep
fi

echo ""
echo "=== user-service 启动方式 ==="
ps aux | grep -i 'user' | grep -v grep

echo ""
echo "=== 检查是否有其他 .env 或 config ==="
find /www/wwwroot/fettle -name ".env" -o -name "config.yaml" -o -name "config.yml" 2>/dev/null | head -10

echo ""
echo "=== 各 .env 的 JWT_SECRET ==="
for f in $(find /www/wwwroot/fettle -name ".env" 2>/dev/null); do
  echo "--- $f ---"
  grep JWT_SECRET "$f" 2>/dev/null || echo "(no JWT_SECRET)"
done
