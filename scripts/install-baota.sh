#!/usr/bin/env bash
# ============================================================
# fettle-standalone 一键安装 · 宝塔面板方案
#
# 用法:
#   sudo bash scripts/install-baota.sh [选项]
#
# 选项:
#   --dry-run           只打印将要执行的动作，不改动任何文件/数据库
#   --force             端口被占用 / 面板已有同名项目时也继续（默认中止）
#   --skip-infra-check  跳过中间件连通性检测
#   --no-build          跳过 Go / ai-engine 构建，只注册项目并启动
#   --force-env         覆盖已存在的 .env（默认保留）
#   -h, --help          显示本帮助
#
# 设计说明:
#   本方案把 6 个 Go 服务 + 1 个 ai-engine 注册成**宝塔「Go 项目」/「Python 项目」**，
#   由宝塔面板负责启动与守护，与这台机器上现有的托管方式完全一致。
#
#   端口段 20001-20007、系统用户 fettle-std、NATS 4223 —— 与主平台 fettle 完全错开，
#   两套可以并存在同一台机器上（共用 PG/Redis/MinIO/MongoDB/Milvus）。
#
#   1. 注册方式是写面板的 site.db（sites 表），字段结构以宝塔 13.x 的
#      goModel.create_project / pythonModel.CreateProject 为准，已逐字段核对。
#   2. 同时复刻面板会生成的旁路产物（启动脚本 / pid 文件 / 日志 / python 项目 env），
#      这样面板**立刻就能认领**这些进程，不会出现「认不出 → 每 120 秒重复拉起」。
#   3. 中间件只检测**不安装**。
#   4. 安全阀：端口体检 + 面板同名项目检查，任一命中即中止。
#   5. 幂等：可反复执行；已存在的 .env 与项目默认不覆盖。
#
#   ⚠️ 本脚本会写入面板数据库。执行前会自动备份 site.db 到
#      /root/fettle-standalone-baota-install-<时间戳>/ ，出问题可原地回滚。
# ============================================================
set -euo pipefail

# ---------- 常量（本机既定值，勿随意改） ----------
APP_NAME="fettle-standalone"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUN_USER="${FTL_STD_RUN_USER:-fettle-std}"
BUILD_DIR="$ROOT_DIR/build"
GO_SERVICES="gateway user-service agent-service chat-service skill-service billing-service"
APP_PORTS="20001 20002 20003 20004 20005 20006 20007"
INFRA_PORTS="5432 6379 4223 27017 19530 9000"
INFRA_NAMES="PostgreSQL Redis NATS MongoDB Milvus MinIO"

# 面板侧约定路径（对齐宝塔 13.x 源码常量）
PANEL_DIR="/www/server/panel"
GO_PROJ_ROOT="/www/server/go_project"          # goModel._go_path
GO_RUN_SCRIPTS="$GO_PROJ_ROOT/vhost/scripts"   # goModel._go_run_scripts
GO_PID_DIR="/var/tmp/gopids"                   # goModel._go_pid_path
GO_LOG_DIR="/www/wwwlogs/go"                   # goModel._go_logs_path
PY_PROJ_ROOT="/www/server/python_project"      # pythonModel._project_path
PY_RUN_SCRIPTS="$PY_PROJ_ROOT/vhost/scripts"
PY_ENV_DIR="$PY_PROJ_ROOT/vhost/env"
PY_PID_DIR="$PY_PROJ_ROOT/vhost/pids"
PY_VENV_ROOT="/www/server/pyporject_evn"       # pythonModel._pyv_path
PY_LOG_ROOT="/www/wwwlogs/python"              # pythonModel._project_logs

# ai-engine（Python 项目）
AI_DIR="$ROOT_DIR/backend/ai-engine"
AI_PROJ_NAME="standaloneengine"
AI_VENV_NAME="fettlept"
AI_VENV="$PY_VENV_ROOT/$AI_VENV_NAME"
AI_PORT=20007

# Go 项目注册表： name|binary|port|extra_env(空格分隔 K=V)
# 注意：standalone 的端口变量名与主平台不同（GATEWAY_PORT / USER_SERVICE_PORT …），
#       不是统一的 APP_PORT —— 已对 backend/*/main.go 逐个核对。
GO_PROJECTS="
standalonegateway|gateway|20001|GATEWAY_PORT=20001
standaloneuserservice|user-service|20002|USER_SERVICE_PORT=20002
standaloneagentservice|agent-service|20003|AGENT_SERVICE_PORT=20003
standalonechatservice|chat-service|20004|CHAT_SERVICE_PORT=20004
standaloneskillservice|skill-service|20005|SKILL_SERVICE_PORT=20005
standalonebillingservice|billing-service|20006|BILLING_SERVICE_PORT=20006
"

