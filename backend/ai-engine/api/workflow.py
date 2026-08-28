"""工作流 API"""
import json
import logging
from typing import Dict, Any, Optional
from uuid import uuid4

from fastapi import APIRouter, HTTPException, Header

from core.workflow.node_types import (
    Workflow, WorkflowInstance,
    Node, Edge, NodeType
)
from core.workflow.workflow_utils import dict_to_workflow
from core.workflow.workflow_engine import workflow_engine
from core.workflow.workflow_db import workflow_db
from core.workflow.scheduler import workflow_scheduler

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/workflows", tags=["workflows"])


@router.get("")
async def list_workflows(page: int = 1, page_size: int = 20,
                          x_tenant_id: Optional[str] = Header(None)):
    if not x_tenant_id:
        return {"items": [], "total": 0, "page": page, "page_size": page_size}
    items, total = await workflow_db.list_workflows(x_tenant_id, page, page_size)
    return {"items": items, "total": total, "page": page, "page_size": page_size}


@router.get("/node-types")
async def get_node_types():
    return [
        {
            "type": "start",
            "name": "开始节点",
            "icon": "▶",
            "description": "工作流入口"
        },
        {
            "type": "llm",
            "name": "LLM调用",
            "icon": "🤖",
            "description": "调用大语言模型生成内容",
            "parameters": {
                "model": {"type": "string", "description": "模型名称"},
                "temperature": {"type": "number", "description": "温度参数"},
                "max_tokens": {"type": "integer", "description": "最大Token数"},
                "prompt": {"type": "string", "description": "提示词模板"},
                "output_key": {"type": "string", "description": "输出变量名"}
            }
        },
        {
            "type": "rag",
            "name": "知识库检索",
            "icon": "📚",
            "description": "从知识库检索相关内容",
            "parameters": {
                "knowledge_base_id": {"type": "string", "description": "知识库ID"},
                "strategy": {
                    "type": "string",
                    "description": "检索策略",
                    "enum": ["simple", "multi_query", "compression"]
                },
                "top_k": {"type": "integer", "description": "返回文档数"},
                "output_key": {"type": "string", "description": "输出变量名"}
            }
        },
        {
            "type": "tool",
            "name": "工具调用",
            "icon": "🔧",
            "description": "调用外部工具",
            "parameters": {
                "tool_name": {"type": "string", "description": "工具名称"},
                "parameters": {"type": "object", "description": "工具参数"},
                "output_key": {"type": "string", "description": "输出变量名"}
            }
        },
        {
            "type": "condition",
            "name": "条件判断",
            "icon": "❓",
            "description": "根据条件分支执行，true分支走sourceHandle=true的边，false分支走sourceHandle=false的边",
            "parameters": {
                "condition": {"type": "string", "description": "条件表达式，如 {{condition_result}} == true"}
            }
        },
        {
            "type": "http",
            "name": "HTTP请求",
            "icon": "🌐",
            "description": "发送HTTP请求调用外部API",
            "parameters": {
                "url": {"type": "string", "description": "请求URL"},
                "method": {"type": "string", "description": "请求方法", "enum": ["GET", "POST", "PUT", "DELETE"]},
                "headers": {"type": "object", "description": "请求头"},
                "body": {"type": "string", "description": "请求体JSON"},
                "output_key": {"type": "string", "description": "输出变量名"}
            }
        },
        {
            "type": "cron",
            "name": "定时触发",
            "icon": "⏰",
            "description": "按 Cron 表达式定时触发工作流",
            "parameters": {
                "cron": {"type": "string", "description": "Cron 表达式，如 0 9 * * *（每天9点）"},
                "timezone": {"type": "string", "description": "时区，如 Asia/Shanghai"}
            }
        },
        {
            "type": "webhook",
            "name": "Webhook触发",
            "icon": "🔗",
            "description": "通过 HTTP POST 触发工作流",
            "parameters": {
                "path": {"type": "string", "description": "Webhook 路径（可选）"},
                "secret": {"type": "string", "description": "签名密钥（可选）"}
            }
        },
        {
            "type": "end",
            "name": "结束节点",
            "icon": "⏹",
            "description": "工作流结束"
        }
    ]


