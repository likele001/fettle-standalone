"""AI Engine 配置管理"""
from typing import Optional, Dict, List
from pydantic_settings import BaseSettings
from pydantic import Field


class Settings(BaseSettings):
    """应用配置"""
    
    model_config = {"protected_namespaces": (), "env_file": ".env", "case_sensitive": False, "extra": "ignore"}
    app_name: str = "ai-engine"
    app_env: str = "development"
    app_http_port: int = 9700
    app_grpc_port: int = 9701
    debug: bool = False
    
    # 数据库配置
    db_host: str = "localhost"
    db_port: int = 5432
    db_user: str = "ai_platform"
    # 密码必须从环境变量注入，禁止默认值/明文
    db_password: str = Field(..., min_length=1, description="数据库密码，必须从 DB_PASSWORD 环境变量注入")
    db_name: str = "ai_platform"
    
    @property
    def database_url(self) -> str:
        return f"postgresql://{self.db_user}:{self.db_password}@{self.db_host}:{self.db_port}/{self.db_name}"
    
    @property
    def database_url_async(self) -> str:
        return f"postgresql+asyncpg://{self.db_user}:{self.db_password}@{self.db_host}:{self.db_port}/{self.db_name}"
    
    # Redis 配置
    redis_host: str = "localhost"
    redis_port: int = 6379
    redis_db: int = 0
    redis_password: Optional[str] = None
    
    @property
    def redis_url(self) -> str:
        if self.redis_password:
            return f"redis://:{self.redis_password}@{self.redis_host}:{self.redis_port}/{self.redis_db}"
        return f"redis://{self.redis_host}:{self.redis_port}/{self.redis_db}"
    
    # Milvus 配置
    milvus_host: str = "localhost"
    milvus_port: int = 19530
    
    # NATS 配置
    nats_url: str = "nats://localhost:4222"
    
    # Celery 配置
    celery_broker_url: str = "redis://localhost:6379/1"
    celery_result_backend: str = "redis://localhost:6379/2"

    # billing-service 内部接口（平台托管计费统一入口）
    billing_service_url: str = "http://localhost:9600"
    billing_internal_token: str = ""
    skill_service_url: str = "http://localhost:9500"
    agent_service_url: str = "http://localhost:9300"
    
    # 模型配置
    default_model: str = "qwen-plus"
    model_timeout: int = 30
    model_max_retries: int = 3
    
    # API 密钥（从环境变量读取）
    openai_api_key: Optional[str] = None
    openai_base_url: Optional[str] = None
    qwen_api_key: Optional[str] = None
    deepseek_api_key: Optional[str] = None
    minimax_api_key: Optional[str] = None
    
    # AI 引擎内部鉴权令牌（必填，从环境/.env 注入；security.py 通过 os.getenv 读取）
    ai_engine_internal_token: Optional[str] = None

    # 嵌入模型配置
    embedding_model: str = "text-embedding-v2"
    embedding_dimension: int = 1536
    
    # RAG 配置
    chunk_size: int = 500
    chunk_overlap: int = 50
    top_k: int = 5
    # RAG 重排 (Re-Rank) 配置
    rerank_enabled: bool = True
    rerank_mode: str = "lexical"      # lexical=无依赖 | neural=CrossEncoder(需装 sentence-transformers)
    rerank_model: str = "BAAI/bge-reranker-v2-m3"
    rerank_candidate_k: int = 20       # 重排前拉取的候选数
    rerank_min_score: float = 0.2     # 候选最低向量分阈值



settings = Settings()
