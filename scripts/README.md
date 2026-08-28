# scripts/ — 一键发布脚本（standalone）

## 入口

| 脚本 | 用途 |
|------|------|
| `deploy.sh` | 全栈一键发布（Go 后端 + ai-engine + 前端） |
| `release-fettle.sh` | Go 后端 6 服务编译与重启 |
| `release-ai-engine.sh` | ai-engine (uvicorn + celery worker) 启停 |
| `rollback.sh` | 按时间戳快照回滚 Go 二进制 |
| `restart-standalone.sh` | **旧版** nohup 重启脚本（保留兼容，可逐步迁移） |

## 端口/路径差异（vs fettle 主项目）

| 项 | fettle | fettle-standalone |
|---|---|---|
| Go 端口 | 9100-9600 | 20001-20006 |
| ai-engine HTTP/gRPC | 9700/9701 | 20007/20008 |
| 日志 | `/www/wwwlogs/go` + `/www/wwwlogs/python/ai-engine` | `./logs/` |
| ai-engine venv | `/www/server/pyporject_evn/fettle` | `/www/server/pyporject_evn/fettlept` |
| Redis DB | 51 + 1/2（celery） | 52 + 1/2（celery） |

## 快速用法

```bash
# 全栈：构建 + 重启 + 前端
bash scripts/deploy.sh

# 只重启 ai-engine（已部署，未改代码）
bash scripts/release-ai-engine.sh --restart

# 只重启指定 Go 服务（如 chat-service）
bash scripts/release-fettle.sh --only chat-service --restart

# 仅构建，不重启（生成新二进制，由宝塔 120s 守护自启）
bash scripts/release-fettle.sh --snapshot --only chat-service

# 回滚到上一个版本
bash scripts/rollback.sh --latest
```

## deploy.sh 选项

```
--no-fe            跳过前端构建
--no-snapshot      不归档 build/ 旧二进制
--no-restart       仅构建，不重启任何服务
--keep <N>         快照保留 N 份（默认 3）
--only <target>    只操作某个目标：go|ai|fe
```

## release-fettle.sh 选项

```
--no-build         跳过构建
--restart          构建后批量重启
--baota|--systemctl|--nohup   重启策略（默认 baota）
--only <svc>       只操作指定服务（chat-service 等）
--log-dir <path>   日志目录（默认 ./logs）
--pid-dir <path>   宝塔 Go 守护 PID 目录（默认 /var/tmp/gopids）
--fe               构建完执行 npm run build + nginx -s reload
--snapshot         构建前归档 build/<svc> 到 build/_rollback/<TS>/
--keep <N>         配合 --snapshot，仅保留最近 N 份（默认 3）
```

## release-ai-engine.sh 选项

```
--status           查看 ai-engine + celery 状态
--start|--stop|--restart
--only <comp>      只操作 uvicorn 或 celery
--cwd <path>       ai-engine 目录（默认 backend/ai-engine）
--venv <path>      Python venv（默认 /www/server/pyporject_evn/fettlept）
--log <path>       uvicorn 日志（默认 ./logs/ai-engine.log）
--celery-log <p>   celery 日志（默认 ./logs/ai-worker.log）
--workers <N>      celery worker 数（默认 4）
--http-port <N>    uvicorn 端口（默认 20007）
--bind <ip:port>   uvicorn 绑定地址（默认 0.0.0.0:20007）
```

## rollback.sh 选项

```
--list               列出所有可用历史版本
--to <timestamp>     回滚到指定时间戳版本（如 20260828_231821）
--latest             回滚到最近一次保留的备份
--keep <N>           保留最近 N 个备份（默认 3），更老的删除
--build-dir <path>   build 目录（默认项目根/build）
--src <dir>          备份目录（默认 build/_rollback）
```

## 部署模型（宝塔）

- **进程管理**：宝塔 Go 项目管理器（120s 守护自启）+ 宝塔 Python 项目管理器
- **重启策略**：脚本发 `pkill`，由宝塔守护自动拉起（120s 内）
- **手动面板操作**：仅在新建服务、修改 env 时需在面板操作
- **环境变量**：面板 env_list `env_file=/www/wwwroot/fettle-standalone/.env` 已配，独立加载

## 注意事项

- 回滚仅作用于 `build/` 二进制；源码错误请用 `git revert`。
- 共享包 `backend/shared/config/config.go` 已解耦硬编码，按可执行文件/cwd 自动加载 `.env`，避免偷读主项目 fettle。
- standalone 与主项目 fettle 端口段、Redis DB、venv 完全隔离。
- 旧的 `restart-standalone.sh` 与新 `deploy.sh`/`release-*.sh` 功能有重叠，建议新流程统一走 `deploy.sh`。