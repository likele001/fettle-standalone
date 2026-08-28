#!/bin/bash
# ============================================================
# fettle 回滚脚本（按时间戳版本回滚 build 二进制）
# 用法: bash scripts/rollback.sh [选项]
#   --list               列出所有可用历史版本
#   --to <timestamp>     回滚到指定时间戳版本（如 20260828_231821）
#   --latest             回滚到最近一次保留的备份
#   --keep <N>           保留最近 N 个备份（默认 3），更老的删除
#   --build-dir <path>   build 目录（默认项目根/build）
#   --src <dir>          备份目录（默认 build/_rollback）
# 备注: 此脚本只动 build/ 下的二进制；源码已用 git 管理，源码错误请 git revert
# ============================================================
set -e
cd "$(dirname "$0")/.."
ROOT="$(pwd)"
BUILD_DIR="${BUILD_DIR:-$ROOT/build}"
SRC_DIR="${SRC_DIR:-$BUILD_DIR/_rollback}"
SERVICES="gateway user-service agent-service chat-service skill-service billing-service"

LIST_ONLY=0
TO_TS=""
USE_LATEST=0
KEEP=3

i=1
while [ $i -le $# ]; do
  arg="${!i}"
  case "$arg" in
    --list)        LIST_ONLY=1 ;;
    --to)
      next=$((i+1))
      TO_TS="${!next}"
      i=$next ;;
    --latest)      USE_LATEST=1 ;;
    --keep)
      next=$((i+1))
      KEEP="${!next}"
      i=$next ;;
    --build-dir)
      next=$((i+1))
      BUILD_DIR="${!next}"
      i=$next ;;
    --src)
      next=$((i+1))
      SRC_DIR="${!next}"
      i=$next ;;
    *) echo "未知参数: $arg"; exit 2 ;;
  esac
  i=$((i+1))
done

if [ ! -d "$SRC_DIR" ]; then
  echo "无回滚备份目录: $SRC_DIR"
  echo "提示：先在 release-fettle.sh 构建后调用 --snapshot 留档"
  exit 1
fi

echo "== 备份目录: $SRC_DIR =="
ls -lh "$SRC_DIR" 2>&1 | tail -20

if [ "$LIST_ONLY" = "1" ]; then
  exit 0
fi

if [ "$USE_LATEST" = "1" ]; then
  TO_TS=$(ls -1 "$SRC_DIR" 2>/dev/null | grep -E '^[0-9]{8}_[0-9]{6}$' | sort -r | head -1)
  if [ -z "$TO_TS" ]; then
    echo "无任何备份快照"
    exit 1
  fi
  echo "自动选择最新版本: $TO_TS"
fi

if [ -z "$TO_TS" ]; then
  echo "用法: --to <timestamp> 或 --latest"
  echo "可用版本:"
  ls -1 "$SRC_DIR" | grep -E '^[0-9]{8}_[0-9]{6}$' | sort -r
  exit 2
fi

SNAP_DIR="$SRC_DIR/$TO_TS"
if [ ! -d "$SNAP_DIR" ]; then
  echo "未找到快照: $SNAP_DIR"
  exit 1
fi

echo "== 回滚至 $TO_TS =="
for svc in $SERVICES; do
  src="$SNAP_DIR/$svc"
  if [ -f "$src" ]; then
    cp -p "$src" "$BUILD_DIR/$svc"
    echo "   OK $svc <- $src"
  else
    echo "   skip $svc（无备份）"
  fi
done

echo ""
echo "回滚完成。下一步："
echo "  bash scripts/release-fettle.sh --no-build --restart   触发宝塔守护拉起"
echo "  或直接：pkill -f build/<svc> && 由面板 120s 守护自启"