# ---------- 选项 ----------
DRY_RUN=0; FORCE=0; SKIP_INFRA=0; DO_BUILD=1; FORCE_ENV=0
while [ $# -gt 0 ]; do
  case "$1" in
    --dry-run)          DRY_RUN=1 ;;
    --force)            FORCE=1 ;;
    --skip-infra-check) SKIP_INFRA=1 ;;
    --no-build)         DO_BUILD=0 ;;
    --force-env)        FORCE_ENV=1 ;;
    -h|--help)          sed -n '2,29p' "$0"; exit 0 ;;
    *) echo "未知参数: $1（用 --help 看用法）" >&2; exit 2 ;;
  esac
  shift
done

# ---------- 输出 helper ----------
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

port_busy() { ss -tlnH 2>/dev/null | grep -q ":$1\b"; }
port_owner() { ss -tlnpH 2>/dev/null | grep ":$1\b" | head -1 | sed -E 's/.*users:\(\("([^"]+)".*/\1/'; }

# 探测面板数据库位置（v11+ 在 data/db/，老版本在 data/）
detect_site_db() {
  for d in "$PANEL_DIR/data/db/site.db" "$PANEL_DIR/data/site.db" "$PANEL_DIR/data/default.db"; do
    if [ -s "$d" ] && sqlite3 "$d" "select count(*) from sites;" >/dev/null 2>&1; then
      echo "$d"; return 0
    fi
  done
  return 1
}

# ============================================================
step "0/9" "环境检查"
# ============================================================
if [ "$DRY_RUN" = "0" ] && [ "$(id -u)" != "0" ]; then
  c_err "需要 root 权限运行（要写面板数据库与 /www 下的目录）"
  exit 1
fi
c_ok "运行身份: $(id -un)"

if [ ! -d "$PANEL_DIR" ]; then
  c_err "未检测到宝塔面板（$PANEL_DIR 不存在）"
  c_err "本脚本只适用于宝塔环境；裸机部署请改用 scripts/install-systemd.sh"
  exit 1
fi
c_ok "宝塔面板: $(cat "$PANEL_DIR/data/version.pl" 2>/dev/null || echo '已安装')"

if ! command -v sqlite3 >/dev/null 2>&1; then
  c_err "缺少 sqlite3 命令（用于读写面板数据库）"
  c_err "安装: apt install -y sqlite3   /   yum install -y sqlite"
  exit 1
fi

SITE_DB="$(detect_site_db)" || { c_err "未能定位面板 site.db（数据库为空或无 sites 表）"; exit 1; }
c_ok "面板数据库: $SITE_DB"

if [ ! -d "$ROOT_DIR/backend" ]; then
  c_err "未在预期目录找到 backend/：$ROOT_DIR"
  exit 1
fi
c_ok "项目目录: $ROOT_DIR"

export PATH=/usr/local/go/bin:$PATH
if [ "$DO_BUILD" = "1" ]; then
  if ! command -v go >/dev/null 2>&1; then
    c_err "未找到 go —— 编译 ${APP_NAME} 后端需要 Go ≥ 1.22"
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

# ============================================================
step "1/9" "端口体检"
# ============================================================
BUSY=""
for p in $APP_PORTS; do port_busy "$p" && BUSY="$BUSY $p"; done

if [ -n "$BUSY" ]; then
  c_warn "以下目标端口已被占用：$BUSY"
  for p in $BUSY; do printf '     :%s  ← %s\n' "$p" "$(port_owner "$p")"; done
  if [ "$FORCE" = "0" ]; then
    c_err "这台机器上似乎已经有 ${APP_NAME} 在运行。"
    c_err "继续安装会与之抢端口。确认要装请加 --force；只想看看会做什么请加 --dry-run。"
    exit 1
  fi
  c_warn "--force 已指定，继续（风险自负）"
else
  c_ok "目标端口 $APP_PORTS 全部空闲"
fi

# ============================================================
step "2/9" "面板项目冲突检查"
# ============================================================
# 注意：用 here-doc（不是管道）读，while 才在**当前 shell** 里跑，变量能带出来
EXIST=""
EXIST_NAMES=""
while IFS= read -r line; do
  [ -z "$line" ] && continue
  EXIST_NAMES="$EXIST_NAMES $(printf '%s' "$line" | cut -d'|' -f1)"
