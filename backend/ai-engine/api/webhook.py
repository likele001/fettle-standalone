"""外部 Webhook 触发器（无需认证）"""
import logging
from typing import Dict, Any

from fastapi import APIRouter, HTTPException

from core.workflow.node_types import WorkflowInstance
from core.workflow.workflow_utils import dict_to_workflow
from core.workflow.workflow_engine import workflow_engine
from core.workflow.workflow_db import workflow_db

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/webhook", tags=["webhook"])


@router.post("/trigger/{workflow_id}")
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
