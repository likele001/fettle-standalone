#!/usr/bin/env bash
#
# 生成 .env 里的随机密钥，避免用示例值上线。
#
# 用法：
#   cp .env.example .env && ./gen-env.sh
#
# 行为：
#   - 只填充「空值」的密钥项，已有值的不覆盖（可反复执行）
#   - 生成后把 .env 权限设为 600
#
set -euo pipefail

ENV_FILE="${1:-.env}"

if [ ! -f "$ENV_FILE" ]; then
    echo "[ERROR] 找不到 $ENV_FILE" >&2
    echo "        先执行: cp .env.example .env" >&2
    exit 1
fi

# 生成随机串：优先 openssl，退回 /dev/urandom
rand_b64() {
    local n="${1:-32}"
    if command -v openssl >/dev/null 2>&1; then
        openssl rand -base64 "$n" | tr -d '\n=+/' | cut -c1-"$n"
    else
        LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c "$n"
    fi
}

# 若某个 KEY 当前是空值，则写入随机值
set_if_empty() {
    local key="$1" value="$2"
    if grep -qE "^${key}=$" "$ENV_FILE"; then
        # 用 | 作分隔符，避免值里的特殊字符出问题
        sed -i "s|^${key}=$|${key}=${value}|" "$ENV_FILE"
        echo "  [生成] ${key}"
    elif grep -qE "^${key}=" "$ENV_FILE"; then
        echo "  [保留] ${key}（已有值）"
    else
        printf '%s=%s\n' "$key" "$value" >> "$ENV_FILE"
        echo "  [追加] ${key}"
    fi
}

echo "正在处理 $ENV_FILE ..."

set_if_empty "JWT_SECRET"          "$(rand_b64 48)"
set_if_empty "POSTGRES_PASSWORD"   "$(rand_b64 24)"
set_if_empty "MINIO_ACCESS_KEY"    "$(rand_b64 20)"
set_if_empty "MINIO_SECRET_KEY"    "$(rand_b64 32)"

chmod 600 "$ENV_FILE"

echo
echo "完成。$ENV_FILE 权限已设为 600。"
echo
echo "还需手动填写的项（用到对应功能时）："
echo "  - QWEN_API_KEY / DEEPSEEK_API_KEY   （AI 对话功能）"
echo "  - WECHAT_APP_ID / WECHAT_APP_SECRET （微信登录/公众号）"
echo "  - ALLOWED_ORIGINS                   （前端域名，默认 localhost 仅本地可用）"
echo
echo "填完后启动：docker compose up -d"