done <<EOF
$GO_PROJECTS
EOF
EXIST_NAMES="$EXIST_NAMES $AI_PROJ_NAME"

for n in $EXIST_NAMES; do
  if sqlite3 "$SITE_DB" "select count(*) from sites where name='$n';" 2>/dev/null | grep -qv '^0$'; then
    EXIST="$EXIST $n"
  fi
done

if [ -n "$EXIST" ]; then
  c_warn "面板里已存在同名项目：$EXIST"
  if [ "$FORCE" = "0" ]; then
    c_err "重复注册会产生两条记录并互相抢端口。"
    c_err "若确认要覆盖配置，请加 --force（脚本会更新已有记录，不会新建重复项）。"
    exit 1
  fi
  c_warn "--force 已指定，将**更新**已有记录（不会新建重复项）"
else
  c_ok "面板中无同名项目，可安全注册"
fi

# ============================================================
step "3/9" "中间件检测（复用机器上已有的，不自动安装）"
# ============================================================
if [ "$SKIP_INFRA" = "1" ]; then
  c_warn "--skip-infra-check 已指定，跳过"
else
  i=1; MISSING=""
  for p in $INFRA_PORTS; do
    name=$(echo "$INFRA_NAMES" | cut -d' ' -f"$i")
    if port_busy "$p"; then c_ok "$name (:$(printf '%-5s' "$p")) 可达"
    else c_warn "$name (:$(printf '%-5s' "$p")) 未检测到"; MISSING="$MISSING $name(:$p)"; fi
    i=$((i + 1))
  done
  if [ -n "$MISSING" ]; then
    cat <<EOF

  以下中间件缺失：$MISSING

  请先装好（或改 .env 指向已有实例）再重跑：
    PostgreSQL ≥ 16   apt install postgresql        (5432)
    Redis      ≥ 7    apt install redis-server      (6379)
    NATS              nats-server -js -sd /var/lib/nats/jetstream   (4223)
                      —— 注意 standalone 用 4223，与主平台的 4222 是两个独立实例
    MongoDB    ≥ 7    apt install mongodb-org       (27017)
    Milvus     ≥ 2.3  官方 docker-compose 或单机二进制  (19530)   —— 与主平台共用
    MinIO             systemd 单机服务               (9000)      —— 与主平台共用

  只想跳过检测继续安装：
    bash scripts/install-baota.sh --skip-infra-check

EOF
    [ "$FORCE" = "0" ] && exit 1
    c_warn "--force 已指定，忽略缺失继续"
  else
    c_ok "六件套全部可达"
  fi
fi

# ============================================================
step "4/9" "创建运行用户与目录"
# ============================================================
if id "$RUN_USER" >/dev/null 2>&1; then
  c_ok "用户 $RUN_USER 已存在"
else
  run useradd -r -s /usr/sbin/nologin -d "$ROOT_DIR" "$RUN_USER"
  c_ok "已创建系统用户 $RUN_USER"
fi

run mkdir -p "$BUILD_DIR" "$GO_RUN_SCRIPTS" "$GO_PID_DIR" "$GO_LOG_DIR" \
             "$PY_RUN_SCRIPTS" "$PY_ENV_DIR" "$PY_PID_DIR" "$PY_LOG_ROOT/$AI_PROJ_NAME"
run chmod 777 "$GO_PID_DIR" 2>/dev/null || true
run chown -R "$RUN_USER:$RUN_USER" "$BUILD_DIR" "$GO_LOG_DIR" 2>/dev/null || true
c_ok "已创建面板侧目录（go_project / python_project / wwwlogs）"

# ============================================================
step "5/9" "准备 .env"
# ============================================================
ENV_FILE="$ROOT_DIR/.env"
ENV_TPL="$ROOT_DIR/.env.example"

if [ -f "$ENV_FILE" ] && [ "$FORCE_ENV" = "0" ]; then
  c_ok ".env 已存在，保留（要覆盖请加 --force-env）"
