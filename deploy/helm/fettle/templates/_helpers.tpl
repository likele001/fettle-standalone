{{/*
============================================================================
通用模板助手
============================================================================
*/}}

{{- define "fettle.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fettle.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "fettle.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fettle.labels" -}}
helm.sh/chart: {{ include "fettle.chart" . }}
{{ include "fettle.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/part-of: fettle
{{- end }}

{{/*
⚠️ selectorLabels 必须保持稳定：改动它会导致 Deployment 的 selector
   不可变字段变更，升级时 kubectl 会报 "field is immutable"。
   因此这里只用 app.kubernetes.io/name + instance，不含版本。
*/}}
{{- define "fettle.selectorLabels" -}}
app.kubernetes.io/name: {{ include "fettle.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "fettle.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "fettle.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
完整镜像地址：registry 为空时退化为 "image:tag"（不带前导斜杠）。
⚠️ 原 chart 用的是 "{{ .Values.global.imageRegistry }}/fettle-{{ $name }}:tag"，
   registry 留空时会渲染成 "/fettle-gateway:latest"，这是非法镜像名。
*/}}
{{- define "fettle.image" -}}
{{- $registry := .registry | default "" -}}
{{- $repo := .repository -}}
{{- $tag := .tag | default "latest" -}}
{{- if $registry -}}
{{- printf "%s/%s:%s" $registry $repo $tag -}}
{{- else -}}
{{- printf "%s:%s" $repo $tag -}}
{{- end -}}
{{- end }}

{{/*
Secret 名称解析：优先用外部 Secret，否则用 chart 自建的那个。
*/}}
{{- define "fettle.secretName" -}}
{{- if and .Values.secrets.externalSecret.enabled .Values.secrets.externalSecret.name -}}
{{- .Values.secrets.externalSecret.name -}}
{{- else -}}
{{- printf "%s-secret" (include "fettle.fullname" .) -}}
{{- end -}}
{{- end }}

{{/*
公共环境变量（所有后端服务共用）。
入参 dict: root(.) + extraEnv(list)
*/}}
{{- define "fettle.commonEnv" -}}
{{- $root := .root -}}
- name: APP_ENV
  value: "production"
- name: GIN_MODE
  value: "release"
- name: LOG_LEVEL
  value: "info"
- name: ALLOWED_ORIGINS
  value: {{ $root.Values.global.allowedOrigins | quote }}
- name: JWT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "fettle.secretName" $root }}
      key: jwt-secret
- name: DB_HOST
  value: {{ $root.Values.infra.postgres.host | quote }}
- name: DB_PORT
  value: {{ $root.Values.infra.postgres.port | quote }}
- name: DB_USER
  value: {{ $root.Values.infra.postgres.user | quote }}
- name: DB_NAME
  value: {{ $root.Values.infra.postgres.database | quote }}
- name: DB_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "fettle.secretName" $root }}
      key: db-password
- name: REDIS_ADDR
  value: {{ printf "%s:%v" $root.Values.infra.redis.host $root.Values.infra.redis.port | quote }}
- name: REDIS_PASSWORD
  valueFrom:
    secretKeyRef:
      name: {{ include "fettle.secretName" $root }}
      key: redis-password
- name: NATS_URL
  value: {{ printf "nats://%s:%v" $root.Values.infra.nats.host $root.Values.infra.nats.port | quote }}
- name: MONGODB_URI
  value: {{ printf "mongodb://%s:%v" $root.Values.infra.mongodb.host $root.Values.infra.mongodb.port | quote }}
- name: MONGODB_DB
  value: {{ $root.Values.infra.mongodb.database | quote }}
- name: AI_ENGINE_ADDR
  value: {{ printf "http://%s:%v" $root.Values.services.aiEngine.image $root.Values.services.aiEngine.port | quote }}
- name: AI_ENGINE_INTERNAL_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{ include "fettle.secretName" $root }}
      key: ai-engine-internal-token
{{- end }}
