"""工作流状态持久化"""
import logging
from typing import Dict, Any, Optional
import sqlite3
import json
from pathlib import Path

logger = logging.getLogger(__name__)


class SQLiteCheckpoint:
    """SQLite 状态持久化"""

    def __init__(self, db_path: str = "workflow_checkpoints.db"):
        self.db_path = db_path
        self._init_db()

    def _init_db(self):
        """初始化数据库"""
        db_dir = Path(self.db_path).parent
        db_dir.mkdir(parents=True, exist_ok=True)

        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute("""
            CREATE TABLE IF NOT EXISTS checkpoints (
                thread_id TEXT PRIMARY KEY,
                workflow_id TEXT NOT NULL,
                tenant_id TEXT NOT NULL,
                state TEXT NOT NULL,
                metadata TEXT,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        """)

        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_checkpoints_workflow_id ON checkpoints(workflow_id)
        """)

        cursor.execute("""
            CREATE INDEX IF NOT EXISTS idx_checkpoints_tenant_id ON checkpoints(tenant_id)
        """)

        conn.commit()
        conn.close()

    def save(self, thread_id: str, workflow_id: str, tenant_id: str, state: Dict[str, Any], metadata: Optional[Dict[str, Any]] = None):
        """保存状态"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute("""
            INSERT OR REPLACE INTO checkpoints
            (thread_id, workflow_id, tenant_id, state, metadata, updated_at)
            VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
        """, (thread_id, workflow_id, tenant_id, json.dumps(state), json.dumps(metadata or {})))

        conn.commit()
        conn.close()

    def load(self, thread_id: str) -> Optional[Dict[str, Any]]:
        """加载状态"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute("""
            SELECT state FROM checkpoints WHERE thread_id = ?
        """, (thread_id,))

        row = cursor.fetchone()
        conn.close()

        if row:
            return json.loads(row[0])
        return None

    def delete(self, thread_id: str):
        """删除状态"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute("""
            DELETE FROM checkpoints WHERE thread_id = ?
        """, (thread_id,))

        conn.commit()
        conn.close()

    def list_by_workflow(self, workflow_id: str) -> list:
        """列出指定工作流的所有状态"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute("""
            SELECT thread_id, tenant_id, state, created_at, updated_at
            FROM checkpoints WHERE workflow_id = ?
        """, (workflow_id,))

        rows = cursor.fetchall()
        conn.close()

        results = []
        for row in rows:
            results.append({
                "thread_id": row[0],
                "tenant_id": row[1],
                "state": json.loads(row[2]),
                "created_at": row[3],
                "updated_at": row[4]
            })
        return results

    def list_by_tenant(self, tenant_id: str) -> list:
        """列出指定租户的所有状态"""
        conn = sqlite3.connect(self.db_path)
        cursor = conn.cursor()

        cursor.execute("""
            SELECT thread_id, workflow_id, state, created_at, updated_at
            FROM checkpoints WHERE tenant_id = ?
        """, (tenant_id,))

        rows = cursor.fetchall()
        conn.close()

        results = []
        for row in rows:
            results.append({
                "thread_id": row[0],
                "workflow_id": row[1],
                "state": json.loads(row[2]),
                "created_at": row[3],
                "updated_at": row[4]
            })
        return results