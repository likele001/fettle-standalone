<template>
  <div class="page-container">
    <div class="page-header">
      <h2>平台概览</h2>
      <p>Fettle 管理后台控制台</p>
    </div>

    <!-- 核心指标 -->
    <el-row :gutter="20">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ stats.totalTenants }}</div>
          <div class="stat-label">总租户数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ stats.activeTenants }}</div>
          <div class="stat-label">活跃租户</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ stats.pendingAudit }}</div>
          <div class="stat-label">待审核</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ stats.bannedTenants }}</div>
          <div class="stat-label">已封禁</div>
        </div>
      </el-col>
    </el-row>

    <!-- 平台数据 -->
    <el-row :gutter="20" style="margin-top: 24px">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ overview.total_agents }}</div>
          <div class="stat-label">总智能体</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ overview.total_conversations }}</div>
          <div class="stat-label">总会话数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ overview.total_messages }}</div>
          <div class="stat-label">总消息数</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ overview.active_channels }}</div>
          <div class="stat-label">活跃渠道</div>
        </div>
      </el-col>
    </el-row>

    <!-- 服务状态 -->
    <el-row :gutter="20" style="margin-top: 24px">
      <el-col :span="24">
        <el-card>
          <template #header>
            <span style="font-weight: 600;">服务状态</span>
          </template>
          <el-table :data="services" stripe size="small">
            <el-table-column prop="name" label="服务名称" width="160" />
            <el-table-column prop="status" label="状态" width="120">
              <template #default="{ row }">
                <el-tag :type="statusTag(row.status)" size="small">
                  {{ statusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="cpu_usage" label="CPU" width="100">
              <template #default="{ row }">{{ row.cpu_usage }}%</template>
            </el-table-column>
            <el-table-column prop="memory_usage" label="内存" width="100">
              <template #default="{ row }">{{ row.memory_usage }}%</template>
            </el-table-column>
            <el-table-column prop="response_time" label="响应(ms)" width="100">
              <template #default="{ row }">{{ row.response_time }}</template>
            </el-table-column>
            <el-table-column prop="uptime" label="运行时间" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getTenants } from '@/api/tenant'
import { getSystemStats, getSystemHealth, type ServiceHealth } from '@/api/monitor'
import { getPlatformOverview } from '@/api/admin-analytics'

const stats = ref({
  totalTenants: 0,
  activeTenants: 0,
  pendingAudit: 0,
  bannedTenants: 0
})

const services = ref<ServiceHealth[]>([])

const overview = ref({
  total_tenants: 0,
  total_agents: 0,
  total_conversations: 0,
  total_messages: 0,
  active_channels: 0
})

const statusTag = (status: string) => {
  const map: Record<string, string> = { healthy: 'success', degraded: 'warning', down: 'danger' }
  return map[status] || 'info'
}

const statusText = (status: string) => {
  const map: Record<string, string> = { healthy: '正常', degraded: '降级', down: '异常' }
  return map[status] || status
}

onMounted(async () => {
  try {
    const res = await getTenants({ page_size: 1 })
    stats.value.totalTenants = res.total
  } catch { /* ignore */ }
  try {
    const active = await getTenants({ status: 'active', page_size: 1 })
    stats.value.activeTenants = active.total
  } catch { /* ignore */ }
  try {
    const pending = await getTenants({ audit_status: 'pending', page_size: 1 })
    stats.value.pendingAudit = pending.total
  } catch { /* ignore */ }
  try {
    const banned = await getTenants({ status: 'banned', page_size: 1 })
    stats.value.bannedTenants = banned.total
  } catch { /* ignore */ }

  try {
    overview.value = await getPlatformOverview()
  } catch { /* ignore */ }

  try {
    const health = await getSystemHealth()
    services.value = health?.services || []
  } catch {
    services.value = []
  }
})
</script>

<style scoped lang="scss">
.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  text-align: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-2px);
  }

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
