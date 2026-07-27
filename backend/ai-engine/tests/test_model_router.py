"""模型路由器的基本测试"""

from core.models.model_router import ModelRouter, TaskType


def test_model_router_initialization():
    """测试模型路由器初始化"""
    router = ModelRouter("config/models.yaml")
    assert len(router.models) > 0
    assert "qwen-plus" in router.models
    assert "deepseek-chat" in router.models


def test_select_model_by_id():
    """测试通过模型ID选择模型"""
    router = ModelRouter("config/models.yaml")
    model = router.select_model(model_id="qwen-plus")
    assert model.provider == "qwen"
    assert model.model_name == "qwen-plus"


def test_select_model_by_task():
    """测试按任务类型选择模型"""
    router = ModelRouter("config/models.yaml")
    model = router.select_model(task_type=TaskType.CODE_GENERATION)
    assert model.model_name == "deepseek-chat"


def test_select_model_by_tenant_plan():
    """测试按租户套餐选择模型"""
    router = ModelRouter("config/models.yaml")
    model = router.select_model(tenant_plan="free")
    assert model.model_name == "qwen-turbo"

    model = router.select_model(tenant_plan="enterprise")
    assert model.model_name == "qwen-max"


def test_embedding_model_selection():
    """测试嵌入模型选择"""
    router = ModelRouter("config/models.yaml")
    emb = router.get_embedding_model("text-embedding-v2")
    assert emb.provider == "qwen"
    assert emb.dimension == 1536


def test_list_models():
    """测试列出所有模型"""
    router = ModelRouter("config/models.yaml")
    models = router.list_models()
    assert len(models) >= 4

    providers = {m["provider"] for m in models}
    assert "qwen" in providers
    assert "deepseek" in providers
