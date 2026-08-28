"""RAG 检索结果重排 (Re-Rank)

设计目标：
- 零外部依赖即可工作（lexical 模式，中文 bigram + 英文词重叠）。
- 可选 neural 模式（sentence-transformers CrossEncoder），未安装或失败时自动回退 lexical。
- 与现有 KnowledgeRetriever 解耦：只接收候选列表 + query，返回重排后列表。
"""
import logging
import re
from typing import List, Dict, Any, Optional

logger = logging.getLogger(__name__)


def _tokenize(text: str) -> List[str]:
    """中英文分词，无外部依赖。

    - 英文/数字：按词
    - 中文：单字 + 相邻 bigram（提升短 query 召回鲁棒性）
    """
    if not text:
        return []
    text = text.lower()
    en = re.findall(r"[a-z0-9]+", text)
    zh = re.findall(r"[一-龥]", text)
    bigrams = ["".join(p) for p in zip(zh, zh[1:])] if len(zh) > 1 else []
    return en + zh + bigrams


class LexicalReranker:
    """基于词重叠的无依赖重排器。返回 0~1 的相关性分值。"""

    def score(self, query: str, doc: str) -> float:
        q_tokens = _tokenize(query)
        if not q_tokens:
            return 0.0
        d_set = set(_tokenize(doc))
        if not d_set:
            return 0.0
        hit = sum(1 for t in q_tokens if t in d_set)
        precision = hit / len(q_tokens)
        coverage = hit / len(set(q_tokens)) if q_tokens else 0.0
        return round(0.6 * precision + 0.4 * coverage, 4)


class NeuralReranker:
    """基于 CrossEncoder 的神经重排。需 sentence-transformers。"""

    def __init__(self, model_name: str = "BAAI/bge-reranker-v2-m3"):
        self.model_name = model_name
        self._model = None

    def _ensure(self):
        if self._model is not None:
            return
        from sentence_transformers import CrossEncoder
        self._model = CrossEncoder(self.model_name)
        logger.info(f"Neural reranker loaded: {self.model_name}")

    def score(self, query: str, doc: str) -> float:
        self._ensure()
        out = self._model.predict([(query, doc)])
        try:
            return float(out[0])
        except (TypeError, IndexError):
            return float(out)


class Reranker:
    """重排门面。根据 mode 选择后端，失败自动回退 lexical。"""

    def __init__(self, mode: str = "lexical",
                 neural_model: str = "BAAI/bge-reranker-v2-m3"):
        self.mode = (mode or "lexical").lower()
        self.neural_model = neural_model
        self._lexical = LexicalReranker()
        self._neural = None

    def rerank(
        self,
        query: str,
        candidates: List[Dict[str, Any]],
        content_key: str = "content",
        top_n: Optional[int] = None,
    ) -> List[Dict[str, Any]]:
        if not candidates:
            return []
        top_n = top_n or len(candidates)

        use_neural = self.mode == "neural"
        if use_neural and self._neural is None:
            try:
                self._neural = NeuralReranker(self.neural_model)
            except Exception as e:
                logger.warning(f"神经重排不可用，回退 lexical: {e}")
                use_neural = False

        scored = []
        for c in candidates:
            content = c.get(content_key, "") or ""
            vec = c.get("score", 0.0) or 0.0
            if use_neural and self._neural is not None:
                s = self._neural.score(query, content)
                item = dict(c)
                item["rerank_score"] = round(float(s), 4)
                item["rerank_mode"] = "neural"
            else:
                lex = self._lexical.score(query, content)
                # 向量分归一化（假设余弦 0~1，阈值约 0.5）
                vec_norm = max(0.0, (vec - 0.5) * 2.0)
                combined = 0.7 * lex + 0.3 * vec_norm
                item = dict(c)
                item["rerank_score"] = round(combined, 4)
                item["rerank_mode"] = "lexical"
            scored.append(item)

        scored.sort(key=lambda x: x["rerank_score"], reverse=True)
        return scored[:top_n]
