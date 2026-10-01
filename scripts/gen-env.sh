#!/usr/bin/env bash
# ============================================================
# fettle-standalone .env 随机密钥填充
#
# 用法:
#   cp .env.example .env && bash scripts/gen-env.sh [.env 路径]
#
# 行为:
#   * 只替换「空值」或「占位符」（change_me / your_xx / xxx / TODO）的密钥项
#   * 已有真实值的项一律保留 —— 可反复执行，不会把线上密钥洗掉
#   * 就地逐行改，保留原有注释与顺序
#
# 不处理的项（第三方凭据，随机生成没有意义，必须人工填）:
#   QWEN_API_KEY / DEEPSEEK_API_KEY / OPENAI_API_KEY
#   SMS_ACCESS_KEY_ID / SMS_ACCESS_KEY_SECRET
#   WECHAT_APP_ID / WECHAT_APP_SECRET
# ============================================================
set -euo pipefail

ENV_FILE="${1:-.env}"

if [ ! -f "$ENV_FILE" ]; then
  echo "[错误] 找不到 $ENV_FILE" >&2
  echo "       先执行: cp .env.example .env" >&2
  exit 1
fi

rand_hex() {
  local n="${1:-32}"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$n"
  else
    LC_ALL=C tr -dc 'a-f0-9' < /dev/urandom | head -c "$((n * 2))"
  fi
}

current_value() {
  grep -E "^$1=" "$ENV_FILE" 2>/dev/null | head -1 | cut -d= -f2- || true
}

is_placeholder() {
  case "${1:-}" in
    ""|change*|CHANGE*|your_*|YOUR_*|xxx*|XXX*|todo*|TODO*|placeholder*|示例*|你的*) return 0 ;;
    *) return 1 ;;
  esac
}

set_value() {
  local key="$1" value="$2"
  if grep -qE "^${key}=" "$ENV_FILE"; then
    sed -i "s|^${key}=.*$|${key}=${value}|" "$ENV_FILE"
  else
    printf '%s=%s\n' "$key" "$value" >> "$ENV_FILE"
  fi
}

fill() {  # $1=KEY  $2=随机字节数  $3=说明
  local key="$1" bytes="$2" desc="$3" cur
  cur="$(current_value "$key")"
  if is_placeholder "$cur"; then
    set_value "$key" "$(rand_hex "$bytes")"
    printf '  \033[32m生成\033[0m %-28s %s\n' "$key" "$desc"
  else
    printf '  \033[90m保留\033[0m %-28s 已有值\n' "$key"
  fi
}

echo "正在处理 $ENV_FILE"
echo
echo "— 应用密钥 —"
fill JWT_SECRET                  32 "JWT 签名密钥（≥32 字符）"
fill AI_ENGINE_INTERNAL_TOKEN    32 "网关/服务 → ai-engine 的内部鉴权令牌"
fill BILLING_INTERNAL_TOKEN      32 "billing/agent/skill 内部接口令牌"

echo
echo "— 数据库与存储 —"
fill DB_PASSWORD                 24 "PostgreSQL 口令"
fill REDIS_PASSWORD              24 "Redis 口令"
fill MINIO_ACCESS_KEY            16 "MinIO Access Key"
fill MINIO_SECRET_KEY            32 "MinIO Secret Key"

chmod 600 "$ENV_FILE"

cat <<'EOF'

完成。权限已设为 600。

还需人工填写（随机值无意义）：
  QWEN_API_KEY / DEEPSEEK_API_KEY      AI 对话功能
  ALLOWED_ORIGINS                      前端域名
  WECHAT_APP_ID / WECHAT_APP_SECRET    微信登录 / 公众号
  SMS_*                                短信验证码

注意：NATS_URL 必须是 nats://127.0.0.1:4223（standalone 用独立 NATS 实例）
EOF
