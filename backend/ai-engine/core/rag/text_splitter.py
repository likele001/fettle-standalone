"""文本分割器"""
import logging
from typing import List, Dict, Any
from dataclasses import dataclass

from .document_loader import Document

logger = logging.getLogger(__name__)


@dataclass
class TextChunk:
    """文本块"""
    content: str
    metadata: Dict[str, Any]
    chunk_id: str = ""


class TextSplitter:
    """文本分割器"""
    
    def __init__(
        self,
        chunk_size: int = 500,
        chunk_overlap: int = 50,
        separators: List[str] = None
    ):
        self.chunk_size = chunk_size
        self.chunk_overlap = chunk_overlap
        self.separators = separators or ["\n\n", "\n", "。", ".", "！", "!", "？", "?", "；", ";", " "]
    
    def split_documents(self, documents: List[Document]) -> List[TextChunk]:
        """分割文档列表"""
        chunks = []
        
        for doc in documents:
            doc_chunks = self.split_text(doc.content, doc.metadata)
            chunks.extend(doc_chunks)
        
        logger.info(f"Split {len(documents)} documents into {len(chunks)} chunks")
        return chunks
    
    def split_text(self, text: str, metadata: Dict[str, Any]) -> List[TextChunk]:
        """分割文本"""
        # 按分隔符分割
        splits = self._split_by_separators(text)
        
        # 合并成块
        chunks = []
        current_chunk = []
        current_length = 0
        
        for split in splits:
            split_length = len(split)
            
            if current_length + split_length > self.chunk_size and current_chunk:
                # 保存当前块
                chunk_content = " ".join(current_chunk)
                chunks.append(TextChunk(
                    content=chunk_content,
                    metadata=metadata.copy(),
                    chunk_id=f"chunk_{len(chunks)}"
                ))
                
                # 保留重叠部分
                overlap_length = 0
                overlap_chunks = []
                for item in reversed(current_chunk):
                    if overlap_length + len(item) <= self.chunk_overlap:
                        overlap_chunks.insert(0, item)
                        overlap_length += len(item)
                    else:
                        break
                
                current_chunk = overlap_chunks
                current_length = overlap_length
            
            current_chunk.append(split)
            current_length += split_length
        
        # 保存最后一块
        if current_chunk:
            chunk_content = " ".join(current_chunk)
            chunks.append(TextChunk(
                content=chunk_content,
                metadata=metadata.copy(),
                chunk_id=f"chunk_{len(chunks)}"
            ))
        
        return chunks
    
    def _split_by_separators(self, text: str) -> List[str]:
        """按分隔符分割文本"""
        if not text:
            return []
        
        # 找到最早出现的分隔符
        for sep in self.separators:
            if sep in text:
                parts = text.split(sep)
                result = []
                for part in parts:
                    if part.strip():
                        result.append(part.strip())
                return result
        
        # 如果没有分隔符，按长度分割
        return [text[i:i+self.chunk_size] for i in range(0, len(text), self.chunk_size)]
