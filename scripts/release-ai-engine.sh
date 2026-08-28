#!/bin/bash
# ============================================================
# fettle ai-engine 一键管理脚本（Python FastAPI + Celery）
# 用法: bash scripts/release-ai-engine.sh [选项]
#   --status           查看 ai-engine + celery 状态
#   --start            启动（uvicorn + celery），已运行则跳过
#   --stop             停止（uvicorn + celery）
#   --restart          重启
#   --only <comp>      只操作某组件：uvicorn|celery
#   --cwd <path>       ai-engine 目录（默认 backend/ai-engine）
#   --venv <path>      Python venv（默认 /www/server/pyporject_evn/fettlept）
#   --log <path>       uvicorn 日志路径（默认 ./logs/ai-engine.log）
#   --celery-log <p>   celery 日志路径（默认 ./logs/ai-worker.log）
#   --workers <N>      celery worker 数（默认 4）
#   --http-port <N>    uvicorn 端口（默认 20007，读 APP_HTTP_PORT/env）
#   --bind <ip:port>   uvicorn 绑定地址（默认 0.0.0.0:20007）
# 说明: 不走宝塔面板，靠 nohup + 进程检测；可与 systemd 共存
#       默认值适配 fettle-standalone（venv fettlept，端口 20007，日志 ./logs/）
# ============================================================
set -e
cd "$(dirname "$0")/.."
ROOT="$(pwd)"

DO_STATUS=0
DO_START=0
DO_STOP=0
DO_RESTART=0
ONLY=""

CWD="${CWD:-$ROOT/backend/ai-engine}"
VENV="${VENV:-/www/server/pyporject_evn/fettlept}"
LOG="${LOG:-$ROOT/logs/ai-engine.log}"
CELERY_LOG="${CELERY_LOG:-$ROOT/logs/ai-worker.log}"
WORKERS="${WORKERS:-4}"
HTTP_PORT="${APP_HTTP_PORT:-20007}"
BIND="${BIND:-0.0.0.0:20007}"

i=1
while [ $i -le $# ]; do
  arg="${!i}"
  case "$arg" in
    --status)  DO_STATUS=1 ;;
    --start)   DO_START=1 ;;
    --stop)    DO_STOP=1 ;;
    --restart) DO_RESTART=1 ;;
    --only)
      next=$((i+1))
      ONLY="${!next}"
      i=$next ;;
    --cwd)
      next=$((i+1))
      CWD="${!next}"
      i=$next ;;
    --venv)
      next=$((i+1))
      VENV="${!next}"
      i=$next ;;
    --log)
      next=$((i+1))
      LOG="${!next}"
      i=$next ;;
    --celery-log)
      next=$((i+1))
      CELERY_LOG="${!next}"
      i=$next ;;
    --workers)
      next=$((i+1))
      WORKERS="${!next}"
      i=$next ;;
    --http-port)
      next=$((i+1))
      HTTP_PORT="${!next}"
      i=$next ;;
    --bind)
      next=$((i+1))
      BIND="${!next}"
      i=$next ;;
    *) echo "未知参数: $arg"; exit 2 ;;
  esac
  i=$((i+1))
done

UVICORN_BIN="$VENV/bin/uvicorn"
CELERY_BIN="$VENV/bin/celery"

if [ ! -x "$UVICORN_BIN" ]; then
  echo "未找到 uvicorn: $UVICORN_BIN"
  exit 1
fi
if [ ! -x "$CELERY_BIN" ]; then
  echo "未找到 celery: $CELERY_BIN"
  exit 1
fi

UVICORN_CMD="setsid nohup $VENV/bin/python3 $UVICORN_BIN main:app --host $BIND"
CELERY_CMD="cd $CWD && set -a; source $CWD/.env; set +a && nohup $CELERY_BIN -A tasks.celery_app worker -Q fettle -n fettle-aiengine@%h --loglevel=info --concurrency=$WORKERS"

uvicorn_pid() {
  pgrep -f "uvicorn main:app.*--port $HTTP_PORT" 2>/dev/null | head -1
}
celery_pids() {
  pgrep -f "celery.*tasks.celery_app.*-Q fettle" 2>/dev/null
}

start_uvicorn() {
  local pid
  pid=$(uvicorn_pid)
  if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
    echo "uvicorn 已在运行 (pid=$pid)"
    return 0
  fi
  mkdir -p "$(dirname "$LOG")"
  (cd "$CWD" && eval $UVICORN_CMD) >> "$LOG" 2>&1 < /dev/null &
  disown 2>/dev/null || true
  sleep 2
  pid=$(uvicorn_pid)
  if [ -n "$pid" ]; then
    echo "uvicorn 已启动 (pid=$pid, http=$BIND, log=$LOG)"
  else
    echo "uvicorn 启动失败，检查 $LOG"
    return 1
  fi
}

stop_uvicorn() {
  local pid
  pid=$(uvicorn_pid)
  if [ -z "$pid" ]; then
    echo "uvicorn 未运行"
    return 0
  fi
  kill "$pid" 2>/dev/null || true
  sleep 2
  kill -9 "$pid" 2>/dev/null || true
  echo "uvicorn 已停止"
}

start_celery() {
  local cnt
  cnt=$(celery_pids | wc -l)
  if [ "$cnt" -gt 0 ]; then
    echo "celery 已在运行 ($cnt 个 worker)"
    return 0
  fi
  bash -c "$CELERY_CMD" >> "$CELERY_LOG" 2>&1 < /dev/null &
  disown 2>/dev/null || true
  sleep 3
  cnt=$(celery_pids | wc -l)
  if [ "$cnt" -gt 0 ]; then
    echo "celery 已启动 ($cnt 个 worker, log=$CELERY_LOG)"
  else
    echo "celery 启动失败，检查 $CELERY_LOG"
    return 1
  fi
}

stop_celery() {
  local pids
  pids=$(celery_pids)
  if [ -z "$pids" ]; then
    echo "celery 未运行"
    return 0
  fi
  for p in $pids; do
    kill "$p" 2>/dev/null || true
  done
  sleep 2
  for p in $pids; do
    kill -9 "$p" 2>/dev/null || true
  done
  echo "celery 已停止"
}

show_status() {
  local pid
  pid=$(uvicorn_pid)
  if [ -n "$pid" ]; then
    echo "uvicorn: 运行中 (pid=$pid, http=$BIND)"
  else
    echo "uvicorn: 未运行"
  fi
  local cnt
  cnt=$(celery_pids | wc -l)
  if [ "$cnt" -gt 0 ]; then
    echo "celery: 运行中 ($cnt 个 worker)"
  else
    echo "celery: 未运行"
  fi
}

case "${DO_STATUS}${DO_START}${DO_STOP}${DO_RESTART}" in
  1000) show_status ;;
  0100)
    [ "$ONLY" = "" ] || [ "$ONLY" = "uvicorn" ] && start_uvicorn
    [ "$ONLY" = "" ] || [ "$ONLY" = "celery" ] && start_celery
    ;;
  0010)
    [ "$ONLY" = "" ] || [ "$ONLY" = "celery" ] && stop_celery
    [ "$ONLY" = "" ] || [ "$ONLY" = "uvicorn" ] && stop_uvicorn
    ;;
  0001)
    [ "$ONLY" = "" ] || [ "$ONLY" = "celery" ] && (stop_celery; start_celery)
    [ "$ONLY" = "" ] || [ "$ONLY" = "uvicorn" ] && (stop_uvicorn; start_uvicorn)
    ;;
  *)
    show_status
    ;;
esac