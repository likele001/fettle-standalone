"""意图分类器 - 规则引擎 + 模型分类"""
import json
import logging
import re
from typing import Dict, Any, Optional, List
from dataclasses import dataclass, field

logger = logging.getLogger(__name__)


@dataclass
class IntentRule:
    """意图规则"""
    id: str
    name: str
    keywords: List[str]
    pattern: Optional[str] = None
    intent: str = ""
    priority: int = 10


class RuleEngine:
    """规则引擎"""

    def __init__(self):
        self.rules: List[IntentRule] = []
        self._load_default_rules()

    def _load_default_rules(self):
        """加载默认规则"""
        self.rules = [
            IntentRule(
                id="greeting_001",
                name="问候语",
                keywords=["你好", "您好", "hi", "hello", "在吗", "在不在"],
                intent="greeting",
                priority=10,
            ),
            IntentRule(
                id="farewell_001",
                name="告别语",
                keywords=["再见", "拜拜", "bye", "goodbye", "谢谢", "感谢"],
                intent="farewell",
                priority=10,
            ),
            IntentRule(
                id="query_order_001",
                name="查询订单",
                keywords=["订单", "查订单", "我的订单", "订单状态", "物流"],
                pattern=r"(查询|查一下|看看).*(订单|物流)",
                intent="query_order",
                priority=20,
            ),
            IntentRule(
                id="query_inventory_001",
                name="查询库存",
                keywords=["库存", "有货", "还有吗", "库存多少"],
                intent="query_inventory",
                priority=20,
            ),
            IntentRule(
                id="complaint_001",
                name="投诉",
                keywords=["投诉", "差评", "不满意", "太慢了", "质量差", "骗子"],
                intent="complaint",
                priority=30,
            ),
            IntentRule(
                id="transfer_human_001",
                name="转人工",
                keywords=["人工", "客服", "真人", "转接", "找客服"],
                intent="transfer_human",
                priority=50,
            ),
        ]

    def match(self, text: str) -> Optional[IntentRule]:
        """匹配规则"""
        text_lower = text.lower()

        # 按优先级排序
        sorted_rules = sorted(self.rules, key=lambda r: r.priority, reverse=True)

        # 先匹配关键词
        for rule in sorted_rules:
            for keyword in rule.keywords:
                if keyword.lower() in text_lower:
                    return rule

        # 再匹配正则
        for rule in sorted_rules:
            if rule.pattern:
                if re.search(rule.pattern, text):
                    return rule

        return None

    def add_rule(self, rule: IntentRule):
        """添加规则"""
        self.rules.append(rule)

    def remove_rule(self, rule_id: str):
        """移除规则"""
        self.rules = [r for r in self.rules if r.id != rule_id]


@dataclass
class IntentResult:
    """意图识别结果"""
    intent: str
    confidence: float
    entities: Dict[str, str] = field(default_factory=dict)
    source: str = "rule"  # rule, model


class IntentClassifier:
    """意图分类器"""

    def __init__(self):
        self.rule_engine = RuleEngine()

    def classify(self, text: str, tenant_id: str = "") -> IntentResult:
        """分类意图"""
        # 1. 先尝试规则匹配
        rule = self.rule_engine.match(text)
        if rule:
            entities = self._extract_entities(text, rule.intent)
            return IntentResult(
                intent=rule.intent,
                confidence=0.95,
                entities=entities,
                source="rule",
            )

        # 2. 规则未命中，返回 unknown（模型分类由 gRPC 层处理）
        return IntentResult(
            intent="unknown",
            confidence=0.5,
            entities={},
            source="model",
        )

    def _extract_entities(self, text: str, intent: str) -> Dict[str, str]:
        """提取实体"""
        entities = {}

        if intent == "query_order":
            order_id = self._extract_order_id(text)
            if order_id:
                entities["order_id"] = order_id

        if intent == "query_inventory":
            product = self._extract_product(text)
            if product:
                entities["product"] = product

        return entities

    def _extract_order_id(self, text: str) -> str:
        """提取订单号"""
        patterns = [
            r"ORD\d{10,}",
            r"订单号[：:]\s*(\d+)",
            r"(\d{10,})",
        ]
        for pattern in patterns:
            match = re.search(pattern, text)
            if match:
                return match.group(1) if match.lastindex else match.group(0)
        return ""

    def _extract_product(self, text: str) -> str:
        """提取商品名"""
        pattern = r"(商品|产品)[：:]\s*([^\s,，。]+)"
        match = re.search(pattern, text)
        if match:
            return match.group(2)
        return ""


# 全局实例
intent_classifier = IntentClassifier()
