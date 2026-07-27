"""高级检索器"""
import logging
from typing import List, Dict, Any, Optional
from langchain_core.retrievers import BaseRetriever
from langchain_core.callbacks import CallbackManagerForRetrieverRun
from langchain_core.documents import Document

from .langchain_retriever import MilvusLangChainRetriever

logger = logging.getLogger(__name__)


class MultiQueryRetriever(BaseRetriever):
    """多查询检索器"""

    retriever: BaseRetriever
    llm: Any
    num_queries: int = 3

    def __init__(self, retriever: BaseRetriever, llm, num_queries: int = 3, **kwargs):
        super().__init__(**kwargs)
        self.retriever = retriever
        self.llm = llm
        self.num_queries = num_queries

    async def _agetrieve(
        self,
        query: str,
        run_manager: Optional[CallbackManagerForRetrieverRun] = None
    ) -> List[Document]:
        """异步检索"""
        queries = await self._generate_queries(query)

        all_docs = []
        seen_content = set()

        for q in queries:
            docs = await self.retriever.agetrieve(q)
            for doc in docs:
                content = doc.page_content
                if content not in seen_content:
                    seen_content.add(content)
                    all_docs.append(doc)

        all_docs.sort(key=lambda x: x.metadata.get("score", 0), reverse=True)
        return all_docs[:len(all_docs)]

    def _retrieve(
        self,
        query: str,
        run_manager: Optional[CallbackManagerForRetrieverRun] = None
    ) -> List[Document]:
        """同步检索"""
        import asyncio
        return asyncio.run(self._agetrieve(query, run_manager))

    async def _generate_queries(self, original_query: str) -> List[str]:
        """生成多个查询"""
        from langchain_core.prompts import ChatPromptTemplate
        from langchain_core.output_parsers import StrOutputParser

        template = """
        你是一个查询改写专家。请将用户的原始查询改写为 {num_queries} 个不同的查询，以便从知识库中检索到更多相关信息。
        
        原始查询：{query}
        
        请输出 {num_queries} 个不同的查询，每行一个。
        """

        prompt = ChatPromptTemplate.from_template(template)
        chain = prompt | self.llm | StrOutputParser()

        result = await chain.ainvoke({
            "query": original_query,
            "num_queries": self.num_queries
        })

        queries = [line.strip() for line in result.strip().split("\n") if line.strip()]
        return queries[:self.num_queries]


class ContextualCompressionRetriever(BaseRetriever):
    """上下文压缩检索器"""

    base_retriever: BaseRetriever
    llm: Any

    def __init__(self, base_retriever: BaseRetriever, llm, **kwargs):
        super().__init__(**kwargs)
        self.base_retriever = base_retriever
        self.llm = llm

    async def _agetrieve(
        self,
        query: str,
        run_manager: Optional[CallbackManagerForRetrieverRun] = None
    ) -> List[Document]:
        """异步检索"""
        docs = await self.base_retriever.agetrieve(query)
        return await self._compress_docs(query, docs)

    def _retrieve(
        self,
        query: str,
        run_manager: Optional[CallbackManagerForRetrieverRun] = None
    ) -> List[Document]:
        """同步检索"""
        import asyncio
        return asyncio.run(self._agetrieve(query, run_manager))

    async def _compress_docs(self, query: str, docs: List[Document]) -> List[Document]:
        """压缩文档，只保留与查询相关的内容"""
        from langchain_core.prompts import ChatPromptTemplate
        from langchain_core.output_parsers import StrOutputParser

        template = """
        你是一个信息压缩专家。请从以下文档中提取与用户查询最相关的内容。
        
        用户查询：{query}
        
        文档内容：
        {document}
        
        请只输出与查询相关的内容，如果没有相关内容，请输出空字符串。
        """

        prompt = ChatPromptTemplate.from_template(template)
        chain = prompt | self.llm | StrOutputParser()

        compressed_docs = []
        for doc in docs:
            compressed_content = await chain.ainvoke({
                "query": query,
                "document": doc.page_content
            })
            if compressed_content.strip():
                compressed_docs.append(Document(
                    page_content=compressed_content.strip(),
                    metadata=doc.metadata
                ))

        return compressed_docs


class RAGRetrieverFactory:
    """RAG 检索器工厂"""

    @classmethod
    def create_retriever(
        cls,
        knowledge_base_id: str,
        tenant_id: str,
        llm=None,
        retrieval_strategy: str = "simple",
        top_k: int = 3,
        **kwargs
    ) -> BaseRetriever:
        """
        创建检索器
        
        Args:
            knowledge_base_id: 知识库 ID
            tenant_id: 租户 ID
            llm: 语言模型
            retrieval_strategy: 检索策略 (simple, multi_query, compression, hybrid)
            top_k: 返回数量
        """
        base_retriever = MilvusLangChainRetriever(
            knowledge_base_id=knowledge_base_id,
            tenant_id=tenant_id,
            top_k=top_k,
            **kwargs
        )

        if retrieval_strategy == "multi_query" and llm:
            return MultiQueryRetriever(retriever=base_retriever, llm=llm, **kwargs)

        elif retrieval_strategy == "compression" and llm:
            return ContextualCompressionRetriever(base_retriever=base_retriever, llm=llm)

        elif retrieval_strategy == "hybrid" and llm:
            multi_query = MultiQueryRetriever(retriever=base_retriever, llm=llm, **kwargs)
            return ContextualCompressionRetriever(base_retriever=multi_query, llm=llm)

        return base_retriever