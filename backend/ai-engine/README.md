# AI Engine 服务

AI 智能体引擎服务，提供模型调用、RAG 检索、工具执行等核心能力。

## 功能特性

- **多模型支持**：通义千问、DeepSeek、MiniMax、OpenAI
- **智能路由**：根据任务类型和租户套餐自动选择最优模型
- **RAG 管道**：文档加载、分割、向量化、检索
- **工具系统**：工具注册、执行、任务规划
- **gRPC 服务**：高性能 RPC 接口
- **异步任务**：Celery 任务队列支持

## 端口

- HTTP API: 9700
- gRPC: 9701

## 快速开始

### 1. 安装依赖

```bash
cd backend/ai-engine
pip install -r requirements.txt
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，设置 API 密钥等配置
```

### 3. 生成 gRPC 代码

```bash
python -m grpc_tools.protoc \
    -I./proto \
    --python_out=./proto \
    --grpc_python_out=./proto \
    ./proto/ai_engine.proto
```

### 4. 启动服务

```bash
python main.py
```

### 5. 启动 Celery Worker（可选）

```bash
celery -A tasks.celery_app worker --loglevel=info
```

## 目录结构

```
backend/ai-engine/
├── main.py                 # FastAPI 入口
├── config/                 # 配置
│   ├── settings.py         # 应用配置
│   └── models.yaml         # 模型配置
├── proto/                  # gRPC 定义
│   └── ai_engine.proto
├── grpc_server/            # gRPC 服务
│   └── server.py
├── api/                    # HTTP API
│   └── health.py
├── core/                   # 核心模块
│   ├── models/             # 模型路由
│   │   ├── model_router.py
│   │   └── providers/      # 模型提供商
│   ├── rag/                # RAG 管道
│   │   ├── document_loader.py
│   │   ├── text_splitter.py
│   │   └── retriever.py
│   ├── memory/             # 记忆管理
│   │   └── working_memory.py
│   └── tools/              # 工具系统
│       ├── tool_registry.py
│       └── tool_executor.py
├── tasks/                  # Celery 任务
│   ├── celery_app.py
│   └── async_tasks.py
└── deploy/
    └── Dockerfile
```

## API 接口

### HTTP API

- `GET /health` - 健康检查
- `GET /ready` - 就绪检查

### gRPC API

- `RecognizeIntent` - 意图识别
- `PlanTask` - 任务规划
- `GenerateReply` - 回复生成
- `StreamReply` - 流式回复
- `GenerateEmbedding` - 嵌入生成
- `RetrieveKnowledge` - 知识检索

## 配置说明

### 模型配置

编辑 `config/models.yaml` 配置可用模型和路由策略。

### 环境变量

参考 `.env.example` 设置环境变量。

关键配置：
- `QWEN_API_KEY` - 通义千问 API 密钥
- `DEEPSEEK_API_KEY` - DeepSeek API 密钥
- `MILVUS_HOST` - Milvus 地址
- `REDIS_HOST` - Redis 地址

## 部署

### Docker

```bash
docker build -t ai-engine -f deploy/Dockerfile .
docker run -p 9700:9700 -p 9701:9701 ai-engine
```

### Docker Compose

参考 `deploy/docker/docker-compose.yml`

## 开发计划

- [x] 基础框架搭建
- [x] 模型路由器
- [x] gRPC 服务定义
- [x] RAG 管道基础
- [x] 工具系统框架
- [ ] 完整 RAG 实现（Milvus 集成）
- [ ] 长期记忆实现
- [ ] 沙箱代码执行
- [ ] 更多工具集成
