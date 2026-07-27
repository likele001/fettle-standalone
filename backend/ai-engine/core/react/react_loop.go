package react

import (
	"context"
	"fmt"
	"time"

	"ai-platform/ai-engine/core/models"
	"ai-platform/ai-engine/core/tools"
)

// ReactLoop ReAct 循环核心
type ReactLoop struct {
	modelRouter  *models.ModelRouter
	toolExecutor *tools.ToolExecutor
	maxSteps     int
	timeout      time.Duration
}

// NewReactLoop 创建 ReAct 循环
func NewReactLoop(modelRouter *models.ModelRouter, toolExecutor *tools.ToolExecutor) *ReactLoop {
	return &ReactLoop{
		modelRouter:  modelRouter,
		toolExecutor: toolExecutor,
		maxSteps:     15,
		timeout:      5 * time.Minute,
	}
}

// Task 任务定义
type Task struct {
	ID             string            `json:"id"`
	TenantID       string            `json:"tenant_id"`
	UserID         string            `json:"user_id"`
	AgentID        string            `json:"agent_id"`
	Intent         string            `json:"intent"`
	Entities       map[string]string `json:"entities"`
	Goal           string            `json:"goal"`
	AvailableTools []string          `json:"available_tools"`
	Context        map[string]string `json:"context"`
}

// Step 执行步骤
type Step struct {
	StepNum      int               `json:"step_num"`
	Thought      string            `json:"thought"`       // 思考过程
	Action       string            `json:"action"`        // 工具名称
	ActionInput  map[string]string `json:"action_input"`  // 工具输入
	Observation  string            `json:"observation"`   // 观察结果
	IsFinal      bool              `json:"is_final"`      // 是否最终步骤
	FinalAnswer  string            `json:"final_answer"`  // 最终答案
}

// TaskResult 任务执行结果
type TaskResult struct {
	Success      bool     `json:"success"`
	Steps        []*Step  `json:"steps"`
	FinalAnswer  string   `json:"final_answer"`
	TotalTokens  int      `json:"total_tokens"`
	Duration     float64  `json:"duration"` // 秒
}

// Execute 执行 ReAct 循环
func (r *ReactLoop) Execute(ctx context.Context, task *Task) (*TaskResult, error) {
	startTime := time.Now()
	
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	result := &TaskResult{
		Success: false,
		Steps:   make([]*Step, 0),
	}

	// 生成初始计划
	plan, err := r.generatePlan(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// 执行每个步骤
	for i, step := range plan {
		if i >= r.maxSteps {
			break
		}

		step.StepNum = i + 1

		// 如果是最终步骤，直接返回答案
		if step.IsFinal {
			result.Steps = append(result.Steps, step)
			result.FinalAnswer = step.FinalAnswer
			result.Success = true
			break
		}

		// 执行工具调用
		observation, err := r.executeStep(ctx, step, task)
		if err != nil {
			step.Observation = fmt.Sprintf("Error: %v", err)
		} else {
			step.Observation = observation
		}

		result.Steps = append(result.Steps, step)

		// 反思：是否需要调整计划
		if r.shouldAdjustPlan(step) {
			newPlan, err := r.adjustPlan(ctx, task, result.Steps)
			if err == nil {
				plan = newPlan
			}
		}
	}

	result.Duration = time.Since(startTime).Seconds()
	return result, nil
}

// generatePlan 生成执行计划
func (r *ReactLoop) generatePlan(ctx context.Context, task *Task) ([]*Step, error) {
	// 构建提示词
	prompt := r.buildPlanPrompt(task)

	// 调用模型生成计划
	provider, err := r.modelRouter.SelectProvider("planning")
	if err != nil {
		return nil, err
	}

	response, err := provider.Chat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析计划
	steps, err := r.parsePlan(response.Content)
	if err != nil {
		return nil, err
	}

	return steps, nil
}

// executeStep 执行单个步骤
func (r *ReactLoop) executeStep(ctx context.Context, step *Step, task *Task) (string, error) {
	// 调用工具
	result, err := r.toolExecutor.Execute(ctx, step.Action, step.ActionInput, task.TenantID)
	if err != nil {
		return "", err
	}

	return result, nil
}

// shouldAdjustPlan 判断是否需要调整计划
func (r *ReactLoop) shouldAdjustPlan(step *Step) bool {
	// 如果观察结果包含错误，可能需要调整
	if len(step.Observation) > 0 && step.Observation[:6] == "Error:" {
		return true
	}
	return false
}

// adjustPlan 调整计划
func (r *ReactLoop) adjustPlan(ctx context.Context, task *Task, executedSteps []*Step) ([]*Step, error) {
	// 构建反思提示词
	prompt := r.buildReflectionPrompt(task, executedSteps)

	// 调用模型重新规划
	provider, err := r.modelRouter.SelectProvider("planning")
	if err != nil {
		return nil, err
	}

	response, err := provider.Chat(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// 解析新计划
	steps, err := r.parsePlan(response.Content)
	if err != nil {
		return nil, err
	}

	return steps, nil
}

// buildPlanPrompt 构建计划提示词
func (r *ReactLoop) buildPlanPrompt(task *Task) string {
	prompt := fmt.Sprintf(`你是一个任务规划助手。请根据用户目标和可用工具，制定执行计划。

用户目标: %s
意图: %s
实体: %v
可用工具: %v

请按以下格式输出计划（JSON）:
[
  {
    "thought": "思考过程",
    "action": "工具名称",
    "action_input": {"参数": "值"},
    "is_final": false
  },
  {
    "thought": "最终思考",
    "is_final": true,
    "final_answer": "最终答案"
  }
]
`, task.Goal, task.Intent, task.Entities, task.AvailableTools)

	return prompt
}

// buildReflectionPrompt 构建反思提示词
func (r *ReactLoop) buildReflectionPrompt(task *Task, executedSteps []*Step) string {
	prompt := fmt.Sprintf(`你是一个任务规划助手。之前的执行计划遇到了问题，请重新规划。

用户目标: %s
已执行步骤:
`, task.Goal)

	for _, step := range executedSteps {
		prompt += fmt.Sprintf("步骤 %d:\n", step.StepNum)
		prompt += fmt.Sprintf("  思考: %s\n", step.Thought)
		prompt += fmt.Sprintf("  动作: %s\n", step.Action)
		prompt += fmt.Sprintf("  观察: %s\n", step.Observation)
	}

	prompt += "\n请重新制定计划（JSON格式）:"

	return prompt
}

// parsePlan 解析计划
func (r *ReactLoop) parsePlan(content string) ([]*Step, error) {
	// TODO: 解析 JSON 格式的计划
	// 这里简化处理
	steps := []*Step{
		{
			Thought:     "解析失败，使用默认步骤",
			IsFinal:     true,
			FinalAnswer: content,
		},
	}
	return steps, nil
}
