package main

import (
	"fmt"
	"regexp"
	"strings"
)

// IntentRule 意图规则
type IntentRule struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
	Pattern  string   `json:"pattern"` // 正则表达式
	Intent   string   `json:"intent"`
	Priority int      `json:"priority"` // 优先级，数字越大优先级越高
}

// RuleEngine 规则引擎
type RuleEngine struct {
	rules []*IntentRule
}

// NewRuleEngine 创建规则引擎
func NewRuleEngine() *RuleEngine {
	engine := &RuleEngine{
		rules: make([]*IntentRule, 0),
	}
	engine.loadDefaultRules()
	return engine
}

// loadDefaultRules 加载默认规则
func (e *RuleEngine) loadDefaultRules() {
	defaultRules := []*IntentRule{
		// 问候
		{
			ID:       "greeting_001",
			Name:     "问候语",
			Keywords: []string{"你好", "您好", "hi", "hello", "在吗", "在不在"},
			Intent:   "greeting",
			Priority: 10,
		},
		// 告别
		{
			ID:       "farewell_001",
			Name:     "告别语",
			Keywords: []string{"再见", "拜拜", "bye", "goodbye", "谢谢", "感谢"},
			Intent:   "farewell",
			Priority: 10,
		},
		// 查询订单
		{
			ID:       "query_order_001",
			Name:     "查询订单",
			Keywords: []string{"订单", "查订单", "我的订单", "订单状态", "物流"},
			Pattern:  `(查询|查一下|看看).*(订单|物流)`,
			Intent:   "query_order",
			Priority: 20,
		},
		// 查询库存
		{
			ID:       "query_inventory_001",
			Name:     "查询库存",
			Keywords: []string{"库存", "有货", "还有吗", "库存多少"},
			Intent:   "query_inventory",
			Priority: 20,
		},
		// 投诉
		{
			ID:       "complaint_001",
			Name:     "投诉",
			Keywords: []string{"投诉", "差评", "不满意", "太慢了", "质量差", "骗子"},
			Intent:   "complaint",
			Priority: 30,
		},
		// 转人工
		{
			ID:       "transfer_human_001",
			Name:     "转人工",
			Keywords: []string{"人工", "客服", "真人", "转接", "找客服"},
			Intent:   "transfer_human",
			Priority: 50,
		},
	}

	e.rules = defaultRules
}

// Match 匹配规则
func (e *RuleEngine) Match(text string) (*IntentRule, bool) {
	text = strings.ToLower(text)

	// 按优先级排序
	sortedRules := make([]*IntentRule, len(e.rules))
	copy(sortedRules, e.rules)
	for i := 0; i < len(sortedRules); i++ {
		for j := i + 1; j < len(sortedRules); j++ {
			if sortedRules[i].Priority < sortedRules[j].Priority {
				sortedRules[i], sortedRules[j] = sortedRules[j], sortedRules[i]
			}
		}
	}

	// 先匹配关键词
	for _, rule := range sortedRules {
		for _, keyword := range rule.Keywords {
			if strings.Contains(text, strings.ToLower(keyword)) {
				return rule, true
			}
		}
	}

	// 再匹配正则
	for _, rule := range sortedRules {
		if rule.Pattern != "" {
			if matched, _ := regexp.MatchString(rule.Pattern, text); matched {
				return rule, true
			}
		}
	}

	return nil, false
}

// AddRule 添加规则
func (e *RuleEngine) AddRule(rule *IntentRule) {
	e.rules = append(e.rules, rule)
}

// RemoveRule 移除规则
func (e *RuleEngine) RemoveRule(ruleID string) {
	for i, rule := range e.rules {
		if rule.ID == ruleID {
			e.rules = append(e.rules[:i], e.rules[i+1:]...)
			return
		}
	}
}

// IntentClassifier 意图分类器
type IntentClassifier struct {
	ruleEngine *RuleEngine
	// TODO: 集成模型分类器
}

// NewIntentClassifier 创建意图分类器
func NewIntentClassifier() *IntentClassifier {
	return &IntentClassifier{
		ruleEngine: NewRuleEngine(),
	}
}

// IntentResult 意图识别结果
type IntentResult struct {
	Intent     string             `json:"intent"`
	Confidence float64            `json:"confidence"`
	Entities   map[string]string  `json:"entities"`
	Source     string             `json:"source"` // rule, model
}

// Classify 分类意图
func (c *IntentClassifier) Classify(text string, tenantID string) (*IntentResult, error) {
	// 1. 先尝试规则匹配
	if rule, matched := c.ruleEngine.Match(text); matched {
		entities := c.extractEntities(text, rule.Intent)
		return &IntentResult{
			Intent:     rule.Intent,
			Confidence: 0.95,
			Entities:   entities,
			Source:     "rule",
		}, nil
	}

	// 2. 规则未命中，调用模型分类
	// TODO: 调用 AI Engine 的意图识别接口
	return &IntentResult{
		Intent:     "unknown",
		Confidence: 0.5,
		Entities:   make(map[string]string),
		Source:     "model",
	}, nil
}

// extractEntities 提取实体
func (c *IntentClassifier) extractEntities(text string, intent string) map[string]string {
	entities := make(map[string]string)

	switch intent {
	case "query_order":
		// 提取订单号
		if orderID := extractOrderID(text); orderID != "" {
			entities["order_id"] = orderID
		}
	case "query_inventory":
		// 提取商品名
		if product := extractProduct(text); product != "" {
			entities["product"] = product
		}
	}

	return entities
}

// extractOrderID 提取订单号
func extractOrderID(text string) string {
	// 匹配订单号模式：ORD + 数字 或 纯数字订单号
	patterns := []string{
		`ORD\d{10,}`,
		`订单号[：:]\s*(\d+)`,
		`(\d{10,})`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, text); matched {
			re := regexp.MustCompile(pattern)
			if matches := re.FindStringSubmatch(text); len(matches) > 1 {
				return matches[1]
			}
			return re.FindString(text)
		}
	}
	return ""
}

// extractProduct 提取商品名
func extractProduct(text string) string {
	// 简单实现：提取"商品"、"产品"后面的内容
	pattern := `(商品|产品)[：:]\s*([^\s,，。]+)`
	if matched, _ := regexp.MatchString(pattern, text); matched {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 2 {
			return matches[2]
		}
	}
	return ""
}

// 辅助函数：正则匹配
func regexpMatchString(pattern, text string) (bool, error) {
	return regexp.MatchString(pattern, text)
}