else
  if [ ! -f "$ENV_TPL" ]; then
    c_err "缺少模板 $ENV_TPL，无法生成 .env"
    exit 1
  fi
  run cp "$ENV_TPL" "$ENV_FILE"
  c_ok "已从 .env.example 生成 .env"
  if [ "$DRY_RUN" = "0" ] && [ -f "$ROOT_DIR/scripts/gen-env.sh" ]; then
    bash "$ROOT_DIR/scripts/gen-env.sh" "$ENV_FILE" >/dev/null 2>&1 \
      && c_ok "已填充随机密钥（JWT_SECRET / 内部令牌 / DB / MinIO）" \
      || c_warn "gen-env.sh 执行失败，请手动检查 .env 中的占位符"
  fi
fi
run chmod 600 "$ENV_FILE"
c_ok ".env 权限 600"

cat <<'EOF'

  还需人工确认 .env 里这几项（缺失会导致对应功能不可用）：
    DB_*        指向 PostgreSQL（本机与主平台共用 127.0.0.1:5432，库名 standalone）
    REDIS_ADDR  指向 Redis（127.0.0.1:6379，REDIS_DB 与主平台错开）
    NATS_URL    必须是 nats://127.0.0.1:4223（standalone 独立实例，不是 4222）
    MILVUS_HOST 若是单机部署应填 127.0.0.1 —— 填容器服务名在宿主机上解析不到
    QWEN_API_KEY / DEEPSEEK_API_KEY   要用 AI 对话必须填

EOF

