#!/usr/bin/env bash
# ============================================================
# fettle-standalone 一键安装 · Docker Compose 方案
#
# 用法:
#   bash scripts/install-docker.sh [选项]
#
# 选项:
#   --dry-run           只打印将要执行的动作，不启动任何容器
#   --force             目标端口已被占用时也继续（默认中止）
#   --no-build          不重新构建镜像（用已有镜像启动）
#   --with-milvus       一并启动 Milvus 栈（etcd + minio + milvus，deploy/milvus）
#   --down              反向操作：停掉本方案起的全部容器（不动数据卷）
#   -h, --help          显示本帮助
#
# 设计说明:
#   1. 一个 compose 文件即可跑完全栈（deploy/docker/docker-compose.yml 自带
#      PostgreSQL / Redis / NATS / MinIO + 6 个 Go 服务 + ai-engine + web-admin）。
#   2. Milvus 不在该 compose 内 —— 需要 RAG/向量检索时才起，用 --with-milvus
#      走 deploy/milvus/docker-compose.yml。两个 compose 分属不同 docker 网络，
#      所以容器里的 ai-engine 要经宿主机网桥访问，脚本会自动把 MILVUS_HOST
#      改成 docker0 的地址（默认 172.17.0.1）。
#   3. 安全阀：开工前做端口体检，若这台机器上已有 standalone 在跑，立即中止。
#   4. 数据落 docker volume，--down 不删卷。
# ============================================================
set -euo pipefail

APP_NAME="fettle-standalone"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DC_DIR="$ROOT_DIR/deploy/docker"
MILVUS_DIR="$ROOT_DIR/deploy/milvus"
ENV_FILE="$DC_DIR/.env"

APP_PORTS="20001 20002 20003 20004 20005 20006 20007 20008"
MILVUS_PORTS="19530 9091"

# ---------- 选项 ----------
DRY_RUN=0; FORCE=0; DO_BUILD=1; WITH_MILVUS=0; DO_DOWN=0
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run)     DRY_RUN=1 ;;
    --force)       FORCE=1 ;;
    --no-build)    DO_BUILD=0 ;;
    --with-milvus) WITH_MILVUS=1 ;;
    --down)        DO_DOWN=1 ;;
    -h|--help)     sed -n '2,22p' "$0"; exit 0 ;;
    *) echo "未知参数: $1（用 --help 看用法）" >&2; exit 2 ;;
  esac
  shift
done

c_ok()   { printf '\033[32m✓\033[0m %s\n' "$*"; }
c_warn() { printf '\033[33m!\033[0m %s\n' "$*"; }
c_err()  { printf '\033[31m✗\033[0m %s\n' "$*" >&2; }
step()   { printf '\n\033[36m[%s]\033[0m %s\n' "$1" "$2"; }

run() {
  if [ "$DRY_RUN" = "1" ]; then
    printf '   \033[90m[dry-run]\033[0m %s\n' "$*"
  else
    "$@"
  fi
}

port_busy()  { ss -tlnH 2>/dev/null | grep -q ":$1\b"; }
port_owner() { ss -tlnpH 2>/dev/null | grep ":$1\b" | head -1 | sed -E 's/.*users:\(\("([^"]+)".*/\1/'; }

MAIN_COMPOSE="docker-compose.yml"
DC="docker compose"

# ============================================================
# 0. 环境检查
# ============================================================
step "0/7" "环境检查"

if ! command -v docker >/dev/null 2>&1; then
  c_err "未找到 docker。安装：curl -fsSL https://get.docker.com | sh"
  exit 1
fi
c_ok "$(docker --version)"

if ! docker compose version >/dev/null 2>&1; then
  c_err "未找到 docker compose v2（需要 ≥ 2.20）"
  c_err "Debian/Ubuntu: apt install docker-compose-plugin"
  exit 1
fi
c_ok "$(docker compose version | head -1)"

