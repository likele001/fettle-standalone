# fettle-standalone 部署方案

## 1. 方案总览

本文档包含三套完整部署方案：
- 宝塔面板单独部署方案
- Docker 单独部署方案
- 原生 Linux 系统命令部署方案

其中每一套方案均包含：
- 环境准备
- 依赖服务安装
- 项目部署步骤
- 启动与验证

---

## 2. 统一准备

### 2.1 系统要求

- 操作系统：Ubuntu 22.04 / Debian 12 / CentOS 7+ / 其他主流 Linux
- 内存：至少 4GB
- 磁盘：至少 20GB 可用
- 端口：20001-20009、9801 等可用

### 2.2 项目目录结构

请确认项目目录包含：
- `docker-compose.yml`
- `.env.example`
- `backend/`
- `frontend/`
- `proto/`
- `deploy/`

建议项目路径：
- `/www/wwwroot/fettle-standalone`

### 2.3 关键依赖服务

项目依赖：
- PostgreSQL
- Redis
- MongoDB
- MinIO
- Milvus
- NATS

这些依赖可以通过宝塔面板安装、Docker 容器启动，或系统原生命令安装。

---

## 3. 宝塔面板单独部署方案

### 3.1 适用场景

适用于希望使用宝塔面板进行项目部署、依赖服务安装与进程管理，但不使用 Docker 容器的环境。

### 3.2 前提条件

- 已安装宝塔面板
- 宝塔面板支持文件管理、终端、计划任务、软件商店
- 系统已安装 Go、Python、Node.js、Nginx（建议通过宝塔软件管理安装）

### 3.3 宝塔面板步骤

#### 3.3.1 安装宝塔面板

1. 在服务器上执行宝塔安装脚本（参考宝塔官网）
2. 登录宝塔管理后台
3. 进入“软件商店”安装必需组件：
   - PostgreSQL
   - Redis
   - MongoDB
   - Nginx
   - Node.js
   - Python
   - Go

#### 3.3.2 上传项目代码

1. 登录宝塔“文件”模块
2. 将 `fettle-standalone` 整体代码上传至服务器，例如：
   - `/www/wwwroot/fettle-standalone`
3. 确认目录下存在 `.env.example`、`backend/`、`frontend/`、`proto/` 等文件

#### 3.3.3 安装依赖服务

使用宝塔面板原生安装以下服务：

- PostgreSQL
- Redis
- MongoDB

如果宝塔软件商店没有直接提供 MinIO、Milvus、NATS，可以通过宝塔“终端”执行原生安装命令：

##### MinIO 原生安装

```bash
# 进入安装目录
cd /usr/local/bin
sudo curl -L https://dl.min.io/server/minio/release/linux-amd64/minio -o minio
sudo chmod +x minio

# 创建数据目录
sudo mkdir -p /var/minio/data
sudo chown -R www:www /var/minio/data

# 创建 systemd 服务
cat <<'EOF' | sudo tee /etc/systemd/system/minio.service
[Unit]
Description=MinIO
After=network.target

[Service]
User=www
Group=www
ExecStart=/usr/local/bin/minio server /var/minio/data --address :9000
Restart=always
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now minio
```

默认地址：`http://127.0.0.1:9000`

##### NATS 原生安装

```bash
sudo curl -L -f https://github.com/nats-io/nats-server/releases/download/v2.10.9/nats-server-v2.10.9-linux-amd64.tar.gz -o /tmp/nats-server.tar.gz
sudo tar xzf /tmp/nats-server.tar.gz -C /tmp
sudo mv /tmp/nats-server-v2.10.9-linux-amd64/nats-server /usr/local/bin/
sudo chmod +x /usr/local/bin/nats-server

cat <<'EOF' | sudo tee /etc/systemd/system/nats.service
[Unit]
Description=NATS Server
After=network.target

[Service]
ExecStart=/usr/local/bin/nats-server -a 0.0.0.0 -p 4222
Restart=always
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now nats
```

默认地址：`nats://127.0.0.1:4222`

##### Milvus 原生安装

Milvus 官方推荐使用 Docker/Compose，但在宝塔面板中如果只允许原生方式，可采用 Milvus 二进制或离线包。以下示例为二进制方式：

