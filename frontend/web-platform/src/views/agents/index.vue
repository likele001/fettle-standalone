<template>
  <div class="page-container">
    <div class="page-header">
      <h2>智能体管理</h2>
    </div>

    <!-- Stats Cards -->
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ agentStats.total || 0 }}</div>
          <div class="stat-label">总智能体数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ agentStats.by_status?.idle || 0 }}</div>
          <div class="stat-label">空闲</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ agentStats.by_status?.processing || 0 }}</div>
          <div class="stat-label">处理中</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ agentStats.by_status?.error || 0 }}</div>
          <div class="stat-label">异常</div>
        </div>
      </el-col>
    </el-row>

    <!-- Agent Table -->
    <el-card>
      <el-table :data="agents" stripe v-loading="loading">
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column prop="agent_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.agent_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'idle' ? 'success' : row.status === 'error' ? 'danger' : 'warning'" size="small">
              {{ row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="tenant_id" label="租户ID" width="280" show-overflow-tooltip />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; display: flex; justify-content: flex-end">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadAgents"
          @current-change="loadAgents"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAdminAgents, getAdminAgentStats } from '@/api/admin-agent'

const agents = ref<any[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const agentStats = ref<any>({})

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

async function loadAgents() {
  loading.value = true
  try {
    const res = await getAdminAgents({ page: currentPage.value, page_size: pageSize.value })
    agents.value = res?.items || []
    total.value = res?.total || 0
  } catch {
    agents.value = []
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    agentStats.value = await getAdminAgentStats()
  } catch {
    agentStats.value = {}
  }
}

onMounted(() => {
  loadAgents()
  loadStats()
})
</script>

<style scoped lang="scss">
.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  text-align: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);

  .stat-value {
    font-size: 32px;
    font-weight: 700;
    color: #1a1a2e;
    margin-bottom: 8px;
  }

  .stat-label {
    font-size: 14px;
    color: #666;
  }
}
</style>
