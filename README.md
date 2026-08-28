# fettle-standalone

<p align="center">
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-AGPL--3.0-blue" alt="License" /></a>
</p>

`fettle-standalone` 是一套面向私有部署的 AI 智能体平台，基于多租户版本改造而来，移除云端多租户功能与复杂 SaaS 特性，保留对话、智能体、技能、知识库和 AI 模型配置等核心功能。

本仓库适合希望在单机或自托管环境中运行完整 AI 平台的用户。

---

## 目录

- [特性](#特性)
- [架构概览](#架构概览)
- [快速启动](#快速启动)
- [部署方式](#部署方式)
- [环境依赖](#环境依赖)
- [环境变量配置](#环境变量配置)
- [服务端口](#服务端口)
- [文档与示例](#文档与示例)
- [项目结构](#项目结构)
- [贡献](#贡献)

---

## 特性

- 单机私有部署，适合内部环境和小型项目
- 保留智能体管理、对话服务、技能市场、计费统计、RAG 知识库等功能
- 支持通义千问、DeepSeek、OpenAI 等模型接入
- 统一 API 网关，前端管理后台、平台前端、渠道接入支持
- 提供原生部署与 Docker 部署方案

---

## 架构概览

本平台由多个服务组成：

- `gateway`：统一外部 API 入口
- `user-service`：用户认证与权限管理
- `agent-service`：智能体业务逻辑
- `chat-service`：对话与会话管理
- `skill-service`：技能管理与执行
- `billing-service`：本地用量与计费统计
- `ai-engine`：AI 模型调用与 RAG 引擎
- `web-admin`：管理后台前端

此外还依赖：PostgreSQL、Redis、MongoDB、MinIO、NATS、Milvus 等组件。

---

## 快速启动

以下示例适用于已安装 Docker 和 Docker Compose 的环境：

```bash
cp .env.example .env
# 编辑 .env，填写数据库、Redis、模型 Key 等配置

docker compose up -d

docker compose logs -f
```

默认管理后台访问：

```text
http://localhost:20009
```

默认管理员账号：
- 用户名：`superadmin`
- 密码：`Admin@2026`

---

## 部署方式

推荐根据实际环境选择部署方案：

| 部署方案 | 适用场景 | 说明 |
|---|---|---|
| Docker Compose | 快速上线 | 推荐首选，适合支持容器的服务器环境 |
| 原生 Linux 安装 | 传统服务器 / 不使用容器 | 适合标准 Linux 系统部署，支持手动管理依赖服务 |
| 宝塔面板原生部署 | 宝塔面板服务器 | 面向宝塔用户，使用宝塔进程、网站、环境管理 |

详细部署方案请参考：
- `docs/宝塔面板-Docker-部署方案.md`
- `docs/宝塔面板原生部署指南.md`

---

## 环境依赖

### 必备组件

- Go 1.20+
- Python 3.12+
- Node.js 18+
- PostgreSQL 14+
- Redis 7+
- MongoDB 6+
- MinIO
- NATS
- Milvus

### 可选组件

- Nginx（前端静态站点托管）

---

## 环境变量配置

复制配置模板：

```bash
cp .env.example .env
```

在 `.env` 中至少修改以下关键项：

- `JWT_SECRET`：JWT 签名密钥，务必修改为安全字符串
- `DB_HOST` / `DB_USER` / `DB_PASSWORD` / `DB_NAME`
- `REDIS_ADDR`
- `MONGO_URI`
- `MINIO_ENDPOINT` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY`
- `NATS_URL`
- `MILVUS_HOST` / `MILVUS_PORT`
- `QWEN_API_KEY` / `DEEPSEEK_API_KEY`

> LLM API Key 也可以通过平台管理后台配置，无需写入 `.env`。

---

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Gateway | 20001 | API 网关 |
| User Service | 20002 | 用户与认证 |
| Agent Service | 20003 | 智能体管理 |
| Chat Service | 20004 | 对话服务 |
| Skill Service | 20005 | 技能市场 |
| Billing Service | 20006 | 计费统计 |
| AI Engine HTTP | 20007 | AI 引擎 HTTP 接口 |
| AI Engine gRPC | 20008 | AI 引擎 gRPC 接口 |
| Web Admin | 20009 | 管理后台 |

## 一键发布脚本（scripts/）

适合**宝塔服务器日常迭代**——代码改动后无需手工编译、重启：

| 脚本 | 作用 |
|------|------|
| `scripts/deploy.sh` | 全栈一键发布（Go 后端 + ai-engine + 前端） |
| `scripts/release-fettle.sh` | Go 后端 6 服务编译与重启（`--only svc --restart`） |
| `scripts/release-ai-engine.sh` | ai-engine (uvicorn + celery worker) 启停 |
| `scripts/rollback.sh` | 按时间戳快照回滚 Go 二进制（`--latest` / `--to`） |
| `scripts/restart-standalone.sh` | 旧版 nohup 重启脚本（兼容保留） |

常用命令：

```bash
# 全栈发布
bash scripts/deploy.sh

# 仅修改某个 Go 服务时（其他服务不重启）
bash scripts/release-fettle.sh --only chat-service --restart

# 仅重启 ai-engine（uvicorn + celery）
bash scripts/release-ai-engine.sh --restart

# 构建失败/效果不对，回滚到上一个版本
bash scripts/rollback.sh --latest
```

详细参数见 [`scripts/README.md`](scripts/README.md)。底层走宝塔 Go 项目管理器 120s 守护自启，无需手工 `nohup`。

> 共享包 `backend/shared/config/config.go` 自动按可执行文件位置加载 `.env`，避免偷读主项目 fettle。

---

## 文档与示例

- `docs/宝塔面板-Docker-部署方案.md`：三套部署方案说明
- `scripts/install_minio_nats_milvus.sh`：MinIO / NATS / Milvus 原生安装示例脚本
- `.env.example`：环境变量模板

---

## 项目结构

```
/ backend/
  / agent-service/
  / ai-engine/
  / billing-service/
  / chat-service/
  / gateway/
  / skill-service/
  / user-service/
/ frontend/
  / web-admin/
  / web-platform/
  / web-website/
  / mini-app/
/ proto/
/ docs/
/ scripts/
```

---

## 贡献

欢迎提交 Issue 或 PR：

- 修复部署文档
- 补充环境变量说明
- 优化服务启动流程
- 增强 AI 引擎与插件支持

---

## 许可

Copyright © 2024-2026 李可乐. All rights reserved.

本项目采用 **GNU Affero General Public License v3.0 (AGPL-3.0)** 开源。

您可以自由使用、修改和分发本项目，但须遵守以下核心条件：

- **如果您修改了本软件并通过网络对外提供服务（如 SaaS），必须同样以 AGPL-3.0 公开您的修改后的源代码。**
- 本软件按"原样"提供，不作任何明示或暗示的担保。

详见 [LICENSE](./LICENSE) 文件。

### 商业授权

如果您需要在闭源商业产品中集成本项目，或需要企业级技术支持，请联系获取商业授权：

- 官网：https://fettle.cenkor.cn
- 邮箱：your-email@example.com

> 商业授权用户不受 AGPL-3.0 开源条款约束，可合法将本软件嵌入闭源产品或作为自有 SaaS 运营。