"""LangChain 封装的检索器"""
import logging
from typing import List, Dict, Any, Optional
from langchain_core.retrievers import BaseRetriever
from langchain_core.callbacks import CallbackManagerForRetrieverRun
from langchain_core.documents import Document as LangChainDocument

from core.rag.knowledge_retriever import KnowledgeRetriever
from config.settings import settings

logger = logging.getLogger(__name__)


class MilvusLangChainRetriever(BaseRetriever):
    """Milvus LangChain 检索器"""

    knowledge_base_id: str
    tenant_id: str
    top_k: int = 3
    score_threshold: float = 0.5

    def __init__(
        self,
        knowledge_base_id: str,
        tenant_id: str,
        top_k: int = 3,
        score_threshold: float = 0.5,
        **kwargs
    ):
        super().__init__(**kwargs)
        self.knowledge_base_id = knowledge_base_id
        self.tenant_id = tenant_id
        self.top_k = top_k
        self.score_threshold = score_threshold
        self._retriever = KnowledgeRetriever()

    async def _agetrieve(
        self,
        query: str,
        run_manager: Optional[CallbackManagerForRetrieverRun] = None
    ) -> List[LangChainDocument]:
        """异步检索"""
        results = await self._retriever.retrieve(
            knowledge_base_id=self.knowledge_base_id,
            query=query,
            tenant_id=self.tenant_id,
            top_k=self.top_k,
            score_threshold=self.score_threshold
        )

        documents = []
        for result in results:
            doc = LangChainDocument(
                page_content=result.get("content", ""),
                metadata={
                    "score": result.get("score", 0),
                    "document_id": result.get("document_id", ""),
                    "knowledge_base_id": self.knowledge_base_id,
                    **result.get("metadata", {})
                }
            )
            documents.append(doc)

        return documents

    def _retrieve(
        self,
        query: str,
        run_manager: Optional[CallbackManagerForRetrieverRun] = None
    ) -> List[LangChainDocument]:
        """同步检索"""
        import asyncio
        return asyncio.run(self._agetrieve(query, run_manager))


class RAGChain:
    """RAG 链"""

    def __init__(self, llm, retriever: BaseRetriever):
        self.llm = llm
        self.retriever = retriever

    async def run(self, query: str) -> str:
        """执行 RAG 查询"""
        from langchain_core.prompts import ChatPromptTemplate
        from langchain_core.output_parsers import StrOutputParser
        from langchain_core.runnables import RunnablePassthrough

        template = """
        你是一个专业的助手，根据提供的上下文信息回答用户问题。
        
        上下文：
        {context}
        
        用户问题：
        {question}
        
        请根据上下文信息回答问题，如果上下文没有相关信息，请明确说明。
        """

        prompt = ChatPromptTemplate.from_template(template)

        rag_chain = (
            {"context": self.retriever | self._format_docs, "question": RunnablePassthrough()}
            | prompt
            | self.llm
            | StrOutputParser()
        )

        return await rag_chain.ainvoke(query)

    def _format_docs(self, docs: List[LangChainDocument]) -> str:
        """格式化文档"""
        formatted = []
        for i, doc in enumerate(docs, 1):
            score = doc.metadata.get("score", 0)
            formatted.append(f"[文档{i}] (相关度: {score:.2f})\n{doc.page_content}")
        return "\n\n---\n\n".join(formatted)