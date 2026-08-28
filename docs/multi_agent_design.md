# 多 Agent 协作（A2A）设计方案

> 状态：待实施（2026-08-20 设计定稿）
> 目标：补齐 2026 基线差距「多 Agent 协作 / A2A」——实现"总控 Agent → 专职 Agent 分工协作"

## 1. 现状

- 工作流引擎 11 节点（LLM/TOOL/RAG/CONDITION/HTTP/CRON/WEBHOOK/INPUT/START/END/TEXT_OUTPUT），`workflow_engine.py` 图遍历执行。
- Agent 有 `WorkflowID` 字段：主管 Agent 对话前 `POST /workflows/{id}/execute` 触发子流程，输出注入 prompt（`workflow_client.go` / chat 链路）。
- **缺口**：无"独立 Agent 节点"——不能在一个工作流里编排多个各带人设/知识库/模型的 Agent 分工协作，无 A2A 互相调用。

## 2. 方案选型

| 方案 | 做法 | 优劣 |
|---|---|---|
| **A（推荐）** | 工作流新增 **AGENT 节点**：节点配置 `{agent_id, input_template, output_key}`，执行时调子 Agent 对话（复用 chat 链路），输出写入 context | 复用成熟的工作流编排（分支/并行/CRON）；改动集中在 ai-engine |
| B | Agent 结构体加 `sub_agent_ids`，主管对话时并行调子 Agent | 编排能力弱（无条件分支），侵入 chat-service 主链路，不推荐 |
| C | 独立 A2A 协议服务（HTTP/gRPC 互通） | 对标 Coze 项目空间，但工作量大，适合后续 |

**采用方案 A**：工作流 = 协作图，AGENT 节点 = 专职岗位。对齐 Dify/Coze 的主流模型。

## 3. 架构设计（方案 A）

```
工作流（主管编排）
  ┌─ START ─┬─ AGENT(客服Agent) ──┐
  │         ├─ AGENT(订单Agent) ──┼─ CONDITION ── TEXT_OUTPUT/LLM 汇总
  │         └─ AGENT(售后Agent) ──┘
```

- **AGENT 节点配置**：
  ```json
  {
    "type": "AGENT",
    "config": {
      "agent_id": "<uuid>",          // 目标 Agent（agent-service 注册）
      "input_template": "{{context.user_input}}",  // 传入子 Agent 的 prompt 模板
      "output_key": "agent_reply_1", // 输出写入 context 的 key
      "timeout": 60
    }
  }
  ```
- **执行器** `workflow_agent_executor.py`：HTTP 调 `agent-service POST /agents/{id}/test`（已有测试对话接口）或 chat-service 对话链路，带 `{input, agent_id, tenant_id, history?}`；子 Agent 自己的知识库/模型/工作流由其自身配置生效。
- **输出**：子 Agent 回复写入 `state["context"][output_key]`，后续 LLM/CONDITION 节点可引用。
- **计费**：子 Agent 调用经 ai-engine chat 链路 → 已统一走 billing-service（本轮计费收敛的成果直接复用）。
- **安全**：防止循环（节点深度限制 + 访问记录），agent_id 校验租户归属。

## 4. 实施步骤（下轮执行）

| # | 文件 | 改动 |
|---|---|---|
| 1 | `ai-engine/core/workflow/node_types.py` | 加 `AGENT` 节点类型（校验 agent_id/output_key） |
| 2 | `ai-engine/core/workflow/workflow_engine.py` | 分发 AGENT 节点 → `workflow_agent_executor`；失败降级（模拟回复 + warning），与现有 TOOL/RAG 降级一致 |
| 3 | 新建 `ai-engine/core/workflow/workflow_agent_executor.py` | 调 agent-service `/agents/{id}/test`（或 ai-engine 对话链路），超时/异常兜底 |
| 4 | `ai-engine/api/workflow.py` | 校验节点配置时支持 AGENT |
| 5 | 前端 `web-admin` 工作流编辑器 | 加 AGENT 节点拖拽 + agent_id 选择器（调 `/api/v1/agents`） |
| 6 | 验证 | 搭建「总控 + 2 专职」工作流，复杂任务拆分回流；子 Agent 调用计入 billing |

## 5. 依赖与风险

- 依赖：agent-service `/agents/{id}/test` 接口存在（`agent_service.go` TestChat）✓
- 风险：子 Agent 触发自己的工作流可能递归 → 执行深度上限（默认 5 层）+ 租户配额
- 可选增强：AGENT 节点支持并行（多个子 Agent 并发调用）、A2A 协议（后续）

## 6. 验收标准

- [ ] 工作流编辑器可拖入 AGENT 节点并选择子 Agent
- [ ] 执行时子 Agent 返回真实回复并写入 context
- [ ] 分支/汇总节点可消费多个子 Agent 输出
- [ ] 子 Agent 对话产生 billing usage 记录
- [ ] 循环/深度超限被拦截