DB_NAME=$(grep -E '^DB_NAME=' "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- || echo "standalone")
DB_USER=$(grep -E '^DB_USER=' "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- || echo "standalone")
c_warn "请确认数据库 $DB_NAME 与角色 $DB_USER 已存在（表结构由服务启动时 AutoMigrate 自动创建）"
printf '       sudo -u postgres psql -c "CREATE USER %s WITH PASSWORD '"'"'<取自 .env 的 DB_PASSWORD>'"'"';"\n' "$DB_USER"
printf '       sudo -u postgres psql -c "CREATE DATABASE %s OWNER %s;"\n' "$DB_NAME" "$DB_USER"

# ============================================================
step "6/9" "编译后端"
# ============================================================
if [ "$DO_BUILD" = "0" ]; then
  c_warn "--no-build 已指定，跳过"
else
  run mkdir -p "$BUILD_DIR/_staging"
  for svc in $GO_SERVICES; do
    if [ ! -d "$ROOT_DIR/backend/$svc" ]; then
      c_err "缺少源码目录 backend/$svc"; exit 1
    fi
    if [ "$DRY_RUN" = "1" ]; then
      printf '   \033[90m[dry-run]\033[0m go build backend/%s → build/%s\n' "$svc" "$svc"
    else
      # 先落暂存再 install —— 运行中的可执行文件不能直接覆盖（ETXTBSY）
      ( cd "$ROOT_DIR/backend/$svc" && go build -o "$BUILD_DIR/_staging/$svc" . ) \
        || { c_err "编译失败: $svc"; exit 1; }
      install -o "$RUN_USER" -g "$RUN_USER" -m 755 "$BUILD_DIR/_staging/$svc" "$BUILD_DIR/$svc"
      c_ok "$svc"
    fi
  done
fi

# ai-engine：venv 建在**面板约定的位置**，这样面板能直接复用
if [ ! -d "$AI_DIR" ]; then
  c_warn "未找到 $AI_DIR，跳过 ai-engine"
elif [ "$DO_BUILD" = "0" ]; then
  c_warn "--no-build 已指定，跳过 ai-engine"
else
  if [ ! -d "$AI_VENV" ]; then
    run mkdir -p "$PY_VENV_ROOT"
    run "$PY_BIN" -m venv "$AI_VENV"
    c_ok "已创建 venv: $AI_VENV（面板约定位置）"
  else
    c_ok "venv 已存在: $AI_VENV"
  fi

  if [ "$DRY_RUN" = "0" ]; then
    "$AI_VENV/bin/pip" install --upgrade pip -q
    if [ -f "$AI_DIR/requirements.txt" ]; then
      "$AI_VENV/bin/pip" install -r "$AI_DIR/requirements.txt" -q && c_ok "ai-engine 依赖已安装"
    else
      c_warn "未找到 requirements.txt，请手动装依赖"
    fi
  fi

  # ai-engine 自己的 .env
  if [ ! -f "$AI_DIR/.env" ] && [ -f "$AI_DIR/.env.example" ]; then
    run cp "$AI_DIR/.env.example" "$AI_DIR/.env"
    c_ok "已从 .env.example 生成 ai-engine/.env"
  fi

  # ai-engine 的 Settings 是 pydantic BaseSettings，model_config.extra 决定多出来的变量怎么办：
  #   extra="forbid" → .env 里多一个未定义变量就**拒绝启动**
  #                    （这台机器 2026-10-01 真实踩过：standaloneengine 报 extra_forbidden 148 次）
  #   extra="ignore" → 多出来的变量被静默忽略，无害（但属冗余）
  # 所以这里按 settings.py 的实际配置定级，并从 Settings 定义**动态推导**合法变量名 ——
  # 不硬编码黑名单，否则会把 ai_engine_internal_token 这类合法字段误报成错误。
  if [ -f "$AI_DIR/.env" ] && [ -f "$AI_DIR/config/settings.py" ]; then
    "$PY_BIN" - "$AI_DIR" <<'PYCHK'
import os, re, sys
ai_dir = sys.argv[1]
envp = os.path.join(ai_dir, ".env")
src = open(os.path.join(ai_dir, "config", "settings.py"), encoding="utf-8").read()

m = re.search(r'"extra"\s*:\s*"([a-z]+)"', src)
extra = m.group(1) if m else "unknown"
# pydantic 字段形如「4 空格缩进 + 小写名 + 类型注解」
allowed = set(re.findall(r'^\s{4}([a-z_][a-z0-9_]*)\s*:\s*', src, re.M))

keys = []
for ln in open(envp, encoding="utf-8"):
    ln = ln.strip()
    if ln and not ln.startswith("#") and "=" in ln:
        keys.append(ln.split("=", 1)[0].strip())

bad = [k for k in keys if k.lower() not in allowed]
if not bad:
    print("   \033[32m✓\033[0m ai-engine/.env 的 %d 个变量都在 Settings 定义内" % len(keys))
elif extra == "forbid":
    print("   \033[31m✗\033[0m ai-engine/.env 有 %d 个变量不在 Settings 定义内，"
          "且 extra=forbid → 服务会启动失败：" % len(bad))
    print("       " + ", ".join(bad))
    print("       删掉: sed -i '/^(%s)=/d' %s" % ("|".join(bad), envp))
else:
    print("   \033[33m!\033[0m ai-engine/.env 有 %d 个冗余变量（extra=%s，会被忽略、无害）："
          % (len(bad), extra))
    print("       " + ", ".join(bad))
PYCHK
  fi

  run chown -R "$RUN_USER:$RUN_USER" "$AI_VENV" 2>/dev/null || true
fi

# ============================================================
step "7/9" "注册宝塔项目（写面板数据库）"
# ============================================================
BACKUP_DIR="/root/${APP_NAME}-baota-install-$(date +%Y%m%d%H%M%S)"

if [ "$DRY_RUN" = "1" ]; then
  printf '   \033[90m[dry-run]\033[0m 备份 %s → %s/site.db\n' "$SITE_DB" "$BACKUP_DIR"
  printf '   \033[90m[dry-run]\033[0m 注册 %s 个 Go 项目 + Python 项目 %s\n' "$(printf '%s' "$GO_PROJECTS" | grep -c '|')" "$AI_PROJ_NAME"
else
  mkdir -p "$BACKUP_DIR"
  cp -a "$SITE_DB" "$BACKUP_DIR/site.db"
  c_ok "已备份面板数据库 → $BACKUP_DIR/site.db"

  if id "$RUN_USER" >/dev/null 2>&1; then :; else c_err "用户 $RUN_USER 不存在"; exit 1; fi

  # 面板的 Python 项目记录里有一项 version（形如 "Python 3.13.13"），取 venv 的实际版本
  PY_VER=""
  if [ -x "$AI_VENV/bin/python" ]; then PY_VER="$("$AI_VENV/bin/python" -V 2>&1)"
  elif command -v "$PY_BIN" >/dev/null 2>&1; then PY_VER="$("$PY_BIN" -V 2>&1)"; fi

  # ---- 用 python 参数化写库（避免 shell 拼 JSON 的引号地狱） ----
  FTL_ROOT="$ROOT_DIR" FTL_USER="$RUN_USER" FTL_BUILD="$BUILD_DIR" FTL_DB="$SITE_DB" \
  FTL_GO_PROJECTS="$GO_PROJECTS" FTL_AI_NAME="$AI_PROJ_NAME" FTL_AI_DIR="$AI_DIR" \
  FTL_AI_VENV="$AI_VENV" FTL_AI_PORT="$AI_PORT" FTL_GO_LOG="$GO_LOG_DIR" \
  FTL_PY_LOG_ROOT="$PY_LOG_ROOT" FTL_PY_VENV_ROOT="$PY_VENV_ROOT" FTL_PY_VER="$PY_VER" \
  "$PY_BIN" - <<'PY'
import os, json, sqlite3, datetime, posixpath

root       = os.environ["FTL_ROOT"]
run_user   = os.environ["FTL_USER"]
build_dir  = os.environ["FTL_BUILD"]
db_path    = os.environ["FTL_DB"]
go_projects= [l for l in os.environ["FTL_GO_PROJECTS"].strip().splitlines() if l.strip()]
ai_name    = os.environ["FTL_AI_NAME"]
ai_dir     = os.environ["FTL_AI_DIR"]
ai_venv    = os.environ["FTL_AI_VENV"]
ai_port    = os.environ["FTL_AI_PORT"]
go_log_dir = os.environ["FTL_GO_LOG"]
py_log_root= os.environ["FTL_PY_LOG_ROOT"]

con = sqlite3.connect(db_path)
cur = con.cursor()
now = datetime.datetime.now().strftime("%Y-%m-%d %H:%M:%S")

# ---------- Go 项目 ----------
for line in go_projects:
    name, binary, port, extra = (line.split("|") + ["", "", "", ""])[:4]
    exe = posixpath.join(build_dir, binary)
    env_list = [{"k": kv.split("=")[0], "v": kv.split("=", 1)[1]}
                for kv in extra.split() if "=" in kv]
    cfg = {
        "ssl_path": "/www/wwwroot/java_node_ssl",
        "project_name": name,
        "project_exe": exe,
        "bind_extranet": 0,
        "domains": [],
        "project_cmd": exe,
        "is_power_on": True,
        "run_user": run_user,
        "port": int(port),
        "project_path": posixpath.join(root, "backend", binary),
        "log_path": go_log_dir,
        "porject_log": 1,     # 面板原字段名就是拼错的 porject_log，照写
        "web_log": 1,
        "env_file": "",
        "env_list": env_list,
        "project_log": 1,
    }
    if cur.execute("select count(*) from sites where name=?", (name,)).fetchone()[0]:
        cur.execute("""update sites set path=?, ps=?, status=?, project_type=?,
                       project_config=?, addtime=? where name=?""",
                    (exe, name, "1", "Go", json.dumps(cfg), now, name))
        print(f"  更新 Go 项目 {name}  :{port}")
    else:
        cur.execute("""insert into sites (name,path,ps,status,type_id,project_type,
                       project_config,addtime) values (?,?,?,?,?,?,?,?)""",
                    (name, exe, name, "1", 0, "Go", json.dumps(cfg), now))
        print(f"  注册 Go 项目 {name}  :{port}")

# ---------- Python 项目 (ai-engine) ----------
py_bin = posixpath.join(ai_venv, "bin", "python")
py_cfg = {
    "pjname": ai_name,
    "port": "",
    "stype": "command",
    "path": ai_dir,
    "user": "www",
    "requirement_path": posixpath.join(ai_dir, "requirements.txt"),
    "env_list": [],
    "env_file": posixpath.join(ai_dir, ".env"),
    "framework": "fastapi",
    "vpath": ai_venv,
    "version": os.environ.get("FTL_PY_VER", ""),
    "python_bin": py_bin,
    "project_cmd": f"uvicorn main:app --host 0.0.0.0 --port {ai_port}",
    "xsgi": "wsgi",
    "rfile": "",
    "call_app": "app",
    "auto_run": True,
    "logpath": posixpath.join(py_log_root, ai_name),
    "is_pypy": False,
    "initialize": "",
    "domains": [],
    "bind_extranet": 0,
    "processes": 4,
    "threads": 2,
    "loglevel": "info",
    "is_http": "is_http",
    "start_sh": f"uvicorn main:app --host 0.0.0.0 --port {ai_port}",
}
if cur.execute("select count(*) from sites where name=?", (ai_name,)).fetchone()[0]:
    cur.execute("""update sites set path=?, ps=?, status=?, project_type=?,
                   project_config=?, addtime=? where name=?""",
                (ai_dir, ai_name, "1", "Python", json.dumps(py_cfg), now, ai_name))
    print(f"  更新 Python 项目 {ai_name}  :{ai_port}")
else:
    cur.execute("""insert into sites (name,path,ps,status,type_id,project_type,
                   project_config,addtime) values (?,?,?,?,?,?,?,?)""",
                (ai_name, ai_dir, ai_name, "1", 0, "Python", json.dumps(py_cfg), now))
    print(f"  注册 Python 项目 {ai_name}  :{ai_port}")

con.commit()
con.close()
print("  面板数据库写入完成")
PY
  c_ok "项目已注册到面板"
fi

# ---- 生成面板侧旁路产物：让面板**立刻认领**这些进程 ----
# Go：启动脚本 + 日志 + pid（复刻 goModel.start_project 的产物结构）
if [ "$DRY_RUN" = "1" ]; then
  printf '   \033[90m[dry-run]\033[0m 生成 Go 启动脚本 → %s/<项目名>.sh\n' "$GO_RUN_SCRIPTS"
  printf '   \033[90m[dry-run]\033[0m 生成 Python 启动脚本 → %s/%s_cmd.sh\n' "$PY_RUN_SCRIPTS" "$AI_PROJ_NAME"
else
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    name=$(printf '%s' "$line" | cut -d'|' -f1)
    binary=$(printf '%s' "$line" | cut -d'|' -f2)
    extra=$(printf '%s' "$line" | cut -d'|' -f4)

    pre_sh=""
    for kv in $extra; do pre_sh="$pre_sh
export $kv"; done

    cat > "$GO_RUN_SCRIPTS/$name.sh" <<GOEOF
#!/bin/bash
PATH=/usr/local/btgo/bin:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin:\$PATH
export PATH
cd $ROOT_DIR/backend/$binary
$pre_sh
nohup $BUILD_DIR/$binary &>> $GO_LOG_DIR/$name.log &
echo \$! > $GO_PID_DIR/$name.pid
GOEOF
    chmod 755 "$GO_RUN_SCRIPTS/$name.sh"
    chown "$RUN_USER:$RUN_USER" "$GO_RUN_SCRIPTS/$name.sh" 2>/dev/null || true
    touch "$GO_LOG_DIR/$name.log"
    chown "$RUN_USER:$RUN_USER" "$GO_LOG_DIR/$name.log" 2>/dev/null || true
  done <<EOF
$GO_PROJECTS
EOF
  c_ok "已生成 6 个 Go 启动脚本与日志文件"

  # Python：cmd.sh + env（复刻 pythonModel.__prepare_cmd_start_conf 的产物）
  SID=$(printf '%s' "$AI_PROJ_NAME" | md5sum | cut -d' ' -f1)   # 面板用 md5(项目名) 作 service sid
  echo "source $AI_DIR/.env" > "$PY_ENV_DIR/$AI_PROJ_NAME.env"

  cat > "$PY_RUN_SCRIPTS/${AI_PROJ_NAME}_cmd.sh" <<PYEOF
#!/bin/bash
PATH=$AI_VENV/bin:$AI_DIR:/bin:/sbin:/usr/bin:/usr/sbin:/usr/local/bin:/usr/local/sbin:~/bin
export PATH
export BT_PYTHON_SERVICE_SID=$SID
source $AI_VENV/bin/activate

source $PY_ENV_DIR/$AI_PROJ_NAME.env
cd $AI_DIR
nohup uvicorn main:app --host 0.0.0.0 --port $AI_PORT &>> $PY_LOG_ROOT/$AI_PROJ_NAME/error.log &
echo \$! > $PY_PID_DIR/$AI_PROJ_NAME.pid
PYEOF
  chmod 755 "$PY_RUN_SCRIPTS/${AI_PROJ_NAME}_cmd.sh"
  c_ok "已生成 Python 启动脚本与 env"

  # python_project_name2env.txt：面板维护的「项目名 → venv 路径」映射，必须同步
  NAME2ENV="$PANEL_DIR/data/python_project_name2env.txt"
  if [ -f "$NAME2ENV" ]; then
    if grep -q "^${AI_PROJ_NAME}:" "$NAME2ENV"; then
      sed -i "s|^${AI_PROJ_NAME}:.*|${AI_PROJ_NAME}:${AI_VENV}|" "$NAME2ENV"
    else
      printf '%s:%s\n' "$AI_PROJ_NAME" "$AI_VENV" >> "$NAME2ENV"
    fi
    c_ok "已同步 panel python_project_name2env.txt"
  else
    c_warn "未找到 $NAME2ENV（面板可能尚未初始化 Python 项目模块），已跳过"
  fi

  # 日志目录归属
  chown -R "$RUN_USER:$RUN_USER" "$PY_LOG_ROOT/$AI_PROJ_NAME" 2>/dev/null || true
  chmod -R 755 "$PY_LOG_ROOT/$AI_PROJ_NAME" 2>/dev/null || true
fi

# ============================================================
step "8/9" "启动服务并健康检查"
# ============================================================
if [ "$DRY_RUN" = "1" ]; then
  printf '   \033[90m[dry-run]\033[0m bash %s/<项目名>.sh   （依次启动 6 个 Go 服务）\n' "$GO_RUN_SCRIPTS"
  printf '   \033[90m[dry-run]\033[0m bash %s/%s_cmd.sh\n' "$PY_RUN_SCRIPTS" "$AI_PROJ_NAME"
else
  FAILED_START=""
  while IFS= read -r line; do
    [ -z "$line" ] && continue
    name=$(printf '%s' "$line" | cut -d'|' -f1)
    if su -s /bin/bash -c "bash $GO_RUN_SCRIPTS/$name.sh" "$RUN_USER" >/dev/null 2>&1; then
      printf '   \033[32m✓\033[0m 已启动 %s\n' "$name"
    else
      printf '   \033[31m✗\033[0m 启动失败 %s\n' "$name"
      FAILED_START="$FAILED_START $name"
    fi
  done <<EOF
$GO_PROJECTS
EOF
  if [ -n "$FAILED_START" ]; then
    c_warn "以下服务启动失败：$FAILED_START"
  fi

  # ai-engine 以 www 身份启动（与面板 Python 项目默认 user=www 一致）
  if su -s /bin/bash -c "bash $PY_RUN_SCRIPTS/${AI_PROJ_NAME}_cmd.sh" www >/dev/null 2>&1; then
    printf '   \033[32m✓\033[0m 已启动 %s\n' "$AI_PROJ_NAME"
  else
    printf '   \033[31m✗\033[0m 启动失败 %s\n' "$AI_PROJ_NAME"
  fi

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
    printf '     tail -50 %s/<项目名>.log\n' "$GO_LOG_DIR"
    printf '     tail -50 %s/%s/error.log\n' "$PY_LOG_ROOT" "$AI_PROJ_NAME"
    printf '     面板 → 项目 → 对应项目 → 日志\n'
  fi
fi

# ============================================================
step "9/9" "完成"
# ============================================================
cat <<EOF

  安装完成。现在可以回到**宝塔面板 → 项目**查看：

     Go 项目      standalonegateway / standaloneuserservice / standaloneagentservice /
                  standalonechatservice / standaloneskillservice / standalonebillingservice
     Python 项目  standaloneengine

  常用操作：
     面板里重启:   项目 → 选中 → 重启
     看实时日志:   tail -f $GO_LOG_DIR/standalonegateway.log
                   tail -f $PY_LOG_ROOT/$AI_PROJ_NAME/error.log
     面板外启动:   bash $GO_RUN_SCRIPTS/standalonegateway.sh
     面板外停止:   kill \$(cat $GO_PID_DIR/standalonegateway.pid)

  nginx 反向代理（面板 → 网站 → 添加站点后配反向代理即可）：
       standalone.your-domain  →  http://127.0.0.1:20001
       前端静态目录             →  $ROOT_DIR/frontend/web-admin/dist

  ⚠️ 三个必须知道的坑
     1. 不要再给同一批端口装 systemd 单元（scripts/install-systemd.sh）——
        这台机器上 standalone 的 17 个 systemd 单元**全是 inactive 且两套命名并存**
        （fettle-standalone-agent.service / fettle-standalone-agent-service.service）。
        一旦 enable，会立刻和面板进程抢 20001-20006，日志被重复启动刷屏。
     2. 面板「Go 项目」的 env_list 优先级**高于** .env
        （godotenv 不覆盖已存在的环境变量）。若以后改了 .env 不生效，
        先到面板项目的「环境变量」里看有没有同名项把它压住了。
     3. ai-engine 的 pydantic Settings：model_config 里 extra="forbid" 时，.env 多一个
        它不认识的变量就**直接启动失败**（本机 2026-10-01 真实踩过，已把 extra 改成 ignore）。
        脚本安装前会自动核对 .env 变量与 Settings 定义。

  回滚（万一要撤销本次改动）：
     cp $BACKUP_DIR/site.db $SITE_DB
     面板 → 项目 → 删除对应项目

EOF

c_ok "${APP_NAME} 宝塔方案安装流程结束"
