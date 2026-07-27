<template>
  <div class="page-container">
    <div class="page-header">
      <h2>系统监控</h2>
      <el-button @click="refreshAll" :loading="loading">
        <el-icon><Refresh /></el-icon>
        刷新
      </el-button>
    </div>

    <!-- 系统概览 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-label">服务总数</div>
          <div class="stat-value">{{ systemStats.total_services }}</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-label">健康服务</div>
          <div class="stat-value success">{{ systemStats.healthy_services }}</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-label">24小时请求数</div>
          <div class="stat-value">{{ formatNumber(systemStats.total_requests_24h) }}</div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-label">平均响应时间</div>
          <div class="stat-value">{{ systemStats.avg_response_time }}ms</div>
        </div>
      </el-col>
    </el-row>

    <!-- 服务健康状态 -->
    <el-card class="section-card">
      <template #header>
        <span style="font-weight: 600;">服务健康状态</span>
      </template>
      <el-table :data="services" stripe style="width: 100%">
        <el-table-column prop="name" label="服务名称" width="180" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="healthTagType(row.status)" size="small">
              <el-icon v-if="row.status === 'healthy'" style="margin-right: 4px;"><CircleCheck /></el-icon>
              <el-icon v-else-if="row.status === 'degraded'" style="margin-right: 4px;"><Warning /></el-icon>
              <el-icon v-else style="margin-right: 4px;"><CircleClose /></el-icon>
              {{ healthLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="运行时间" width="140">
          <template #default="{ row }">
            {{ row.uptime }}
          </template>
        </el-table-column>
        <el-table-column label="CPU 使用率" width="140">
          <template #default="{ row }">
            <el-progress
              :percentage="row.cpu_usage"
              :color="progressColor(row.cpu_usage)"
              :stroke-width="8"
              style="width: 100px"
            />
          </template>
        </el-table-column>
        <el-table-column label="内存使用率" width="140">
          <template #default="{ row }">
            <el-progress
              :percentage="row.memory_usage"
              :color="progressColor(row.memory_usage)"
              :stroke-width="8"
              style="width: 100px"
            />
          </template>
        </el-table-column>
        <el-table-column label="响应时间" width="120">
          <template #default="{ row }">
            <span :class="{ 'text-warning': row.response_time > 500, 'text-danger': row.response_time > 1000 }">
              {{ row.response_time }}ms
            </span>
          </template>
        </el-table-column>
        <el-table-column label="最后检查" width="170">
          <template #default="{ row }">
            {{ formatTime(row.last_check) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100" fixed="right">
          <template #default="{ row }">
            <el-popconfirm
              :title="`确定要重启服务 ${row.name} 吗？`"
              @confirm="handleRestart(row.name)"
            >
              <template #reference>
                <el-button type="warning" link size="small">重启</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- API 调用统计 -->
    <el-card class="section-card">
      <template #header>
        <span style="font-weight: 600;">API 调用统计（24小时）</span>
      </template>
      <el-table :data="apiStats" stripe style="width: 100%">
        <el-table-column prop="service" label="服务" width="180" />
        <el-table-column label="调用次数" width="140">
          <template #default="{ row }">
            {{ formatNumber(row.calls_24h) }}
          </template>
        </el-table-column>
        <el-table-column label="成功率" width="140">
          <template #default="{ row }">
            <el-progress
              :percentage="row.success_rate"
              :color="successColor(row.success_rate)"
              :stroke-width="8"
              style="width: 100px"
            />
          </template>
        </el-table-column>
        <el-table-column label="平均延迟" width="120">
          <template #default="{ row }">
            {{ row.avg_latency }}ms
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, CircleCheck, Warning, CircleClose } from '@element-plus/icons-vue'
import { getSystemHealth, getSystemStats, getApiCallStats, restartService, type ServiceHealth, type SystemStats, type ApiCallStats } from '@/api/monitor'

const loading = ref(false)
const services = ref<ServiceHealth[]>([])
const systemStats = ref<SystemStats>({
  total_services: 0,
  healthy_services: 0,
  total_requests_24h: 0,
  avg_response_time: 0,
  error_rate: 0
})
const apiStats = ref<ApiCallStats[]>([])

function healthTagType(status: string) {
  const map: Record<string, string> = { healthy: 'success', degraded: 'warning', down: 'danger' }
  return map[status] || 'info'
}

function healthLabel(status: string) {
  const map: Record<string, string> = { healthy: '健康', degraded: '降级', down: '离线' }
  return map[status] || status
}

function progressColor(value: number) {
  if (value < 60) return '#10b981'
  if (value < 80) return '#f59e0b'
  return '#ef4444'
}

function successColor(value: number) {
  if (value >= 99) return '#10b981'
  if (value >= 95) return '#f59e0b'
  return '#ef4444'
}

function formatNumber(num: number) {
  if (num >= 10000) return (num / 10000).toFixed(1) + '万'
  if (num >= 1000) return (num / 1000).toFixed(1) + 'k'
  return num.toString()
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

async function loadHealth() {
  try {
    const health = await getSystemHealth()
    services.value = health?.services || []
  } catch {
    services.value = []
  }
}

async function loadStats() {
  try {
    systemStats.value = await getSystemStats()
  } catch {
    // ignore
  }
}

async function loadApiStats() {
  try {
    apiStats.value = await getApiCallStats()
  } catch {
    apiStats.value = []
  }
}

async function handleRestart(serviceName: string) {
  try {
    await restartService(serviceName)
    ElMessage.success(`服务 ${serviceName} 重启成功`)
    setTimeout(() => refreshAll(), 2000)
  } catch (e: any) {
    ElMessage.error(e.message || '重启失败')
  }
}

async function refreshAll() {
  loading.value = true
  try {
    await Promise.all([loadHealth(), loadStats(), loadApiStats()])
    ElMessage.success('数据已刷新')
  } finally {
    loading.value = false
  }
}

onMounted(() => refreshAll())
</script>

<style scoped lang="scss">
.stats-row {
  margin-bottom: 20px;

  .stat-card {
    background: #fff;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);

    .stat-label {
      font-size: 14px;
      color: #64748b;
      margin-bottom: 8px;
    }

    .stat-value {
      font-size: 24px;
      font-weight: 700;
      color: #1e293b;

      &.success {
        color: #10b981;
      }
    }
  }
}

.section-card {
  margin-bottom: 20px;
}

.text-warning {
  color: #f59e0b;
  font-weight: 600;
}

.text-danger {
  color: #ef4444;
  font-weight: 600;
}
</style>
