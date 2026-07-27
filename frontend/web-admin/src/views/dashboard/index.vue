<template>
  <div class="dashboard">
    <div class="welcome">
      <h1>欢迎回来，{{ userStore.userInfo?.name || '用户' }}</h1>
      <p>这是 AI 智能体平台管理后台，您可以从左侧菜单开始管理您的智能体。</p>
    </div>

    <!-- 核心指标 -->
    <el-row :gutter="[12, 12]" class="stats-row">
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.agents }}</div>
          <div class="stat-label">智能体</div>
          <div class="stat-trend" :class="trendClass(stats.agentTrend)">
            <el-icon v-if="stats.agentTrend > 0"><Top /></el-icon>
            <el-icon v-else-if="stats.agentTrend < 0"><Bottom /></el-icon>
            {{ Math.abs(stats.agentTrend) }}%
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.conversations }}</div>
          <div class="stat-label">今日会话</div>
          <div class="stat-trend" :class="trendClass(stats.convTrend)">
            <el-icon v-if="stats.convTrend > 0"><Top /></el-icon>
            <el-icon v-else-if="stats.convTrend < 0"><Bottom /></el-icon>
            {{ Math.abs(stats.convTrend) }}%
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.messages }}</div>
          <div class="stat-label">今日消息</div>
          <div class="stat-trend" :class="trendClass(stats.msgTrend)">
            <el-icon v-if="stats.msgTrend > 0"><Top /></el-icon>
            <el-icon v-else-if="stats.msgTrend < 0"><Bottom /></el-icon>
            {{ Math.abs(stats.msgTrend) }}%
          </div>
        </el-card>
      </el-col>
      <el-col :xs="12" :sm="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.knowledgeBases }}</div>
          <div class="stat-label">知识库</div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 收入与配额 -->
    <el-row :gutter="[12, 12]" style="margin-bottom: 16px">
      <el-col :xs="24" :md="16">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px;">
              <span style="font-weight: 600">近 7 天消息量趋势</span>
              <el-radio-group v-model="trendPeriod" size="small" @change="loadTrendData">
                <el-radio-button label="7d">7天</el-radio-button>
                <el-radio-button label="30d">30天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="chart-container">
            <v-chart :option="trendChartOption" style="height: 220px" autoresize />
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center">
              <span style="font-weight: 600">资源配额</span>
              <el-tag
                :type="planBadgeType"
                size="small"
                effect="dark"
              >{{ planDisplayName }}</el-tag>
            </div>
          </template>
          <div class="quota-list">
            <div class="quota-item" v-for="q in quotaList" :key="q.label">
              <div class="quota-header">
                <span>{{ q.label }}</span>
                <span class="quota-num">已使用 {{ q.used }}/{{ q.limit }} {{ q.unit }}</span>
              </div>
              <el-progress
                :percentage="q.percent"
                :status="q.percent >= 90 ? 'exception' : q.percent >= 70 ? 'warning' : ''"
                :stroke-width="8"
              />
            </div>
          </div>
          <div class="plan-info">
            <span class="plan-expire" v-if="planExpire">到期：{{ planExpire }}</span>
            <span class="plan-expire" v-else-if="quotaInfo?.is_free_plan">免费版</span>
            <el-button
              v-if="quotaInfo?.is_free_plan"
              type="warning"
              size="small"
              class="upgrade-btn"
              @click="$router.push('/billing')"
            >升级套餐</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 最近会话 & 快捷操作 -->
    <el-row :gutter="[12, 12]">
      <el-col :xs="24" :md="14">
        <el-card shadow="hover">
          <template #header>
            <div style="display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px;">
              <span style="font-weight: 600">最近会话</span>
              <el-button text type="primary" size="small" @click="$router.push('/conversations')">
                查看全部
              </el-button>
            </div>
          </template>
          <div class="table-wrapper">
            <el-table :data="recentConversations" size="small" stripe>
              <el-table-column prop="user_name" label="用户" width="120" />
              <el-table-column prop="agent_name" label="智能体" width="120" />
              <el-table-column prop="last_message" label="最后消息" show-overflow-tooltip />
              <el-table-column prop="created_at" label="时间" width="140">
                <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
              </el-table-column>
              <el-table-column label="状态" width="70">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                    {{ row.status === 'active' ? '进行中' : '已结束' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
          <el-empty v-if="recentConversations.length === 0" description="暂无会话" :image-size="60" />
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card shadow="hover">
          <template #header>
            <span style="font-weight: 600">快捷操作</span>
          </template>
          <div class="quick-actions-grid">
            <div class="action-item" @click="$router.push('/agents')">
              <el-icon :size="28" color="#409EFF"><Plus /></el-icon>
              <span>创建智能体</span>
            </div>
            <div class="action-item" @click="$router.push('/knowledge')">
              <el-icon :size="28" color="#67C23A"><Upload /></el-icon>
              <span>上传文档</span>
            </div>
            <div class="action-item" @click="$router.push('/conversations')">
              <el-icon :size="28" color="#E6A23C"><ChatDotRound /></el-icon>
              <span>查看会话</span>
            </div>
            <div class="action-item" @click="$router.push('/channels')">
              <el-icon :size="28" color="#F56C6C"><Connection /></el-icon>
              <span>渠道接入</span>
            </div>
            <div class="action-item" @click="$router.push('/billing')">
              <el-icon :size="28" color="#909399"><CreditCard /></el-icon>
              <span>套餐计费</span>
            </div>
            <div class="action-item" @click="$router.push('/settings')">
              <el-icon :size="28" color="#909399"><Setting /></el-icon>
              <span>系统设置</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  Plus, Upload, ChatDotRound, Connection,
  Top, Bottom, CreditCard, Setting
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { getDashboardStats, getConversationTrend, getQuotaInfo, getRecentConversations, type QuotaInfo, type RecentConversation } from '@/api/analytics'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
} from 'echarts/components'

use([
  CanvasRenderer,
  BarChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
])

const userStore = useUserStore()

const stats = ref({
  agents: 0,
  conversations: 0,
  messages: 0,
  knowledgeBases: 0,
  agentTrend: 0,
  convTrend: 0,
  msgTrend: 0
})

const trendPeriod = ref('7d')
const trendData = ref<{ date: string; value: number }[]>([])

const quotaInfo = ref<QuotaInfo | null>(null)

const quotaList = ref([
  { label: '智能体', used: 0, limit: 5, percent: 0, unit: '个' },
  { label: '消息量', used: 0, limit: 10000, percent: 0, unit: '条' },
  { label: '知识库', used: 0, limit: 3, percent: 0, unit: '个' }
])

const planDisplayName = computed(() => {
  if (!quotaInfo.value) return '免费版'
  const name = quotaInfo.value.plan_name
  if (name) return name
  if (quotaInfo.value.is_free_plan) return '免费版'
  return '专业版'
})

const planBadgeType = computed(() => {
  if (!quotaInfo.value) return 'info'
  if (quotaInfo.value.is_free_plan) return 'info'
  if (quotaInfo.value.plan_type === 'enterprise') return 'danger'
  return 'primary'
})

const planExpire = computed(() => {
  if (!quotaInfo.value?.plan_expire_at) return ''
  const d = new Date(quotaInfo.value.plan_expire_at)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
})

const recentConversations = ref<RecentConversation[]>([])

const trendChartOption = computed(() => {
  const labels = trendData.value.map(item => item.date)
  const values = trendData.value.map(item => item.value)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: labels,
      axisLine: {
        lineStyle: {
          color: '#ebeef5'
        }
      },
      axisLabel: {
        color: '#909399',
        fontSize: 12
      }
    },
    yAxis: {
      type: 'value',
      axisLine: {
        show: false
      },
      axisLabel: {
        color: '#909399',
        fontSize: 12
      },
      splitLine: {
        lineStyle: {
          color: '#ebeef5',
          type: 'dashed'
        }
      }
    },
    series: [
      {
        name: '消息量',
        type: 'bar',
        data: values,
        barWidth: '50%',
        itemStyle: {
          borderRadius: [4, 4, 0, 0],
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: '#409EFF' },
              { offset: 1, color: '#79BBFF' }
            ]
          }
        }
      }
    ]
  }
})

