"""核心模块"""

model_router = None
ModelRouter = None
TaskType = None
working_memory_manager = None
WorkingMemoryManager = None
DocumentLoader = None
TextSplitter = None
MilvusRetriever = None


def _lazy_import():
    global model_router, ModelRouter, TaskType
    global working_memory_manager, WorkingMemoryManager
    global DocumentLoader, TextSplitter, MilvusRetriever
    
    from .models.model_router import model_router as mr, ModelRouter, TaskType
    model_router, ModelRouter, TaskType = mr, ModelRouter, TaskType
    
    try:
        from .memory.working_memory import working_memory_manager as wmm, WorkingMemoryManager
        working_memory_manager, WorkingMemoryManager = wmm, WorkingMemoryManager
    except ImportError:
        pass
    
    try:
        from .rag import DocumentLoader, TextSplitter, MilvusRetriever
    except ImportError:
        pass


def get_model_router():
    if model_router is None:
        _lazy_import()
    return model_router


def get_working_memory_manager():
    if working_memory_manager is None:
        _lazy_import()
    return working_memory_manager


__all__ = [
    'model_router',
    'ModelRouter',
    'TaskType',
    'working_memory_manager',
    'WorkingMemoryManager',
    'DocumentLoader',
    'TextSplitter',
    'MilvusRetriever',
    'get_model_router',
    'get_working_memory_manager',
]