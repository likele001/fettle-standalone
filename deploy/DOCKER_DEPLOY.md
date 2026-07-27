# Fettle Standalone - Docker 一键部署

## 环境要求

- Docker >= 24.0
- Docker Compose >= 2.20
- 宝塔面板（用于域名 SSL + 反向代理）

## 快速部署

```bash
# 进入部署目录
cd /www/wwwroot/fettle-standalone/deploy/docker

# 一键构建并启动（首次约 5-10 分钟）
docker compose up -d --build

# 查看所有容器状态
docker compose ps
```

## 服务列表

| 服务 | 端口 | 说明 |
|------|------|------|
| PostgreSQL | - | 数据库（内置，数据持久化到 volume） |
| Redis | - | 缓存（内置） |
| Gateway | 20001 | API 网关入口 |
| User Service | 20002 | 用户/认证服务 |
| Agent Service | 20003 | 智能体编排服务 |
| Chat Service | 20004 | 对话服务 |
| Skill Service | 20005 | 技能市场服务 |
| Billing Service | 20006 | 计费服务 |
| AI Engine | 20007 / 20008 | AI 引擎（HTTP + gRPC） |
| Web Admin | 20009 | 管理后台 |

## 宝塔反向代理配置

在宝塔面板添加两个站点，分别对应管理后台和 API：

### 站点一：管理后台

**域名：** 例如 `admin.fettle.cenkor.cn`

**反向代理配置（Nginx）：**

```nginx
# 宝塔站点设置 → 反向代理 → 添加
# 目标 URL：http://127.0.0.1:20009
# 发送域名：$host

# 或者在站点设置 → 配置文件添加以下内容：
location / {
    proxy_pass http://127.0.0.1:20009;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### 站点二：API 网关

**域名：** 例如 `api.fettle.cenkor.cn`（也可以和管理后台共用域名，加 /api/ 前缀）

**反向代理配置：**

```nginx
location / {
    proxy_pass http://127.0.0.1:20001;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

> **注意：** 在宝塔添加站点后，会自动申请 SSL 证书，无需手动配置 HTTPS。

### 完整 Nginx 配置示例（单域名方案）

如果想用同一个域名（例如 `fettle.cenkor.cn`），配置如下：

```nginx
server {
    listen 80;
    listen 443 ssl;
    server_name fettle.cenkor.cn;

    ssl_certificate /www/server/panel/vhost/cert/fettle.cenkor.cn/fullchain.pem;
    ssl_certificate_key /www/server/panel/vhost/cert/fettle.cenkor.cn/privkey.pem;

    # 管理后台
    location / {
        proxy_pass http://127.0.0.1:20009;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # API 网关
    location /api/ {
        proxy_pass http://127.0.0.1:20001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## 常用管理命令

```bash
# 查看所有容器
docker compose ps

# 查看日志
docker compose logs -f          # 全部
docker compose logs -f gateway  # 单个服务

# 重启单个服务
docker compose restart user-service

# 重新构建并启动（代码更新后）
docker compose up -d --build <service>

# 停止所有服务
docker compose down
```

## 配置信息

- 数据库：`standalone / CHANGE_ME`
- Redis：无密码
- JWT 密钥：`CHANGE_ME_openssl_rand_hex_32`

如需修改，编辑 `deploy/docker/docker-compose.yml` 后重新构建对应服务即可。

## 架构

```
用户 → 宝塔 Nginx（SSL+域名）→ Docker 容器（20001/20009）
```
