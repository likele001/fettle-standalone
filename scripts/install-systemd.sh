#!/usr/bin/env bash
# ============================================================
# fettle-standalone 一键安装 · 裸机 systemd 方案
#
# 用法:
#   sudo bash scripts/install-systemd.sh [选项]
#
# 选项:
#   --dry-run           只打印将要执行的动作，不改动任何文件/服务
#   --force             目标端口已被占用时也继续（默认中止）
#   --skip-infra-check  跳过中间件连通性检测
#   --no-build          跳过 Go / ai-engine 构建，只装单元并启动
#   --with-fe           额外构建前端（web-admin）
#   --force-env         覆盖已存在的 .env（默认保留）
#   -h, --help          显示本帮助
#
# 设计说明:
#   1. 中间件（PostgreSQL/Redis/NATS/MongoDB/Milvus/MinIO）只检测**不安装**，
#      缺失时给出安装指引后退出 —— 复用机器上已有的。
#      注意 standalone 有自己的 NATS 实例（端口 4223），不与人共用 4222。
#   2. 幂等：可反复执行；已存在的 .env 与单元默认不覆盖。
#   3. 安全阀：开工前先做端口体检，若发现这台机器上已经有 standalone 在跑
#      （裸机或宝塔托管），立即中止并列出占位进程。
#   4. 本脚本不碰宝塔面板项目。若同时用面板，务必确认同一服务**只有一个权威**
#      （面板或 systemd 二选一），否则会每 120 秒抢一次端口。
# ============================================================
set -euo pipefail

# ---------- 常量（本机既定值，勿随意改） ----------
APP_NAME="fettle-standalone"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_USER="${FTL_RUN_USER:-fettle-std}"
LOG_DIR="/var/log/fettle"
BUILD_DIR="$ROOT_DIR/build"
UNIT_SRC="$ROOT_DIR/deploy/systemd"
UNIT_DST="/etc/systemd/system"

GO_SERVICES="gateway user-service agent-service chat-service skill-service billing-service"
APP_PORTS="20001 20002 20003 20004 20005 20006 20007"
INFRA_PORTS="5432 6379 4223 27017 19530 9000"
INFRA_NAMES="PostgreSQL Redis NATS(4223) MongoDB Milvus MinIO"

# ---------- 选项 ----------
DRY_RUN=0; FORCE=0; SKIP_INFRA=0; DO_BUILD=1; WITH_FE=0; FORCE_ENV=0
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run)          DRY_RUN=1 ;;
    --force)            FORCE=1 ;;
    --skip-infra-check) SKIP_INFRA=1 ;;
    --no-build)         DO_BUILD=0 ;;
    --with-fe)          WITH_FE=1 ;;
    --force-env)        FORCE_ENV=1 ;;
    -h|--help)          sed -n '2,21p' "$0"; exit 0 ;;
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

# ============================================================
# 0. 环境检查
# ============================================================
step "0/9" "环境检查"

if [ "$DRY_RUN" = "0" ] && [ "$(id -u)" != "0" ]; then
  c_err "需要 root 权限运行（systemd 单元与 /var/log 都要写）"
  exit 1
fi
c_ok "运行身份: $(id -un)"

if ! command -v systemctl >/dev/null 2>&1; then
  c_err "未找到 systemctl —— 本脚本只支持 systemd 发行版"
  c_err "如需 Docker 部署请改用 scripts/install-docker.sh"
  exit 1
fi
c_ok "systemd: $(systemctl --version | head -1)"

[ -d "$ROOT_DIR/backend" ] || { c_err "未在预期目录找到 backend/：$ROOT_DIR"; exit 1; }
c_ok "项目目录: $ROOT_DIR"

export PATH=/usr/local/go/bin:$PATH
if [ "$DO_BUILD" = "1" ]; then
  if ! command -v go >/dev/null 2>&1; then
    c_err "未找到 go —— 编译后端需要 Go ≥ 1.22"
    c_err "安装: wget https://go.dev/dl/go1.22.5.linux-amd64.tar.gz && tar -C /usr/local -xzf go1.22.5.linux-amd64.tar.gz"
    exit 1
  fi
  c_ok "go: $(go version | awk '{print $3}')"
fi

PY_BIN=""
for cand in python3.13 python3.12 python3.11 python3; do
  command -v "$cand" >/dev/null 2>&1 && { PY_BIN="$cand"; break; }
done
[ -n "$PY_BIN" ] || { c_err "未找到 python3（AI 引擎需要 ≥ 3.11）"; exit 1; }
c_ok "python: $PY_BIN ($($PY_BIN -V 2>&1 | awk '{print $2}'))"