function trendClass(val: number) {
  if (val > 0) return 'up'
  if (val < 0) return 'down'
  return ''
}

function formatTime(t: string) {
  if (!t) return '-'
  const d = new Date(t)
  return `${d.getMonth() + 1}/${d.getDate()} ${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}`
}

async function loadTrendData() {
  try {
    const days = trendPeriod.value === '7d' ? 7 : 30
    const data = await getConversationTrend({ days })
    if (data) {
      trendData.value = data.map((item: { date: string; value: number }) => ({
        date: item.date,
        value: item.value
      }))
    }
  } catch {
    const days = trendPeriod.value === '7d' ? 7 : 30
    const now = new Date()
    trendData.value = Array.from({ length: days }, (_, i) => {
      const d = new Date(now)
      d.setDate(d.getDate() - (days - 1 - i))
      return {
        date: `${d.getMonth() + 1}/${d.getDate()}`,
        value: Math.floor(Math.random() * 200) + 50
      }
    })
  }
}

async function loadQuota() {
  try {
    const data = await getQuotaInfo()
    if (data) {
      quotaInfo.value = data
      quotaList.value[0].used = data.agent_used || 0
      quotaList.value[0].limit = data.agent_limit || 5
      quotaList.value[0].percent = Math.round((quotaList.value[0].used / quotaList.value[0].limit) * 100)

      quotaList.value[1].used = data.message_used || 0
      quotaList.value[1].limit = data.message_limit || 10000
      quotaList.value[1].percent = Math.round((quotaList.value[1].used / quotaList.value[1].limit) * 100)

      quotaList.value[2].used = data.kb_used || 0
      quotaList.value[2].limit = data.kb_limit || 3
      quotaList.value[2].percent = Math.round((quotaList.value[2].used / quotaList.value[2].limit) * 100)
    }
  } catch {
    // Fallback: keep defaults
  }
}

