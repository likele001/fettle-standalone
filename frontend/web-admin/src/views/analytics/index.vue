<template>
  <div class="analytics">
    <div class="page-header">
      <h2>数据分析</h2>
      <el-date-picker
        v-model="dateRange"
        type="daterange"
        range-separator="至"
        start-placeholder="开始日期"
        end-placeholder="结束日期"
        @change="loadData"
      />
    </div>

    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.totalConversations }}</div>
          <div class="stat-label">总会话数</div>
          <div class="stat-trend" :class="stats.conversationTrend >= 0 ? 'up' : 'down'">
            {{ stats.conversationTrend >= 0 ? '+' : '' }}{{ stats.conversationTrend }}%
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.totalMessages }}</div>
          <div class="stat-label">总消息数</div>
          <div class="stat-trend" :class="stats.messageTrend >= 0 ? 'up' : 'down'">
            {{ stats.messageTrend >= 0 ? '+' : '' }}{{ stats.messageTrend }}%
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.avgResponseTime }}s</div>
          <div class="stat-label">平均响应时间</div>
          <div class="stat-trend" :class="stats.responseTimeTrend <= 0 ? 'up' : 'down'">
            {{ stats.responseTimeTrend <= 0 ? '' : '+' }}{{ stats.responseTimeTrend }}%
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.satisfactionRate }}%</div>
          <div class="stat-label">满意度</div>
          <div class="stat-trend" :class="stats.satisfactionTrend >= 0 ? 'up' : 'down'">
            {{ stats.satisfactionTrend >= 0 ? '+' : '' }}{{ stats.satisfactionTrend }}%
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="chart-row">
      <el-col :span="16">
        <el-card>
          <template #header>
            <span>会话趋势</span>
          </template>
          <div class="chart-container">
            <v-chart :option="trendChartOption" style="height: 280px" autoresize />
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card>
          <template #header>
            <span>渠道分布</span>
          </template>
          <div class="chart-container">
            <v-chart :option="channelChartOption" style="height: 280px" autoresize />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" class="chart-row">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>热门意图 Top 10</span>
          </template>
          <el-table :data="topIntents" stripe>
            <el-table-column prop="intent" label="意图" />
            <el-table-column prop="count" label="次数" width="100" />
            <el-table-column prop="percentage" label="占比" width="100">
              <template #default="{ row }">{{ row.percentage }}%</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>智能体使用情况</span>
          </template>
          <el-table :data="agentUsage" stripe>
            <el-table-column prop="name" label="智能体" />
            <el-table-column prop="conversations" label="会话数" width="100" />
            <el-table-column prop="messages" label="消息数" width="100" />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { getDashboardStats, getAgentUsage, getConversationTrend, getChannelDistribution, getTopIntents } from '@/api/analytics'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
} from 'echarts/components'

use([
  CanvasRenderer,
  BarChart,
  LineChart,
  PieChart,
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
])

const dateRange = ref<[Date, Date]>()

const stats = ref({
  totalConversations: 0,
  totalMessages: 0,
  avgResponseTime: 0,
  satisfactionRate: 0,
  conversationTrend: 0,
  messageTrend: 0,
  responseTimeTrend: 0,
  satisfactionTrend: 0
})

const trendData = ref<{ date: string; value: number }[]>([])
const channelData = ref<{ name: string; value: number }[]>([])

const topIntents = ref<{ intent: string; count: number; percentage: number }[]>([])

const agentUsage = ref<{ name: string; conversations: number; messages: number }[]>([])

const trendChartOption = computed(() => {
  const labels = trendData.value.map(item => item.date)
  const values = trendData.value.map(item => item.value)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        crossStyle: {
          color: '#999'
        }
      }
    },
    legend: {
      data: ['会话数'],
      top: 0,
      textStyle: {
        color: '#909399',
        fontSize: 12
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      top: '15%',
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
        name: '会话数',
        type: 'line',
        smooth: true,
        data: values,
        lineStyle: {
          width: 3,
          color: '#409EFF'
        },
        itemStyle: {
          color: '#409EFF',
          borderWidth: 2,
          borderColor: '#fff'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
              { offset: 1, color: 'rgba(64, 158, 255, 0.05)' }
            ]
          }
        }
      }
    ]
  }
})

const channelChartOption = computed(() => {
  const colors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399']
  
  return {
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left',
      textStyle: {
        color: '#909399',
        fontSize: 12
      }
    },
    series: [
      {
        name: '渠道分布',
        type: 'pie',
        radius: ['45%', '70%'],
        center: ['60%', '55%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 6,
          borderColor: '#fff',
          borderWidth: 2
        },
        label: {
          show: true,
          formatter: '{b}: {d}%',
          color: '#606266',
          fontSize: 12
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: true
        },
        data: channelData.value.map((item, index) => ({
          value: item.value,
          name: item.name,
          itemStyle: {
            color: colors[index % colors.length]
          }
        }))
      }
    ]
  }
})

const loadData = async () => {
  try {
    const d = await getDashboardStats()
    if (d) {
      stats.value.totalConversations = d.total_conversations || 0
      stats.value.totalMessages = d.total_messages || 0
    }
  } catch { /* ignore */ }

  try {
    const usage = await getAgentUsage()
    if (usage && usage.length > 0) {
      agentUsage.value = usage.map((u: { agent_name: string; conversation_count: number; message_count: number }) => ({
        name: u.agent_name || '未知',
        conversations: u.conversation_count || 0,
        messages: u.message_count || 0
      }))
    }
  } catch { /* ignore */ }

  await loadTrendData()
  loadChannelData()
  await loadTopIntents()
}

async function loadTrendData() {
  try {
    const data = await getConversationTrend({ days: 7 })
    if (data) {
      trendData.value = data.map((item: { date: string; value: number }) => ({
        date: item.date,
        value: item.value
      }))
    }
  } catch {
    trendData.value = []
  }
}

async function loadChannelData() {
  try {
    const data = await getChannelDistribution()
    if (data && data.length > 0) {
      channelData.value = data.map((item: { channel: string; count: number }) => ({
        name: item.channel,
        value: item.count
      }))
    } else {
      channelData.value = []
    }
  } catch {
    channelData.value = []
  }
}

async function loadTopIntents() {
  try {
    const data = await getTopIntents({ limit: 10 })
    if (data && data.length > 0) {
      topIntents.value = data
    }
  } catch {
    topIntents.value = []
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.analytics {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.stats-row {
  margin-bottom: 24px;

  .stat-card {
    text-align: center;

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
      margin-top: 8px;
      font-size: 12px;

      &.up {
        color: #67C23A;
      }

      &.down {
        color: #F56C6C;
      }
    }
  }
}

.chart-row {
  margin-bottom: 24px;

  .chart-container {
    height: 280px;
  }
}
</style>