```bash
# 下载并解压 Milvus 二进制
sudo curl -L -f https://github.com/milvus-io/milvus/releases/download/v2.3.0/milvus-standalone-linux-amd64.tar.gz -o /tmp/milvus.tar.gz
sudo tar xzf /tmp/milvus.tar.gz -C /tmp
sudo mkdir -p /usr/local/milvus
sudo mv /tmp/milvus/* /usr/local/milvus/

# 准备数据目录
sudo mkdir -p /var/lib/milvus
sudo chown -R www:www /var/lib/milvus

# 启动 Milvus
sudo -u www /usr/local/milvus/bin/milvus run standalone &
```

默认地址：`127.0.0.1:19530`

记录账号密码并写入 `.env`。

建议参考源码脚本：`scripts/install_minio_nats_milvus.sh`，该脚本包含 MinIO、NATS、Milvus 的原生安装命令示例。

#### 3.3.4 编辑环境变量

1. 进入宝塔“文件”模块
2. 复制 `.env.example` 为 `.env`
3. 编辑 `.env`，至少设置：
   - `JWT_SECRET`
   - `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
   - `REDIS_ADDR`
   - `MONGO_URI`
   - `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`
   - `NATS_URL`
   - `MILVUS_HOST`, `MILVUS_PORT`
   - `QWEN_API_KEY` / `DEEPSEEK_API_KEY`

#### 3.3.5 构建与启动后端服务

在宝塔“终端”进入项目目录：

```bash
cd /www/wwwroot/fettle-standalone
```

依次编译 Go 服务：

```bash
cd backend/gateway
go build -o gateway ./main.go
cd ../user-service
go build -o user-service ./main.go
cd ../agent-service
go build -o agent-service ./main.go
cd ../chat-service
go build -o chat-service ./main.go
cd ../skill-service
go build -o skill-service ./main.go
cd ../billing-service
go build -o billing-service ./main.go
```

启动服务：

```bash
cd /www/wwwroot/fettle-standalone/backend/user-service
./user-service &
cd ../agent-service
./agent-service &
cd ../chat-service
./chat-service &
cd ../skill-service
./skill-service &
cd ../billing-service
./billing-service &
cd ../gateway
./gateway &
```

启动 AI 引擎：

```bash
cd /www/wwwroot/fettle-standalone/backend/ai-engine
python main.py &
```

将前端构建为静态文件并配置 Nginx：

```bash
cd /www/wwwroot/fettle-standalone/frontend/web-admin
npm install
npm run build
```

使用宝塔“网站”模块创建一个站点，指向 `frontend/web-admin/dist` 目录。

#### 3.3.6 验证服务

- Web 管理后台：`http://服务器IP:20009`（或宝塔网站绑定域名）
- Gateway：`http://服务器IP:20001`

查看运行日志：

```bash
ps -ef | grep gateway
```

如果使用宝塔“计划任务”，可为每个后端服务创建启动脚本并设置开机自启。

#### 3.3.7 宝塔运维管理

- 使用宝塔“文件”管理修改代码和配置
- 使用宝塔“终端”执行编译与启动命令
- 使用宝塔“计划任务”创建服务启动脚本
- 使用宝塔“网站”模块托管前端静态站点

---

## 4. Docker 单独部署方案

### 4.1 适用场景

适用于已具备 Docker/Docker Compose 环境，希望直接通过 Docker 进行项目部署的用户。

### 4.2 先决条件

- 已安装 Docker
- 已安装 Docker Compose v2 或 docker-compose
- 可访问系统终端

### 4.3 安装 Docker / Docker Compose

#### Ubuntu / Debian

```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg lsb-release
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl enable --now docker
```

#### CentOS

```bash
sudo yum install -y yum-utils
git curl
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl enable --now docker
```

#### 验证

```bash
docker version
docker compose version
```

### 4.4 依赖服务部署（Docker 方式）

如果依赖服务还未准备好，可以直接使用 Docker 启动：

```bash
docker run -d --name fettle-postgres \
  -e POSTGRES_USER=ai_platform \
  -e POSTGRES_PASSWORD=ai_platform_secret \
  -e POSTGRES_DB=ai_platform \
  -p 5432:5432 postgres:15

docker run -d --name fettle-redis \
  -p 6379:6379 redis:7

docker run -d --name fettle-mongodb \
  -p 27017:27017 \
  mongo:6.0

docker run -d --name fettle-minio \
  -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=ai_platform \
  -e MINIO_ROOT_PASSWORD=ai_platform_secret \
  minio/minio server /data

docker run -d --name fettle-nats \
  -p 4222:4222 nats:latest

docker run -d --name fettle-milvus \
  -p 19530:19530 \
  -e "TZ=UTC" \
  milvusdb/milvus:v2.3.0
```