if [ "$WITH_FE" = "1" ]; then
  command -v node >/dev/null 2>&1 || { c_err "未找到 node（--with-fe 需要 Node ≥ 18）"; exit 1; }
  c_ok "node: $(node -v)"
fi

# ============================================================
# 1. 端口体检（安全阀）
# ============================================================
step "1/9" "端口体检"

BUSY=""
for p in $APP_PORTS; do port_busy "$p" && BUSY="$BUSY $p"; done

if [ -n "$BUSY" ]; then
  c_warn "以下目标端口已被占用：$BUSY"
  for p in $BUSY; do printf '     :%s  ← %s\n' "$p" "$(port_owner "$p")"; done
  if [ "$FORCE" = "0" ]; then
    c_err "这台机器上似乎已经有 ${APP_NAME} 在运行（裸机或宝塔托管）。"
    c_err "继续安装会与之抢端口。确认要装请加 --force；只想看看会做什么请加 --dry-run。"
    exit 1
  fi
  c_warn "--force 已指定，继续（风险自负）"
else
  c_ok "目标端口 $APP_PORTS 全部空闲"
fi

# ============================================================
# 2. 中间件检测（只检测，不安装）
# ============================================================
step "2/9" "中间件检测（复用机器上已有的，不自动安装）"

if [ "$SKIP_INFRA" = "1" ]; then
  c_warn "--skip-infra-check 已指定，跳过"
else
  i=1; MISSING=""
  for p in $INFRA_PORTS; do
    name=$(echo "$INFRA_NAMES" | cut -d' ' -f"$i")
    if port_busy "$p"; then
      c_ok "$name (:$(printf '%-5s' "$p")) 可达"
    else
      c_warn "$name (:$(printf '%-5s' "$p")) 未检测到"
      MISSING="$MISSING $name(:$p)"
    fi
    i=$((i + 1))
  done

  if [ -n "$MISSING" ]; then
    cat <<EOF

  以下中间件缺失：$MISSING

  先装好（或改成指向已有实例）再重跑：
    PostgreSQL ≥ 16   apt install postgresql          (5432)
    Redis      ≥ 7    apt install redis-server        (6379)
    NATS              nats-server -js -sd /var/lib/nats/jetstream-standalone  (4223)
                      —— standalone 用独立实例，不与人共用 4222；必须带 -js
    MongoDB    ≥ 7    apt install mongodb-org         (27017)
    Milvus     ≥ 2.3  deploy/milvus/docker-compose.yml  (19530)
    MinIO             systemd 单机服务                  (9000)

  只想跳过检测继续（例如中间件在别的机器上）：
    bash scripts/install-systemd.sh --skip-infra-check

EOF
    [ "$FORCE" = "0" ] && exit 1
    c_warn "--force 已指定，忽略缺失继续"
  else
    c_ok "六件套全部可达"
  fi
fi

# ============================================================
# 3. 用户与目录
# ============================================================
step "3/9" "创建运行用户与目录"

if id "$RUN_USER" >/dev/null 2>&1; then
  c_ok "用户 $RUN_USER 已存在"
else
  run useradd -r -s /usr/sbin/nologin -d "$ROOT_DIR" "$RUN_USER"
  c_ok "已创建系统用户 $RUN_USER"
fi

run mkdir -p "$LOG_DIR" "$BUILD_DIR"
run chown -R "$RUN_USER:$RUN_USER" "$BUILD_DIR" 2>/dev/null || true
c_ok "日志目录: $LOG_DIR"

# ============================================================
# 4. .env
# ============================================================
step "4/9" "准备 .env"

ENV_FILE="$ROOT_DIR/.env"
ENV_TPL="$ROOT_DIR/.env.example"

if [ -f "$ENV_FILE" ] && [ "$FORCE_ENV" = "0" ]; then
  c_ok ".env 已存在，保留（要覆盖请加 --force-env）"
else
  [ -f "$ENV_TPL" ] || { c_err "缺少模板 $ENV_TPL，无法生成 .env"; exit 1; }
  run cp "$ENV_TPL" "$ENV_FILE"
  c_ok "已从 .env.example 生成 .env"

  if [ "$DRY_RUN" = "0" ]; then
    if [ -f "$ROOT_DIR/scripts/gen-env.sh" ]; then
      bash "$ROOT_DIR/scripts/gen-env.sh" "$ENV_FILE" >/dev/null 2>&1 \
        && c_ok "已填充随机密钥" || c_warn "gen-env.sh 执行失败，请手工检查 .env"
    else
      c_warn "未找到 scripts/gen-env.sh，请手工填写 .env 中的密钥"
    fi
  fi
fi