onMounted(async () => {
  try {
    const data = await getDashboardStats()
    if (data) {
      stats.value.agents = data.total_agents || 0
      stats.value.conversations = data.today_conversations || 0
      stats.value.messages = data.today_messages || 0
      stats.value.knowledgeBases = data.total_knowledge_bases || 0
      stats.value.agentTrend = data.agent_trend || 0
      stats.value.convTrend = data.conv_trend || 0
      stats.value.msgTrend = data.msg_trend || 0
    }
  } catch {
    // 静默处理
  }

  loadQuota()
  loadTrendData()
  loadRecentConversations()
})

async function loadRecentConversations() {
  try {
    const data = await getRecentConversations({ limit: 10 })
    if (data) {
      recentConversations.value = data
    }
  } catch {
    recentConversations.value = []
  }
}
</script>

<style scoped lang="scss">
.dashboard {
  padding: 0;
}

.welcome {
  margin-bottom: 24px;

  h1 {
    font-size: 24px;
    color: #333;
    margin: 0 0 8px 0;
  }

  p {
    color: #666;
    margin: 0;
  }
}

.stats-row {
  margin-bottom: 20px;

  .stat-card {
    text-align: center;
    cursor: default;

    .stat-value {
      font-size: 32px;
      font-weight: bold;
      color: #409EFF;
    }

    .stat-label {
      color: #999;
      margin-top: 8px;
      font-size: 14px;
    }

    .stat-trend {
      font-size: 12px;
      margin-top: 4px;
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 2px;

      &.up {
        color: #67C23A;
      }

      &.down {
        color: #F56C6C;
      }
    }
  }
}

.chart-container {
  min-height: 180px;
}

.quota-list {
  .quota-item {
    margin-bottom: 16px;

    &:last-child {
      margin-bottom: 0;
    }

    .quota-header {
      display: flex;
      justify-content: space-between;
      margin-bottom: 6px;
      font-size: 14px;

      .quota-num {
        color: #999;
        font-size: 13px;
      }
    }
  }
}

.plan-info {
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  gap: 8px;

  .plan-expire {
    font-size: 12px;
    color: #999;
  }

  .upgrade-btn {
    margin-left: auto;
  }
}

.quick-actions-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;

  .table-wrapper {
    overflow-x: auto;

    :deep(.el-table) {
      min-width: 500px;
    }
  }

  .action-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 16px 8px;
    border-radius: 8px;
    cursor: pointer;
    transition: background 0.2s;

    &:hover {
      background: #f5f7fa;
    }

    span {
      font-size: 13px;
      color: #666;
    }
  }
}
</style>