> 注：Milvus 生产建议使用官方推荐部署方式，可根据需要扩展为 `docker-compose` 或 Kubernetes。

### 4.5 项目部署步骤

#### 4.5.1 上传或克隆代码

```bash
cd /www/wwwroot
# 若已在服务器上，则使用 git clone 或 scp 上传
# git clone <your-repo-url> fettle-standalone
cd fettle-standalone
```

#### 4.5.2 复制并编辑环境变量

```bash
cp .env.example .env
```

使用编辑器修改 `.env`：

```bash
nano .env
```

配置关键项：
- `JWT_SECRET`
- `DB_HOST=127.0.0.1`
- `DB_PASSWORD=ai_platform_secret`
- `REDIS_ADDR=127.0.0.1:6379`
- `MONGO_URI=mongodb://127.0.0.1:27017`
- `MINIO_ENDPOINT=127.0.0.1:9000`
- `MINIO_ACCESS_KEY=ai_platform`
- `MINIO_SECRET_KEY=ai_platform_secret`
- `NATS_URL=nats://127.0.0.1:4222`
- `MILVUS_HOST=127.0.0.1`
- `MILVUS_PORT=19530`

#### 4.5.3 启动服务

```bash
docker compose up -d
```

如果 `docker compose` 不可用：

```bash
docker-compose up -d
```

### 4.6 验证服务

```bash
docker compose ps
docker compose logs -f gateway
```

访问：
- `http://服务器IP:20001`
- `http://服务器IP:20009`

---

## 5. 原生 Linux 系统命令部署方案

### 5.1 适用场景

适用于不依赖宝塔面板，只使用系统原生命令完成环境搭建与部署的用户。

### 5.2 系统原生安装 Docker

#### Ubuntu / Debian

```bash
sudo apt update
sudo apt install -y ca-certificates curl gnupg lsb-release
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl enable --now docker
```

#### CentOS

```bash
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl enable --now docker
```

#### 验证安装

```bash
sudo docker version
sudo docker compose version
```

### 5.3 安装依赖服务（系统命令 + Docker）

#### PostgreSQL

```bash
sudo docker run -d --name fettle-postgres \
  -e POSTGRES_USER=ai_platform \
  -e POSTGRES_PASSWORD=ai_platform_secret \
  -e POSTGRES_DB=ai_platform \
  -p 5432:5432 postgres:15
```

#### Redis

```bash
sudo docker run -d --name fettle-redis \
  -p 6379:6379 redis:7
```

#### MongoDB

```bash
sudo docker run -d --name fettle-mongodb \
  -p 27017:27017 mongo:6.0
```

#### MinIO

```bash
sudo docker run -d --name fettle-minio \
  -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=ai_platform \
  -e MINIO_ROOT_PASSWORD=ai_platform_secret \
  minio/minio server /data
```

#### NATS

```bash
sudo docker run -d --name fettle-nats \
  -p 4222:4222 nats:latest
```

#### Milvus

```bash
sudo docker run -d --name fettle-milvus \
  -p 19530:19530 \
  milvusdb/milvus:v2.3.0
```

### 5.4 部署项目

#### 5.4.1 获取项目

```bash
cd /www/wwwroot
# 若已上传代码，则直接进入目录
# git clone <仓库地址> fettle-standalone
cd fettle-standalone
```

#### 5.4.2 复制与编辑环境变量

```bash
cp .env.example .env
nano .env
```

配置关键变量：
- `JWT_SECRET`
- `DB_HOST=127.0.0.1`
- `DB_PORT=5432`
- `DB_USER=ai_platform`
- `DB_PASSWORD=ai_platform_secret`
- `DB_NAME=ai_platform`
- `REDIS_ADDR=127.0.0.1:6379`
- `MONGO_URI=mongodb://127.0.0.1:27017`
- `MINIO_ENDPOINT=127.0.0.1:9000`
- `MINIO_ACCESS_KEY=ai_platform`
- `MINIO_SECRET_KEY=ai_platform_secret`
- `NATS_URL=nats://127.0.0.1:4222`
- `MILVUS_HOST=127.0.0.1`
- `MILVUS_PORT=19530`

