#!/bin/bash
# ============================================================
# fettle 全栈一键发布（Go 后端 + ai-engine + 前端）
# 用法: bash scripts/deploy.sh [选项]
#   --no-fe            跳过前端构建
#   --no-snapshot      不归档 build/ 旧二进制
#   --no-restart       仅构建，不重启任何服务
#   --keep <N>         快照保留 N 份（默认 3）
#   --only <target>    只操作某个目标：go|ai|fe
# 说明: 内部调用 release-fettle.sh + release-ai-engine.sh
#       默认走宝塔 Go 项目管理器 120s 守护策略，不动面板
# ============================================================
set -e
cd "$(dirname "$0")/.."
ROOT="$(pwd)"

DO_FE=1
DO_SNAPSHOT=1
DO_RESTART=1
ONLY=""
KEEP=3

i=1
while [ $i -le $# ]; do
  arg="${!i}"
  case "$arg" in
    --no-fe)       DO_FE=0 ;;
    --no-snapshot) DO_SNAPSHOT=0 ;;
    --no-restart)  DO_RESTART=0 ;;
    --keep)
      next=$((i+1))
      KEEP="${!next}"
      i=$next ;;
    --only)
      next=$((i+1))
      ONLY="${!next}"
      i=$next ;;
    --help|-h)
      sed -n '2,15p' "$0"
      exit 0 ;;
    *) echo "未知参数: $arg"; exit 2 ;;
  esac
  i=$((i+1))
done

GO_ARGS=()
GO_ARGS+=("--baota")
[ "$DO_SNAPSHOT" = "1" ] && GO_ARGS+=("--snapshot" "--keep" "$KEEP")
[ "$DO_RESTART"  = "0" ] && GO_ARGS+=("--no-build")
[ "$DO_RESTART"  = "1" ] && GO_ARGS+=("--restart")
if [ "$DO_RESTART" = "1" ] && [ "$DO_FE" = "1" ]; then
  GO_ARGS+=("--fe")
fi

AI_ARGS=("--restart")

if [ -z "$ONLY" ] || [ "$ONLY" = "go" ]; then
  echo "############ Go 后端 ############"
  bash "$ROOT/scripts/release-fettle.sh" "${GO_ARGS[@]}"
fi
if [ -z "$ONLY" ] || [ "$ONLY" = "ai" ]; then
  echo ""
  echo "############ AI Engine (Python) ############"
  bash "$ROOT/scripts/release-ai-engine.sh" "${AI_ARGS[@]}"
fi
if [ -z "$ONLY" ] || [ "$ONLY" = "fe" ]; then
  if [ "$DO_FE" = "0" ]; then
    echo ""
    echo "前端跳过（--no-fe）"
  else
    echo ""
    echo "############ 前端 ############"
    (cd "$ROOT/frontend/web-admin" && npm run build) || { echo "前端构建失败"; exit 1; }
    if command -v nginx >/dev/null; then
      nginx -s reload && echo "nginx reloaded" || echo "nginx reload 失败，请手动"
    fi
    echo "dist/ 已更新，如未生效请 Ctrl+Shift+R 清缓存"
  fi
fi

echo ""
echo "=== 全栈发布完成 ==="
bash "$ROOT/scripts/release-fettle.sh" --no-build 2>&1 | grep -E "完成|FAIL" || true
bash "$ROOT/scripts/release-ai-engine.sh" --status 2>&1