@router.post("/validate")
async def validate_workflow(data: Dict[str, Any]):
    nodes = data.get("nodes", [])
    edges = data.get("edges", [])

    if not nodes:
        return {"valid": False, "message": "至少需要一个节点"}

    start_nodes = [n for n in nodes if n.get("type") == "start"]
    if not start_nodes:
        return {"valid": False, "message": "缺少开始节点"}
    if len(start_nodes) > 1:
        return {"valid": False, "message": "只能有一个开始节点"}

    end_nodes = [n for n in nodes if n.get("type") == "end"]
    if not end_nodes:
        return {"valid": False, "message": "缺少结束节点"}

    node_ids = {n.get("id") for n in nodes}
    for edge in edges:
        if edge.get("source") not in node_ids:
            return {"valid": False, "message": f"边引用了不存在的源节点: {edge.get('source')}"}
        if edge.get("target") not in node_ids:
            return {"valid": False, "message": f"边引用了不存在的目标节点: {edge.get('target')}"}

    return {"valid": True, "message": "工作流配置有效"}


@router.get("/{workflow_id}")
async def get_workflow(workflow_id: str,
                       x_tenant_id: Optional[str] = Header(None)):
    tenant = x_tenant_id or "default"
    workflow = await workflow_db.get_workflow(workflow_id, tenant)
    if not workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")
    return workflow


@router.post("")
async def create_workflow(data: Dict[str, Any],
                          x_tenant_id: Optional[str] = Header(None)):
    tenant_id = x_tenant_id or "default"
    workflow = await workflow_db.create_workflow(data, tenant_id)
    nodes = data.get("nodes", [])
    status = data.get("status", "draft")
    await workflow_db.sync_workflow_triggers(workflow["id"], tenant_id, nodes, "active" if status == "published" else "draft")
    try:
        await workflow_scheduler.reload_triggers()
    except Exception:
        pass
    return workflow


@router.put("/{workflow_id}")
async def update_workflow(workflow_id: str, data: Dict[str, Any],
                          x_tenant_id: Optional[str] = Header(None)):
    tenant_id = x_tenant_id or "default"
    workflow = await workflow_db.update_workflow(workflow_id, data, tenant_id)
    if not workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")
    if "nodes" in data or "status" in data:
        nodes = data.get("nodes", workflow.get("nodes", []))
        status = data.get("status", workflow.get("status", "draft"))
        await workflow_db.sync_workflow_triggers(workflow_id, tenant_id, nodes, "active" if status == "published" else "draft")
        try:
            await workflow_scheduler.reload_triggers()
        except Exception:
            pass
    return workflow


@router.delete("/{workflow_id}")
async def delete_workflow(workflow_id: str,
                          x_tenant_id: Optional[str] = Header(None)):
    tenant_id = x_tenant_id or "default"
    deleted = await workflow_db.delete_workflow(workflow_id, tenant_id)
    if not deleted:
        raise HTTPException(status_code=404, detail="Workflow not found")
    try:
        await workflow_scheduler.reload_triggers()
    except Exception:
        pass
    return {"message": "Workflow deleted"}


