"""Cron 定时调度器"""
import json
import logging
from datetime import datetime, timezone
from typing import Optional

from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger
from croniter import croniter

from .node_types import WorkflowInstance
from .workflow_engine import workflow_engine
from .workflow_db import workflow_db

logger = logging.getLogger(__name__)


class WorkflowScheduler:

    def __init__(self):
        self.scheduler: Optional[AsyncIOScheduler] = None
        self._job_map: dict = {}

    async def start(self):
        if self.scheduler and self.scheduler.running:
            return
        self.scheduler = AsyncIOScheduler()
        self.scheduler.add_job(
            self._sync_cron_jobs, "interval", seconds=30,
            id="_cron_sync", name="Sync cron triggers"
        )
        self.scheduler.start()
        await self._sync_cron_jobs()
        logger.info("Workflow scheduler started")

    async def stop(self):
        if self.scheduler:
            self.scheduler.shutdown(wait=False)
            self._job_map.clear()
            logger.info("Workflow scheduler stopped")

    async def reload_triggers(self):
        await self._sync_cron_jobs()

    async def _sync_cron_jobs(self):
        try:
            triggers = await workflow_db.list_active_triggers_by_type("cron")
            active_ids = {t["id"] for t in triggers}

            for job_id in list(self._job_map.keys()):
                if job_id not in active_ids:
                    try:
                        self.scheduler.remove_job(job_id)
                    except Exception:
                        pass
                    del self._job_map[job_id]
                    logger.info(f"Removed cron job: {job_id}")

            for t in triggers:
                tid = t["id"]
                if tid in self._job_map:
                    continue
                config = t["config"] or {}
                cron_expr = config.get("cron", "0 9 * * *")
                timezone_str = config.get("timezone", "Asia/Shanghai")

                try:
                    parts = cron_expr.strip().split()
                    if len(parts) == 5:
                        self.scheduler.add_job(
                            self._execute_workflow,
                            CronTrigger(
                                minute=parts[0], hour=parts[1],
                                day=parts[2], month=parts[3], day_of_week=parts[4],
                                timezone=timezone_str
                            ),
                            args=[t["workflow_id"], t["tenant_id"], t["node_id"]],
                            id=tid, name=f"cron-{t['workflow_id'][:8]}",
                            replace_existing=True
                        )
                        self._job_map[tid] = True
                        logger.info(f"Registered cron job {tid}: {cron_expr}")
                except Exception as e:
                    logger.error(f"Failed to register cron job {tid}: {e}")

        except Exception as e:
            logger.error(f"Sync cron jobs failed: {e}")

    async def _execute_workflow(self, workflow_id: str, tenant_id: str, node_id: str):
        try:
            db_workflow = await workflow_db.get_workflow(workflow_id, tenant_id)
            if not db_workflow:
                logger.warning(f"Cron trigger: workflow {workflow_id} not found")
                return

            from .workflow_utils import dict_to_workflow
            workflow = dict_to_workflow(db_workflow)

            now_str = datetime.now(timezone.utc).isoformat()
            input_data = {
                "trigger": {
                    "type": "cron",
                    "fired_at": now_str,
                    "node_id": node_id,
                }
            }

            instance_data = await workflow_db.create_instance(
                workflow_id, tenant_id, "system", input_data
            )

            instance = WorkflowInstance(
                instance_id=instance_data["instance_id"],
                workflow_id=workflow_id,
                tenant_id=tenant_id,
                user_id="system",
                input_data=input_data,
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

            logger.info(f"Cron workflow {workflow_id} executed: success={result.success}")
        except Exception as e:
            logger.error(f"Cron workflow execution failed {workflow_id}: {e}")


workflow_scheduler = WorkflowScheduler()