run chmod 600 "$ENV_FILE"
c_ok ".env 权限 600"

cat <<'EOF'

  还需人工确认 .env 里这几项：
    DB_*        指向实际的 PostgreSQL（默认 127.0.0.1:5432）
    REDIS_ADDR  指向实际的 Redis（默认 127.0.0.1:6379）
    NATS_URL    必须是 nats://127.0.0.1:4223（standalone 独立实例）
    QWEN_API_KEY / DEEPSEEK_API_KEY   要用 AI 对话必须填
    ALLOWED_ORIGINS                   前端域名

EOF

DB_NAME=$(grep -E '^DB_NAME=' "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- || echo "standalone")
DB_USER=$(grep -E '^DB_USER=' "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- || echo "standalone")
c_warn "请确认数据库 $DB_NAME 与角色 $DB_USER 已存在（表结构由服务启动时 AutoMigrate 自动创建）"
printf '     建库示例:\n'
printf '       sudo -u postgres psql -c "CREATE USER %s WITH PASSWORD '"'"'<取自 .env 的 DB_PASSWORD>'"'"';"\n' "$DB_USER"
printf '       sudo -u postgres psql -c "CREATE DATABASE %s OWNER %s;"\n' "$DB_NAME" "$DB_USER"

# ============================================================
# 5. 编译 Go 服务
# ============================================================
step "5/9" "编译后端（6 个 Go 服务）"

if [ "$DO_BUILD" = "0" ]; then
  c_warn "--no-build 已指定，跳过"
else
  run mkdir -p "$BUILD_DIR/_staging"
  for svc in $GO_SERVICES; do
    [ -d "$ROOT_DIR/backend/$svc" ] || { c_err "缺少源码目录 backend/$svc"; exit 1; }
    if [ "$DRY_RUN" = "1" ]; then
      printf '   \033[90m[dry-run]\033[0m go build backend/%s → build/%s\n' "$svc" "$svc"
    else
      ( cd "$ROOT_DIR/backend/$svc" && go build -o "$BUILD_DIR/_staging/$svc" . ) \
        || { c_err "编译失败: $svc"; exit 1; }
      install -o "$RUN_USER" -g "$RUN_USER" -m 755 "$BUILD_DIR/_staging/$svc" "$BUILD_DIR/$svc"
      c_ok "$svc"
    fi
  done

  for pl in wechat-plugin wecom-plugin feishu-plugin dingtalk-plugin douyin-plugin; do
    [ -d "$ROOT_DIR/plugins/$pl" ] || continue
    if [ "$DRY_RUN" = "1" ]; then
      printf '   \033[90m[dry-run]\033[0m go build plugins/%s\n' "$pl"
    else
      ( cd "$ROOT_DIR/plugins/$pl" && go build -o "$BUILD_DIR/_staging/$pl" . ) \
        && install -o "$RUN_USER" -g "$RUN_USER" -m 755 "$BUILD_DIR/_staging/$pl" "$BUILD_DIR/$pl" \
        && c_ok "$pl" || c_warn "$pl 编译失败，跳过（不影响主链路）"
    fi
  done
fi

# ============================================================
# 6. AI 引擎
# ============================================================
step "6/9" "准备 AI 引擎（ai-engine）"

AI_DIR="$ROOT_DIR/backend/ai-engine"
AI_VENV="$AI_DIR/venv"

if [ ! -d "$AI_DIR" ]; then
  c_warn "未找到 $AI_DIR，跳过"
elif [ "$DO_BUILD" = "0" ]; then
  c_warn "--no-build 已指定，跳过"
else
  if [ ! -d "$AI_VENV" ]; then
    run "$PY_BIN" -m venv "$AI_VENV"
    c_ok "已创建 venv: $AI_VENV"
  else
    c_ok "venv 已存在"
  fi

  if [ "$DRY_RUN" = "0" ]; then
    "$AI_VENV/bin/pip" install --upgrade pip -q
    if [ -f "$AI_DIR/requirements.txt" ]; then
      "$AI_VENV/bin/pip" install -r "$AI_DIR/requirements.txt" -q && c_ok "依赖已安装"
    elif [ -f "$AI_DIR/pyproject.toml" ]; then
      ( cd "$AI_DIR" && "$AI_VENV/bin/pip" install -e . -q ) && c_ok "依赖已安装"
    else
      c_warn "未找到 requirements.txt / pyproject.toml，请手动装依赖"
    fi
    if [ ! -f "$AI_DIR/.env" ] && [ -f "$AI_DIR/.env.example" ]; then
      cp "$AI_DIR/.env.example" "$AI_DIR/.env"
      c_ok "已从 .env.example 生成 ai-engine/.env"
    fi

    # ⚠️ ai-engine 的 Settings 是 extra=forbid（pydantic v2 默认行为）：
    #    .env 里出现模型未定义的变量会**直接启动失败**，日志报
    #    "ValidationError ... Extra inputs are not permitted"。
    #    最常见的就是把根 .env 的 JWT_SECRET 顺手同步了进来。
    if [ -f "$AI_DIR/.env" ]; then
      for bad in JWT_SECRET BILLING_INTERNAL_TOKEN AI_ENGINE_INTERNAL_TOKEN \
                 REDIS_ADDR ALLOWED_ORIGINS GATEWAY_PORT WEB_ADMIN_PORT JWT_EXPIRE_HOURS; do
        if grep -qE "^${bad}=" "$AI_DIR/.env"; then
          c_warn "ai-engine/.env 含有 $bad —— 不在 ai-engine 的 Settings 定义内，会导致启动失败"
          c_warn "  修法: sed -i '/^${bad}=/d' $AI_DIR/.env"
        fi
      done
    fi

    chown -R "$RUN_USER:$RUN_USER" "$AI_VENV" 2>/dev/null || true
  fi
fi

# ============================================================
# 7. 安装 systemd 单元
# ============================================================
step "7/9" "安装 systemd 单元"

[ -d "$UNIT_SRC" ] || { c_err "缺少单元模板目录 $UNIT_SRC"; exit 1; }

for f in "$UNIT_SRC"/*.service; do
  [ -f "$f" ] || continue
  base="$(basename "$f")"
  if [ "$DRY_RUN" = "1" ]; then
    printf '   \033[90m[dry-run]\033[0m install %s → %s\n' "$base" "$UNIT_DST/$base"
  else
    # 模板里的路径是默认安装路径，实际路径不同时替换成当前路径
    sed "s|/www/wwwroot/fettle-standalone|$ROOT_DIR|g" "$f" > "$UNIT_DST/$base"
    chmod 644 "$UNIT_DST/$base"
    c_ok "已安装 $base"
  fi
done

run systemctl daemon-reload

# ============================================================
# 8. 启动 + 健康检查
# ============================================================
step "8/9" "启动服务"

UNITS=""
for svc in $GO_SERVICES; do UNITS="$UNITS fettle-standalone-go@$svc.service"; done
UNITS="$UNITS fettle-standalone-ai-engine.service"

if [ "$DRY_RUN" = "1" ]; then
  printf '   \033[90m[dry-run]\033[0m systemctl enable --now%s\n' "$UNITS"
else
  # shellcheck disable=SC2086
  systemctl enable --now $UNITS
  c_ok "单元已 enable --now"

  printf '\n   等待服务就绪'
  for _ in $(seq 1 20); do
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
      printf '   \033[31m✗\033[0m :%-5s /health %s\n' "$p" "${code:-失败}"
      FAIL=1
    fi
  done

  if [ "$FAIL" = "1" ]; then
    c_warn "有服务未通过健康检查，排查："
    printf '     journalctl -u fettle-standalone-go@<服务名> -n 50 --no-pager\n'
    printf '     tail -50 %s/<服务名>.log\n' "$LOG_DIR"
  fi
fi

# ============================================================
# 9. 收尾
# ============================================================
step "9/9" "完成"

if [ "$WITH_FE" = "1" ] && [ "$DRY_RUN" = "0" ]; then
  if command -v npm >/dev/null 2>&1 && [ -d "$ROOT_DIR/frontend/web-admin" ]; then
    ( cd "$ROOT_DIR/frontend/web-admin" && npm ci --silent && npm run build ) \
      && c_ok "前端 web-admin 已构建" || c_warn "前端构建失败"
  fi
fi

cat <<EOF

  安装完成。下一步：

  1) 配置 nginx 反向代理
       admin.your-domain  →  http://127.0.0.1:20001
       api.your-domain    →  http://127.0.0.1:20001
       前端静态目录        →  $ROOT_DIR/frontend/web-admin/dist

  2) 常用运维命令
       查看状态:  systemctl status 'fettle-standalone-go@*'
       看日志:    journalctl -u fettle-standalone-go@gateway -f
       重启单个:  systemctl restart fettle-standalone-go@chat-service

  3) ⚠️ 不要和宝塔面板双头托管
       同一服务只能有一个权威。若面板里也登记了同名 Go 项目，
       必须把面板侧的「开机启动」关掉，否则两边会每 120 秒抢一次端口。

  4) 中途想看会做什么而不真执行
       bash scripts/install-systemd.sh --dry-run

EOF

c_ok "fettle-standalone 裸机安装流程结束"