@router.post("/{workflow_id}/execute")
async def execute_workflow(workflow_id: str, data: Dict[str, Any],
                           x_tenant_id: Optional[str] = Header(None),
                           x_user_id: Optional[str] = Header(None)):
    tenant_id = x_tenant_id or "default"
    user_id = x_user_id or ""

    # 技能桥接：加载租户已安装技能为可执行工具（失败不阻断）
    try:
        from core.tools.skill_bridge import ensure_tenant_skills
        await ensure_tenant_skills(tenant_id)
    except Exception:
        pass

    try:
        db_workflow = await workflow_db.get_workflow(workflow_id, tenant_id)
        if not db_workflow:
            logger.error(f"Workflow not found: {workflow_id} tenant: {tenant_id}")
            raise HTTPException(status_code=404, detail="Workflow not found")

        logger.info(f"Found workflow: {workflow_id} nodes: {len(db_workflow.get('nodes', []))} edges: {len(db_workflow.get('edges', []))} start_node: {db_workflow.get('start_node')}")

        workflow = dict_to_workflow(db_workflow)
        input_data = data.get("input_data", {})

        logger.info(f"Workflow execution started: {workflow_id} input: {input_data}")

        instance_data = await workflow_db.create_instance(
            workflow_id, tenant_id, user_id, input_data
        )

        instance = WorkflowInstance(
            instance_id=instance_data["instance_id"],
            workflow_id=workflow_id,
            tenant_id=tenant_id,
            user_id=user_id,
            input_data=input_data,
        )

        result = await workflow_engine.execute(workflow, instance)

        logger.info(f"Workflow execution completed: {workflow_id} success: {result.success} duration: {result.duration} error: {result.error}")

        if result.success:
            await workflow_db.update_instance(
                instance_data["instance_id"], "success", result.output
            )
        else:
            await workflow_db.update_instance(
                instance_data["instance_id"], "failed", result.output, result.error
            )

        return {
            "success": result.success,
            "output": result.output,
            "error": result.error,
            "instance_id": instance_data["instance_id"],
            "duration": result.duration,
        }
    except Exception as e:
        logger.error(f"Workflow execution error: {workflow_id} error: {e}", exc_info=True)
        raise HTTPException(status_code=500, detail=f"Workflow execution failed: {str(e)}")


@router.post("/{workflow_id}/stream")
async def stream_execute_workflow(workflow_id: str, data: Dict[str, Any],
                                  x_tenant_id: Optional[str] = Header(None),
                                  x_user_id: Optional[str] = Header(None)):
    tenant_id = x_tenant_id or "default"
    user_id = x_user_id or ""

    # 技能桥接
    try:
        from core.tools.skill_bridge import ensure_tenant_skills
        await ensure_tenant_skills(tenant_id)
    except Exception:
        pass

    db_workflow = await workflow_db.get_workflow(workflow_id, tenant_id)
    if not db_workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")

    workflow = dict_to_workflow(db_workflow)
    input_data = data.get("input_data", {})

    instance_data = await workflow_db.create_instance(
        workflow_id, tenant_id, user_id, input_data
    )

    instance = WorkflowInstance(
        instance_id=instance_data["instance_id"],
        workflow_id=workflow_id,
        tenant_id=tenant_id,
        user_id=user_id,
        input_data=input_data,
    )

    events = []
    async for event in workflow_engine.stream_execute(workflow, instance):
        events.append(event)
        if event.get("is_final"):
            break

    await workflow_db.update_instance(
        instance_data["instance_id"], "success", {"events": events}
    )

    return {"events": events, "instance_id": instance_data["instance_id"]}


@router.get("/{workflow_id}/instances")
async def list_instances(workflow_id: str, page: int = 1, page_size: int = 20):
    items, total = await workflow_db.list_instances(workflow_id, page, page_size)
    return {"items": items, "total": total, "page": page, "page_size": page_size}


@router.post("/webhook/{workflow_id}")
async def webhook_trigger(workflow_id: str, data: Dict[str, Any]):
    """外部系统通过 Webhook 触发工作流执行"""
    db_workflow = await workflow_db.get_workflow(workflow_id)
    if not db_workflow:
        raise HTTPException(status_code=404, detail="Workflow not found")

    nodes = db_workflow.get("nodes", [])
    has_webhook = any(n.get("type") == "webhook" for n in nodes)
    if not has_webhook:
        raise HTTPException(status_code=400, detail="Workflow has no webhook trigger")

    workflow = dict_to_workflow(db_workflow)
    tenant_id = db_workflow.get("tenant_id", "default")

    instance_data = await workflow_db.create_instance(
        workflow_id, tenant_id, "webhook", data
    )

    instance = WorkflowInstance(
        instance_id=instance_data["instance_id"],
        workflow_id=workflow_id,
        tenant_id=tenant_id,
        user_id="webhook",
        input_data=data,
    )

    result = await workflow_engine.execute(workflow, instance)

    if result.success:
        await workflow_db.update_instance(
            instance_data["instance_id"], "success", result.output
        )
    else:
        await workflow_db.update_instance(
            instance_data["instance_id"], "failed", result.output, result.error
        )

    return {
        "success": result.success,
        "output": result.output,
        "error": result.error,
        "instance_id": instance_data["instance_id"],
        "duration": result.duration,
    }