if ! docker info >/dev/null 2>&1; then
  c_err "当前用户无权访问 Docker。用 sudo 运行，或把用户加进 docker 组："
  c_err "  sudo usermod -aG docker \$USER && newgrp docker"
  exit 1
fi

[ -f "$DC_DIR/$MAIN_COMPOSE" ] || { c_err "缺少 $DC_DIR/$MAIN_COMPOSE"; exit 1; }
c_ok "compose 文件就绪"

# ============================================================
# 1. --down
# ============================================================
if [ "$DO_DOWN" = "1" ]; then
  step "1/1" "停止本方案的容器（保留数据卷）"
  run $DC -f "$DC_DIR/$MAIN_COMPOSE" down
  [ -f "$MILVUS_DIR/docker-compose.yml" ] && run $DC -f "$MILVUS_DIR/docker-compose.yml" down
  c_ok "已停止。数据卷仍在：docker volume ls | grep fettle"
  exit 0
fi

# ============================================================
# 2. 端口体检（安全阀）
# ============================================================
step "1/7" "端口体检"

CHECK_PORTS="$APP_PORTS"
[ "$WITH_MILVUS" = "1" ] && CHECK_PORTS="$CHECK_PORTS $MILVUS_PORTS"

BUSY=""
for p in $CHECK_PORTS; do
  port_busy "$p" && BUSY="$BUSY $p"
done

if [ -n "$BUSY" ]; then
  c_warn "以下端口已被占用：$BUSY"
  for p in $BUSY; do printf '     :%s  ← %s\n' "$p" "$(port_owner "$p")"; done
  if [ "$FORCE" = "0" ]; then
    c_err "这台机器上似乎已经有 ${APP_NAME} 在运行（裸机或宝塔托管）。"
    c_err "继续会与它抢端口。确认要装加 --force；只想看会做什么加 --dry-run。"
    exit 1
  fi
  c_warn "--force 已指定，继续（风险自负）"
else
  c_ok "全部目标端口空闲"
fi

# ============================================================
# 3. 准备 .env
# ============================================================
step "2/7" "准备 $ENV_FILE"

if [ -f "$ENV_FILE" ]; then
  c_ok ".env 已存在，保留"
else
  TPL="$DC_DIR/.env.example"
  [ -f "$TPL" ] || { c_err "缺少 $TPL，无法生成 .env"; exit 1; }
  run cp "$TPL" "$ENV_FILE"
  if [ "$DRY_RUN" = "0" ] && [ -x "$DC_DIR/gen-env.sh" ]; then
    ( cd "$DC_DIR" && ./gen-env.sh "$ENV_FILE" >/dev/null 2>&1 ) \
      && c_ok "已生成随机密钥（JWT_SECRET / POSTGRES_PASSWORD / MinIO）" \
      || c_warn "gen-env.sh 执行失败，请手工检查 .env"
  fi
  c_ok "已从 .env.example 生成 .env"
fi

# Milvus 地址修正：两个 compose 不在同一 docker 网络，容器里必须走宿主机网桥
if [ "$WITH_MILVUS" = "1" ] && [ "$DRY_RUN" = "0" ]; then
  BRIDGE_IP=$(ip -4 addr show docker0 2>/dev/null | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -1)
  BRIDGE_IP="${BRIDGE_IP:-172.17.0.1}"
  if grep -qE '^MILVUS_HOST=' "$ENV_FILE"; then
    sed -i "s|^MILVUS_HOST=.*|MILVUS_HOST=${BRIDGE_IP}|" "$ENV_FILE"
  else
    printf 'MILVUS_HOST=%s\n' "$BRIDGE_IP" >> "$ENV_FILE"
  fi
  c_ok "MILVUS_HOST 已指向宿主机网桥 ${BRIDGE_IP}:19530"
  c_warn "注意：.env.example 里原本写的是 milvus-standalone，但主 compose 内并无该服务，那个值容器里解析不到"
