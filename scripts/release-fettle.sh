#!/bin/bash
# ============================================================
# fettle 一键构建 + 重启脚本（Go 后端 6 服务）
# 用法: bash scripts/release-fettle.sh [选项]
#   --no-build         跳过构建
#   --build            仅构建，不重启（默认）
#   --restart          构建后批量重启
#   --baota|--systemctl|--nohup   重启策略（默认 baota）
#   --only <svc>       只操作指定服务（如 chat-service）
#   --log-dir <path>   日志目录（默认 /www/wwwlogs/go，宝塔标准）
#   --pid-dir <path>   宝塔 Go 守护 PID 目录（默认 /var/tmp/gopids）
#   --skip-fe          跳过前端 npm run build
#   --fe               构建完执行 npm run build + nginx -s reload
#   --snapshot         构建前归档 build/<svc> 到 build/_rollback/<TS>/，可配合 --keep
#   --keep <N>         配合 --snapshot，仅保留最近 N 份快照（默认 3）
# 说明: 产物输出到 build/；日志与 PID 目录可参数化，便于非宝塔部署
#       支持 fettle（默认）与 fettle-standalone 独立部署
# ============================================================
set -e
cd "$(dirname "$0")/.."
ROOT="$(pwd)"
BACKEND="$ROOT/backend"
BUILD="$ROOT/build"
LOG_DIR="${LOG_DIR:-$ROOT/logs}"
PID_DIR="${PID_DIR:-/var/tmp/gopids}"
export PATH=/usr/local/go/bin:$PATH

DO_BUILD=1
DO_RESTART=0
RESTART_MODE="baota"
ONLY_SVC=""
DO_FE=0
SKIP_FE=0
SNAPSHOT=0
KEEP=3

i=1
while [ $i -le $# ]; do
  arg="${!i}"
  case "$arg" in
    --no-build) DO_BUILD=0 ;;
    --restart)  DO_RESTART=1 ;;
    --baota)     RESTART_MODE="baota" ;;
    --systemctl) RESTART_MODE="systemctl" ;;
    --nohup)    RESTART_MODE="nohup" ;;
    --only)
      next=$((i+1))
      ONLY_SVC="${!next}"
      i=$next ;;
    --log-dir)
      next=$((i+1))
      LOG_DIR="${!next}"
      i=$next ;;
    --pid-dir)
      next=$((i+1))
      PID_DIR="${!next}"
      i=$next ;;
    --fe)       DO_FE=1 ;;
    --skip-fe)  SKIP_FE=1 ;;
    --build)    ;;
    --snapshot) SNAPSHOT=1 ;;
    --keep)
      next=$((i+1))
      KEEP="${!next}"
      i=$next ;;
    *) echo "未知参数: $arg"; exit 2 ;;
  esac
  i=$((i+1))
done

SERVICES="gateway user-service agent-service chat-service skill-service billing-service"
if [ -n "$ONLY_SVC" ]; then
  SERVICES="$ONLY_SVC"
fi

# ---------- 1. 构建 ----------
if [ "$DO_BUILD" = "1" ]; then
  mkdir -p "$BUILD"
  if [ "$SNAPSHOT" = "1" ]; then
    SNAP_TS=$(date +%Y%m%d_%H%M%S)
    SNAP_DIR="$BUILD/_rollback/$SNAP_TS"
    mkdir -p "$SNAP_DIR"
    echo "== snapshot: $SNAP_DIR =="
    for svc in $SERVICES; do
      if [ -f "$BUILD/$svc" ]; then
        cp -p "$BUILD/$svc" "$SNAP_DIR/$svc"
      fi
    done
    # 清理旧快照，仅保留最近 KEEP 个
    if [ -d "$BUILD/_rollback" ]; then
      ls -1dt "$BUILD/_rollback"/*/ 2>/dev/null | tail -n +$((KEEP + 1)) | xargs -r rm -rf
    fi
  fi
  for svc in $SERVICES; do
    echo "== build $svc =="
    (cd "$BACKEND/$svc" && go build -o "$BUILD/$svc" .) || { echo "!! FAIL $svc"; exit 1; }
    echo "   -> $BUILD/$svc ($(date +%H:%M:%S))"
  done
  echo "构建完成"
fi

# ---------- 2. 重启 ----------
restart_baota() {
  echo "== 宝塔 Go 项目管理器批量重启（pkill + 守护自启） =="
  for svc in $SERVICES; do
    local pid_file="${PID_DIR}/${svc}.pid"
    if [ -f "$pid_file" ]; then
      local pid
      pid=$(cat "$pid_file" 2>/dev/null || echo "")
      if [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null; then
        echo "   kill $svc (pid=$pid)"
        kill "$pid" 2>/dev/null || true
        sleep 2
        kill -9 "$pid" 2>/dev/null || true
      fi
    fi
    pkill -f "$BUILD/$svc" 2>/dev/null || true
    sleep 1
    if pgrep -f "$BUILD/$svc" >/dev/null; then
      echo "   守护未拉起，等待 3s"
      sleep 3
    fi
    if pgrep -f "$BUILD/$svc" >/dev/null; then
      echo "   OK $svc (pid=$(pgrep -f $BUILD/$svc | head -1))"
    else
      echo "   !! $svc 未启动，检查 ${LOG_DIR}/${svc}.log"
    fi
  done
}

restart_systemctl() {
  echo "== systemctl restart =="
  systemctl restart fettle-gateway fettle-user fettle-agent fettle-chat fettle-skill fettle-billing
  sleep 2
  systemctl is-active fettle-gateway fettle-user fettle-agent fettle-chat fettle-skill fettle-billing
}

restart_nohup() {
  echo "== nohup 兜底重启 =="
  for svc in $SERVICES; do
    pkill -f "$BUILD/$svc" 2>/dev/null || true
    sleep 1
    nohup "$BUILD/$svc" >> "${LOG_DIR}/${svc}.log" 2>&1 < /dev/null &
    disown 2>/dev/null || true
    sleep 1
    if pgrep -f "$BUILD/$svc" >/dev/null; then
      echo "   OK $svc"
    else
      echo "   !! $svc 未启动"
    fi
  done
}

if [ "$DO_RESTART" = "1" ]; then
  case "$RESTART_MODE" in
    baota)     restart_baota ;;
    systemctl) restart_systemctl ;;
    nohup)     restart_nohup ;;
  esac
fi

# ---------- 3. 自检 ----------
if [ "$DO_RESTART" = "1" ]; then
  echo "== 健康检查 =="
  for port in 9100 9200 9300 9400 9500 9600; do
    code=$(curl -s -o /dev/null -w "%{http_code}" --max-time 3 "http://127.0.0.1:${port}/health" || echo "000")
    echo "   :${port}/health -> ${code}"
  done
fi

# ---------- 4. 前端 ----------
if [ "$DO_FE" = "1" ] && [ "$SKIP_FE" = "0" ]; then
  echo "== 构建前端 =="
  (cd "$ROOT/frontend/web-admin" && npm run build) || { echo "!! 前端构建失败"; exit 1; }
  if command -v nginx >/dev/null; then
    nginx -s reload && echo "   nginx reloaded" || echo "   nginx reload 失败，请手动"
  fi
  echo "   dist/ 已更新，如未生效请 Ctrl+Shift+R 清缓存"
fi

echo "完成。若改动 Python（ai-engine），请重启 fettle-ai-engine。"