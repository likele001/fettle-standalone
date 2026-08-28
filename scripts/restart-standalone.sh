#!/bin/bash
# ============================================================
# fettle-standalone 一键重启脚本
# 用法: bash scripts/restart-standalone.sh [--with-worker]
#   --with-worker: 同时启动 RAG celery worker
# 说明: standalone 为 nohup 裸进程部署（端口段 20001-20009），
#       与 fettle 主部署（9100-9700）完全隔离。
# 注意: 停止时同时匹配绝对路径(/www/wwwroot/fettle-standalone/build/x)
#       与相对路径(./build/x)两种进程 cmdline 形态。
# ============================================================
set -e
cd /www/wwwroot/fettle-standalone
mkdir -p logs

WITH_WORKER=0
[ "$1" = "--with-worker" ] && WITH_WORKER=1

echo "== 1. 停止旧进程 =="
for p in gateway user-service agent-service chat-service skill-service billing-service; do
  pkill -f "fettle-standalone/build/$p" 2>/dev/null && echo "stopped $p (abs)" || true
  pkill -f "\./build/$p" 2>/dev/null && echo "stopped $p (rel)" || true
done
pkill -f "uvicorn main:app.*--port 20007" 2>/dev/null && echo "stopped ai-engine" || true
  pkill -f "fettlept/bin/uvicorn" 2>/dev/null && echo "stopped ai-engine (fettlept)" || true
if [ "$WITH_WORKER" = "1" ]; then
  pkill -f "celery -A tasks.celery_app" 2>/dev/null && echo "stopped celery worker" || true
fi
sleep 2

# 加载 standalone 根 .env（端口/DB/JWT 等），确保服务监听 20001-20007 段
set -a
[ -f .env ] && source .env
set +a

echo "== 2. 启动 Go 服务（nohup，读 standalone .env -> 20001-20006）=="
for p in gateway user-service agent-service chat-service skill-service billing-service; do
  nohup ./build/$p > logs/$p.log 2>&1 &
  echo "started $p (PID $!)"
done

echo "== 3. 启动 ai-engine (HTTP 20007 / gRPC 20008, fettlept venv) =="
cd backend/ai-engine
nohup /www/server/pyporject_evn/fettlept/bin/uvicorn main:app --host 0.0.0.0 --port 20007 > ../../logs/ai-engine.log 2>&1 &
echo "started ai-engine (PID $!)"
cd ../..

if [ "$WITH_WORKER" = "1" ]; then
  echo "== 4. 启动 RAG celery worker =="
  cd backend/ai-engine
  nohup /www/server/pyporject_evn/fettlept/bin/python3.13 -m celery -A tasks.celery_app worker --loglevel=info --concurrency=2 --hostname=standalone-worker@%h > ../../logs/ai-worker.log 2>&1 &
  echo "started celery worker (PID $!)"
  cd ../..
fi

echo "== 5. 就绪检查 =="
sleep 4
for port in 20001 20002 20003 20004 20005 20006 20007; do
  ss -tln | grep -q ":$port " && echo "OK  :$port" || echo "MISS :$port"
done
echo "完成。日志在 logs/*.log；失败排查: tail logs/<svc>.log"