fi

run chmod 600 "$ENV_FILE"

cat <<'EOF'

  需人工确认的项（不填也能起，但对应功能不可用）：
    QWEN_API_KEY / DEEPSEEK_API_KEY   AI 对话
    ALLOWED_ORIGINS                   前端域名，默认 http://localhost:20009（仅本地可用）

EOF

# ============================================================
# 4. Milvus（可选）
# ============================================================
if [ "$WITH_MILVUS" = "1" ]; then
  step "3/7" "启动 Milvus 栈（etcd + minio + milvus）"
  if [ ! -f "$MILVUS_DIR/docker-compose.yml" ]; then
    c_warn "未找到 $MILVUS_DIR/docker-compose.yml，跳过"
  else
    run $DC -f "$MILVUS_DIR/docker-compose.yml" up -d
    c_ok "Milvus 已启动（:19530 / :9091）"
  fi
else
  step "3/7" "Milvus 未启用（需要向量检索时加 --with-milvus）"
  c_warn "若 ai-engine 日志报连不上 Milvus 属预期，RAG 相关功能需先起 Milvus"
fi

# ============================================================
# 5. 起全栈
# ============================================================
step "4/7" "构建并启动全栈（PG/Redis/NATS/MinIO + 6 Go 服务 + ai-engine + web-admin）"

if [ "$DO_BUILD" = "1" ]; then
  run $DC -f "$DC_DIR/$MAIN_COMPOSE" up -d --build
else
  run $DC -f "$DC_DIR/$MAIN_COMPOSE" up -d
fi
c_ok "容器已启动"

# ============================================================
# 6. 健康检查
# ============================================================
step "5/7" "健康检查"

if [ "$DRY_RUN" = "0" ]; then
  printf '   等待服务就绪'
  for _ in $(seq 1 40); do
    sleep 2; printf '.'
    ready=1
    for p in $APP_PORTS; do port_busy "$p" || { ready=0; break; }; done
    [ "$ready" = "1" ] && break
  done
  printf '\n\n'

  FAIL=0
  for p in $APP_PORTS; do
    code=$(curl -s -o /dev/null -w '%{http_code}' --max-time 3 "http://127.0.0.1:$p/health" 2>/dev/null || true)
    if [ "$code" = "200" ]; then
      printf '   \033[32m✓\033[0m :%-5s /health 200\n' "$p"
    else
      printf '   \033[33m·\033[0m :%-5s /health %s\n' "$p" "${code:-无响应}"
      FAIL=1
    fi
  done
  [ "$FAIL" = "1" ] && c_warn "有服务未通过，排查：docker compose -f $DC_DIR/$MAIN_COMPOSE logs -f <服务名>"
fi

# ============================================================
# 7. 收尾
# ============================================================
step "6/7" "容器状态"
run $DC -f "$DC_DIR/$MAIN_COMPOSE" ps

step "7/7" "完成"
cat <<EOF

  fettle-standalone Docker 安装完成。下一步：

  1) 配置 nginx 反向代理（宝塔里加站点即可）
       管理后台  →  http://127.0.0.1:20001
       API       →  http://127.0.0.1:20001
       前端静态  →  $ROOT_DIR/frontend/web-admin/dist

  2) 常用命令（在 $DC_DIR 下）
       看状态:  docker compose ps
       看日志:  docker compose logs -f gateway
       重启:    docker compose restart chat-service
       全停:    bash scripts/install-docker.sh --down      （保留数据卷）

  3) 中间件复用宿主机已有实例
       不想用容器里的 PG/Redis/NATS/MinIO，就编辑 $ENV_FILE 把它们改成外部地址，
       并把 compose 里对应服务注释掉（或改用 deploy/docker/compose 的 infra 部分）。

  4) 中途想看会做什么而不真执行
       bash scripts/install-docker.sh --dry-run

EOF

c_ok "fettle-standalone Docker 安装流程结束"
