"""RAG 文档加载器"""
import logging
from typing import List, Dict, Any
from pathlib import Path
from dataclasses import dataclass

logger = logging.getLogger(__name__)


@dataclass
class Document:
    """文档对象"""
    content: str
    metadata: Dict[str, Any]
    doc_id: str = ""


class DocumentLoader:
    """文档加载器"""
    
    @staticmethod
    def load_pdf(file_path: str) -> List[Document]:
        """加载 PDF 文件"""
        try:
            from pypdf import PdfReader
            
            reader = PdfReader(file_path)
            documents = []
            
            for i, page in enumerate(reader.pages):
                text = page.extract_text()
                if text.strip():
                    documents.append(Document(
                        content=text,
                        metadata={
                            "source": file_path,
                            "page": i + 1,
                            "type": "pdf"
                        }
                    ))
            
            logger.info(f"Loaded {len(documents)} pages from {file_path}")
            return documents
        
        except Exception as e:
            logger.error(f"Failed to load PDF {file_path}: {e}")
            return []
    
    @staticmethod
    def load_docx(file_path: str) -> List[Document]:
        """加载 Word 文档"""
        try:
            from docx import Document as DocxDocument
            
            doc = DocxDocument(file_path)
            paragraphs = []
            
            for para in doc.paragraphs:
                if para.text.strip():
                    paragraphs.append(para.text)
            
            if paragraphs:
                return [Document(
                    content="\n\n".join(paragraphs),
                    metadata={
                        "source": file_path,
                        "type": "docx"
                    }
                )]
            
            return []
        
        except Exception as e:
            logger.error(f"Failed to load DOCX {file_path}: {e}")
            return []
    
    @staticmethod
    def load_txt(file_path: str) -> List[Document]:
        """加载文本文件"""
        try:
            with open(file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            
            if content.strip():
                return [Document(
                    content=content,
                    metadata={
                        "source": file_path,
                        "type": "txt"
                    }
                )]
            
            return []
        
        except Exception as e:
            logger.error(f"Failed to load TXT {file_path}: {e}")
            return []
    
    @classmethod
    def load(cls, file_path: str) -> List[Document]:
        """根据文件类型加载"""
        path = Path(file_path)
        suffix = path.suffix.lower()
        
        if suffix == '.pdf':
            return cls.load_pdf(file_path)
        elif suffix == '.docx':
            return cls.load_docx(file_path)
        elif suffix == '.txt':
            return cls.load_txt(file_path)
        else:
            logger.warning(f"Unsupported file type: {suffix}")
            return []
