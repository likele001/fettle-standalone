#!/usr/bin/env bash
# ============================================================================
# gen-secret.sh —— 从环境变量生成 K8s Secret 清单
# ----------------------------------------------------------------------------
# 用途：避免把明文凭据写进 Git。脚本读取环境变量，输出可直接 apply 的
#       Secret YAML 到 stdout。
#
# 用法：
#   cd deploy/k8s
#   export DB_PASSWORD="$(openssl rand -hex 24)"
#   export JWT_SECRET="$(openssl rand -hex 32)"
#   export AI_ENGINE_INTERNAL_TOKEN="$(openssl rand -hex 32)"
#   export BILLING_INTERNAL_TOKEN="$(openssl rand -hex 32)"
#   export MINIO_ACCESS_KEY="fettle"
#   export MINIO_SECRET_KEY="$(openssl rand -hex 24)"
#   export CHANNEL_SECRET="$(openssl rand -hex 24)"
#   ./gen-secret.sh | kubectl apply -f -
#
# 或先生成到文件人工检查：
#   ./gen-secret.sh > 01-secret.yaml && kubectl apply -f 01-secret.yaml
#
# 重复执行即为原地更新（幂等），改配置后重新跑一次即可。
# ============================================================================
set -euo pipefail

NAMESPACE="${NAMESPACE:-fettle}"
SECRET_NAME="${SECRET_NAME:-fettle-secret}"

# ---- 必填项校验：缺失直接退出，不生成半成品 Secret ----
required=(DB_PASSWORD JWT_SECRET AI_ENGINE_INTERNAL_TOKEN BILLING_INTERNAL_TOKEN)
missing=()
for v in "${required[@]}"; do
  if [[ -z "${!v:-}" ]]; then
    missing+=("$v")
  fi
done

if (( ${#missing[@]} > 0 )); then
  cat >&2 <<EOF
[ERROR] 以下必填环境变量未设置：
$(printf '        - %s\n' "${missing[@]}")

快速生成（复制粘贴即可）：
  export DB_PASSWORD="\$(openssl rand -hex 24)"
  export JWT_SECRET="\$(openssl rand -hex 32)"
  export AI_ENGINE_INTERNAL_TOKEN="\$(openssl rand -hex 32)"
  export BILLING_INTERNAL_TOKEN="\$(openssl rand -hex 32)"

生成后重新执行本脚本。
EOF
  exit 1
fi

# ---- 弱值检查：拦住明显的占位符与出厂默认 ----
weak_check() {
  local name="$1" val="$2"
  case "$val" in
    CHANGE_ME|CHANGE_ME_use_env_var|change-me|change-me-in-production|\
    your-password|your-jwt-secret-key|minioadmin|password|admin|123456|__REQUIRED__)
      echo "[ERROR] $name 的值是占位符/出厂默认值（$val），请替换为随机值。" >&2
      exit 1
      ;;
  esac
  if (( ${#val} < 16 )); then
    echo "[WARN] $name 长度仅 ${#val}，建议至少 16 字符。" >&2
  fi
}
for v in "${required[@]}"; do weak_check "$v" "${!v}"; done
[[ -n "${MINIO_SECRET_KEY:-}" ]] && weak_check MINIO_SECRET_KEY "$MINIO_SECRET_KEY"
[[ -n "${CHANNEL_SECRET:-}"   ]] && weak_check CHANNEL_SECRET   "$CHANNEL_SECRET"

# ---- 输出 ----
cat <<EOF
# 由 gen-secret.sh 生成于 $(date -u +%Y-%m-%dT%H:%M:%SZ)
# 请勿提交到 Git（deploy/k8s/.gitignore 已忽略 *.generated.yaml）
apiVersion: v1
kind: Secret
metadata:
  name: ${SECRET_NAME}
  namespace: ${NAMESPACE}
  labels:
    app.kubernetes.io/name: fettle
    app.kubernetes.io/part-of: fettle
type: Opaque
stringData:
  DB_PASSWORD: $(printf '%q' "${DB_PASSWORD}" | sed "s/^/'/;s/\$/'/")
  JWT_SECRET: $(printf '%q' "${JWT_SECRET}" | sed "s/^/'/;s/\$/'/")
  AI_ENGINE_INTERNAL_TOKEN: $(printf '%q' "${AI_ENGINE_INTERNAL_TOKEN}" | sed "s/^/'/;s/\$/'/")
  BILLING_INTERNAL_TOKEN: $(printf '%q' "${BILLING_INTERNAL_TOKEN}" | sed "s/^/'/;s/\$/'/")
  REDIS_PASSWORD: '${REDIS_PASSWORD:-}'
  MINIO_ACCESS_KEY: '${MINIO_ACCESS_KEY:-}'
  MINIO_SECRET_KEY: '${MINIO_SECRET_KEY:-}'
  MONGODB_PASSWORD: '${MONGODB_PASSWORD:-}'
  NATS_PASSWORD: '${NATS_PASSWORD:-}'
  CHANNEL_SECRET: '${CHANNEL_SECRET:-}'
EOF

echo "[OK] Secret ${NAMESPACE}/${SECRET_NAME} 已生成" >&2