#### 5.4.3 启动服务

```bash
sudo docker compose up -d
```

如果未提供 `docker compose`：

```bash
sudo docker-compose up -d
```

### 5.5 验证运行

```bash
sudo docker compose ps
sudo docker compose logs -f gateway
```

访问：
- `http://localhost:20001`
- `http://localhost:20009`

### 5.6 系统原生日志与管理命令

```bash
sudo docker logs -f fettle-postgres
sudo docker logs -f fettle-redis
sudo docker logs -f fettle-mongodb
sudo docker logs -f fettle-minio
sudo docker logs -f fettle-nats
sudo docker logs -f fettle-milvus
```

若服务依赖采用系统原生安装，可使用对应服务管理命令：
- `sudo systemctl status postgresql`
- `sudo systemctl status redis`
- `sudo systemctl status mongod`

---

## 6. 验证与故障排查

### 6.1 访问验证

- Gateway：`http://服务器IP:20001`
- Web 管理后台：`http://服务器IP:20009`

### 6.2 常见问题

- `JWT_SECRET` 未设置：服务启动失败
- 数据库连接失败：检查 `.env` 和依赖服务地址
- 端口冲突：检查是否有重复占用端口
- AI 引擎异常：确认 `ai-engine` 服务是否已启动

### 6.3 日志检查

```bash
docker compose logs -f gateway

docker compose logs -f ai-engine
```

---

## 7. 方案总结

本文件提供三套部署方案：
- `宝塔面板单独部署方案`：适合使用宝塔管理项目与容器的环境
- `Docker 单独部署方案`：适合已安装 Docker/Docker Compose 的直接部署场景
- `原生 Linux 系统命令部署方案`：适合不依赖宝塔，仅使用系统命令完成部署的场景

请根据实际服务器环境选择对应方案，并按照步骤执行。


- 宝塔面板可直接安装
- 创建数据库 `ai_platform`
- 创建用户 `ai_platform`
- 设置密码与 `.env` 对应

### 9.2 Redis

- 使用宝塔安装或 Docker 部署
- 默认地址 `127.0.0.1:6379`

### 9.3 MongoDB

- 使用宝塔安装或 Docker 部署
- 默认 URI `mongodb://localhost:27017`

### 9.4 MinIO

- 建议使用 Docker 容器部署
- 默认地址 `localhost:9000`
- 设置 `MINIO_ACCESS_KEY`、`MINIO_SECRET_KEY`

### 9.5 NATS

- 使用 Docker 容器部署
- 默认地址 `nats://localhost:4222`

### 9.6 Milvus

- 使用 Docker 容器部署
- 默认地址 `localhost:19530`
- 若不使用 Milvus，可先留空并在后续配置中调整

---

## 10 启动顺序与故障排查

### 10.1 启动顺序

1. Redis、PostgreSQL、MongoDB、MinIO、NATS、Milvus
2. `docker compose up -d` 启动 `fettle-standalone`
3. 访问 `web-admin`

### 10.2 常见问题

- `JWT_SECRET` 未设置：服务启动失败
- 数据库连接失败：检查 `.env` 中数据库地址和密码
- 端口冲突：检查端口是否被其他服务占用
- AI 引擎不可用：确认 `ai-engine` 容器正常运行

### 10.3 日志查看命令

```bash
docker compose logs -f gateway
docker compose logs -f user-service
docker compose logs -f agent-service
docker compose logs -f chat-service
docker compose logs -f billing-service
docker compose logs -f ai-engine
```

---

## 11 方案总结

该方案通过宝塔面板辅助管理并使用 Docker Compose 统一部署 `fettle-standalone`，兼顾系统原生服务与容器管理。

步骤简要：
1. 安装 Docker、Docker Compose、宝塔面板和 Docker 插件
2. 准备依赖服务(PostgreSQL/Redis/MongoDB/MinIO/NATS/Milvus)
3. 上传项目并配置 `.env`
4. 使用 `docker compose up -d` 启动服务
5. 访问 `http://服务器IP:20009` 验证

---

## 12 附录：`docker-compose.yml` 关键说明

项目根目录已存在 `docker-compose.yml`，它定义了所有核心服务容器启动配置。请确保该文件与本方案中的端口与环境变量一致。

若需要进一步定制容器持久化数据目录或端口映射，可修改 `docker-compose.yml` 中相应服务配置。
