"""工作流 PostgreSQL 存储"""
import json
import logging
from typing import Optional, Dict, Any, List, Tuple
from uuid import uuid4
from datetime import datetime

import asyncpg

from config.settings import settings

logger = logging.getLogger(__name__)


class WorkflowDB:
    """工作流数据库访问"""

    def __init__(self):
        self.pool = None

    async def init_pool(self):
        if self.pool:
            return
        try:
            self.pool = await asyncpg.create_pool(
                host=settings.db_host,
                port=settings.db_port,
                user=settings.db_user,
                password=settings.db_password,
                database=settings.db_name,
                min_size=1,
                max_size=5,
                command_timeout=30,
            )
            await self._init_tables()
            logger.info("Workflow database pool initialized")
        except Exception as e:
            logger.error(f"Failed to initialize workflow database pool: {e}")
            raise

    async def _init_tables(self):
        async with self.pool.acquire() as conn:
            await conn.execute("""
                CREATE TABLE IF NOT EXISTS workflows (
                    id UUID PRIMARY KEY,
                    tenant_id UUID NOT NULL,
                    name VARCHAR(255) NOT NULL DEFAULT 'Untitled',
                    description TEXT DEFAULT '',
                    nodes JSONB DEFAULT '[]'::jsonb,
                    edges JSONB DEFAULT '[]'::jsonb,
                    start_node VARCHAR(255) DEFAULT '',
                    is_public BOOLEAN DEFAULT FALSE,
                    status VARCHAR(20) DEFAULT 'draft',
                    created_at TIMESTAMP DEFAULT NOW(),
                    updated_at TIMESTAMP DEFAULT NOW()
                )
            """)
            await conn.execute("""
                CREATE TABLE IF NOT EXISTS workflow_instances (
                    id UUID PRIMARY KEY,
                    workflow_id UUID NOT NULL,
                    tenant_id UUID NOT NULL,
                    user_id VARCHAR(255) DEFAULT '',
                    status VARCHAR(20) DEFAULT 'running',
                    input_data JSONB DEFAULT '{}'::jsonb,
                    output_data JSONB DEFAULT '{}'::jsonb,
                    node_states JSONB DEFAULT '{}'::jsonb,
                    error TEXT DEFAULT '',
                    started_at TIMESTAMP DEFAULT NOW(),
                    completed_at TIMESTAMP,
                    created_at TIMESTAMP DEFAULT NOW()
                )
            """)
            await conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_workflows_tenant ON workflows(tenant_id)
            """)
            await conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_workflow_instances_workflow ON workflow_instances(workflow_id)
            """)
            await conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_workflow_instances_tenant ON workflow_instances(tenant_id)
            """)
            await conn.execute("""
                CREATE TABLE IF NOT EXISTS workflow_triggers (
                    id UUID PRIMARY KEY,
                    workflow_id UUID NOT NULL,
                    tenant_id UUID NOT NULL,
                    trigger_type VARCHAR(20) NOT NULL,
                    node_id VARCHAR(255) NOT NULL,
                    config JSONB DEFAULT '{}'::jsonb,
                    status VARCHAR(20) DEFAULT 'active',
                    created_at TIMESTAMP DEFAULT NOW(),
                    updated_at TIMESTAMP DEFAULT NOW()
                )
            """)
            await conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_wf_triggers_type ON workflow_triggers(trigger_type, status)
            """)
            await conn.execute("""
                CREATE INDEX IF NOT EXISTS idx_wf_triggers_workflow ON workflow_triggers(workflow_id)
            """)

    async def list_workflows(self, tenant_id: str, page: int = 1, page_size: int = 20) -> Tuple[List[Dict], int]:
        await self.init_pool()
        offset = (page - 1) * page_size
        async with self.pool.acquire() as conn:
            total = await conn.fetchval(
                "SELECT COUNT(*) FROM workflows WHERE tenant_id = $1", tenant_id
            )
            rows = await conn.fetch(
                """SELECT id, tenant_id, name, description, nodes, edges, start_node,
                          is_public, status, created_at, updated_at
                   FROM workflows WHERE tenant_id = $1
                   ORDER BY updated_at DESC LIMIT $2 OFFSET $3""",
                tenant_id, page_size, offset
            )
            items = [self._row_to_workflow(r) for r in rows]
            return items, total

    async def get_workflow(self, workflow_id: str, tenant_id: str = None) -> Optional[Dict]:
        await self.init_pool()
        async with self.pool.acquire() as conn:
            if tenant_id:
                row = await conn.fetchrow(
                    "SELECT * FROM workflows WHERE id = $1 AND tenant_id = $2",
                    workflow_id, tenant_id
                )
            else:
                row = await conn.fetchrow(
                    "SELECT * FROM workflows WHERE id = $1", workflow_id
                )
            return self._row_to_workflow(row) if row else None

    async def create_workflow(self, data: Dict[str, Any], tenant_id: str) -> Dict:
        await self.init_pool()
        workflow_id = str(uuid4())
        now = datetime.now()
        nodes = data.get("nodes", [])
        edges = data.get("edges", [])
        async with self.pool.acquire() as conn:
            await conn.execute(
                """INSERT INTO workflows (id, tenant_id, name, description, nodes, edges, start_node, is_public, created_at, updated_at)
                   VALUES ($1, $2, $3, $4, $5::jsonb, $6::jsonb, $7, $8, $9, $10)""",
                workflow_id, tenant_id,
                data.get("name", "Untitled"),
                data.get("description", ""),
                json.dumps(nodes), json.dumps(edges),
                data.get("start_node", ""),
                data.get("is_public", False),
                now, now
            )
        return await self.get_workflow(workflow_id)

    async def update_workflow(self, workflow_id: str, data: Dict[str, Any], tenant_id: str) -> Optional[Dict]:
        await self.init_pool()
        existing = await self.get_workflow(workflow_id, tenant_id)
        if not existing:
            return None

        now = datetime.now()
        sets = ["updated_at = $2"]
        params = [workflow_id, now]
        idx = 3
        for field in ("name", "description", "start_node"):
            if field in data:
                sets.append(f"{field} = ${idx}")
                params.append(data[field])
                idx += 1
        if "is_public" in data:
            sets.append(f"is_public = ${idx}")
            params.append(bool(data["is_public"]))
            idx += 1
        if "status" in data:
            sets.append(f"status = ${idx}")
            params.append(data["status"])
            idx += 1
        if "nodes" in data:
            sets.append(f"nodes = ${idx}::jsonb")
            params.append(json.dumps(data["nodes"]))
            idx += 1
        if "edges" in data:
            sets.append(f"edges = ${idx}::jsonb")
            params.append(json.dumps(data["edges"]))
            idx += 1

        query = f"UPDATE workflows SET {', '.join(sets)} WHERE id = $1 AND tenant_id = ${idx}"
        params.append(tenant_id)
        async with self.pool.acquire() as conn:
            await conn.execute(query, *params)
        return await self.get_workflow(workflow_id)

    async def sync_workflow_triggers(self, workflow_id: str, tenant_id: str, nodes: list, status: str = "active"):
        await self.init_pool()
        async with self.pool.acquire() as conn:
            await conn.execute(
                "DELETE FROM workflow_triggers WHERE workflow_id = $1", workflow_id
            )
            for node in nodes:
                ntype = node.get("type", "")
                if ntype not in ("cron", "webhook"):
                    continue
                trigger_id = str(uuid4())
                now = datetime.now()
                config = node.get("config", {})
                await conn.execute(
                    """INSERT INTO workflow_triggers (id, workflow_id, tenant_id, trigger_type, node_id, config, status, created_at, updated_at)
                       VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)""",
                    trigger_id, workflow_id, tenant_id, ntype, node.get("id", ""),
                    json.dumps(config), status, now, now
                )

    async def list_active_triggers_by_type(self, trigger_type: str) -> List[Dict]:
        await self.init_pool()
        async with self.pool.acquire() as conn:
            rows = await conn.fetch(
                """SELECT t.id, t.workflow_id, t.tenant_id, t.trigger_type, t.node_id, t.config, w.status as wf_status
                   FROM workflow_triggers t
                   JOIN workflows w ON t.workflow_id = w.id
                   WHERE t.trigger_type = $1 AND t.status = 'active' AND w.status = 'published'""",
                trigger_type
            )
            return [{
                "id": str(r["id"]),
                "workflow_id": str(r["workflow_id"]),
                "tenant_id": str(r["tenant_id"]),
                "trigger_type": r["trigger_type"],
                "node_id": r["node_id"],
                "config": r["config"] or {},
            } for r in rows]

    async def get_workflow_with_nodes(self, workflow_id: str) -> Optional[Dict]:
        return await self.get_workflow(workflow_id)

    async def delete_workflow(self, workflow_id: str, tenant_id: str) -> bool:
        await self.init_pool()
        async with self.pool.acquire() as conn:
            result = await conn.execute(
                "DELETE FROM workflows WHERE id = $1 AND tenant_id = $2",
                workflow_id, tenant_id
            )
            await conn.execute(
                "DELETE FROM workflow_instances WHERE workflow_id = $1",
                workflow_id
            )
            await conn.execute(
                "DELETE FROM workflow_triggers WHERE workflow_id = $1",
                workflow_id
            )
            return "DELETE 1" in result

    async def create_instance(self, workflow_id: str, tenant_id: str, user_id: str, input_data: dict) -> Dict:
        await self.init_pool()
        instance_id = str(uuid4())
        now = datetime.now()
        async with self.pool.acquire() as conn:
            await conn.execute(
                """INSERT INTO workflow_instances (id, workflow_id, tenant_id, user_id, status, input_data, started_at, created_at)
                   VALUES ($1, $2, $3, $4, 'running', $5::jsonb, $6, $7)""",
                instance_id, workflow_id, tenant_id, user_id,
                json.dumps(input_data), now, now
            )
        return {
            "instance_id": instance_id,
            "workflow_id": workflow_id,
            "tenant_id": tenant_id,
            "status": "running",
            "input_data": input_data,
            "started_at": now.isoformat(),
        }

    async def update_instance(self, instance_id: str, status: str, output_data: dict = None, error: str = None):
        await self.init_pool()
        now = datetime.now()
        async with self.pool.acquire() as conn:
            if status in ("success", "failed") and output_data is not None:
                await conn.execute(
                    """UPDATE workflow_instances SET status = $1, output_data = $2::jsonb,
                              error = $3, completed_at = $4 WHERE id = $5""",
                    status, json.dumps(output_data) if output_data else "{}",
                    error or "", now, instance_id
                )
            else:
                await conn.execute(
                    "UPDATE workflow_instances SET status = $1 WHERE id = $2",
                    status, instance_id
                )

    async def list_instances(self, workflow_id: str, page: int = 1, page_size: int = 20) -> Tuple[List[Dict], int]:
        await self.init_pool()
        offset = (page - 1) * page_size
        async with self.pool.acquire() as conn:
            total = await conn.fetchval(
                "SELECT COUNT(*) FROM workflow_instances WHERE workflow_id = $1", workflow_id
            )
            rows = await conn.fetch(
                """SELECT id, workflow_id, tenant_id, user_id, status, input_data, output_data, error, started_at, completed_at, created_at
                   FROM workflow_instances WHERE workflow_id = $1
                   ORDER BY created_at DESC LIMIT $2 OFFSET $3""",
                workflow_id, page_size, offset
            )
            items = []
            for r in rows:
                items.append({
                    "instance_id": str(r["id"]),
                    "workflow_id": str(r["workflow_id"]),
                    "tenant_id": str(r["tenant_id"]) if r["tenant_id"] else "",
                    "user_id": r["user_id"] or "",
                    "status": r["status"],
                    "input_data": r["input_data"] or {},
                    "output_data": r["output_data"] or {},
                    "error": r["error"] or "",
                    "started_at": r["started_at"].isoformat() if r["started_at"] else "",
                    "completed_at": r["completed_at"].isoformat() if r["completed_at"] else "",
                    "created_at": r["created_at"].isoformat() if r["created_at"] else "",
                })
            return items, total

    def _row_to_workflow(self, row) -> Optional[Dict]:
        if not row:
            return None
        return {
            "id": str(row["id"]),
            "tenant_id": str(row["tenant_id"]) if row["tenant_id"] else "",
            "name": row["name"],
            "description": row["description"] or "",
            "nodes": row["nodes"] or [],
            "edges": row["edges"] or [],
            "start_node": row["start_node"] or "",
            "is_public": row["is_public"] or False,
            "status": row["status"] or "draft",
            "created_at": row["created_at"].isoformat() if row["created_at"] else "",
            "updated_at": row["updated_at"].isoformat() if row["updated_at"] else "",
        }


workflow_db = WorkflowDB()
