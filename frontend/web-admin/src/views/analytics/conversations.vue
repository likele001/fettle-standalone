<template>
  <div class="analytics-conversations">
    <div class="page-header">
      <h2>对话分析</h2>
      <div class="header-actions">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          @change="loadData"
        />
        <el-button type="primary" @click="loadData">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.totalConversations }}</div>
          <div class="stat-label">总对话数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.avgMessages }}</div>
          <div class="stat-label">平均消息数</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.avgDuration }}分钟</div>
          <div class="stat-label">平均对话时长</div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.completionRate }}%</div>
          <div class="stat-label">完成率</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card class="filter-card" shadow="never">
      <el-row :gutter="16" align="middle">
        <el-col :span="6">
          <el-input v-model="filters.keyword" placeholder="搜索客户名称" clearable @clear="loadData">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.channel" placeholder="渠道" clearable @change="loadData">
            <el-option label="网页" value="web" />
            <el-option label="微信" value="wechat" />
            <el-option label="抖音" value="douyin" />
            <el-option label="企微" value="wecom" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.status" placeholder="状态" clearable @change="loadData">
            <el-option label="活跃" value="active" />
            <el-option label="已关闭" value="closed" />
            <el-option label="等待中" value="waiting" />
          </el-select>
        </el-col>
        <el-col :span="6" style="text-align: right">
          <el-button @click="resetFilters">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-table :data="conversations" stripe v-loading="loading">
      <el-table-column prop="title" label="会话主题" min-width="180" show-overflow-tooltip />
      <el-table-column prop="customer_name" label="客户" width="120" />
      <el-table-column prop="channel" label="渠道" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ channelText(row.channel) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message_count" label="消息数" width="80" />
      <el-table-column prop="duration" label="时长" width="100">
        <template #default="{ row }">{{ formatDuration(row.duration) }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="160">
        <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">{{ statusText(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" @click="viewDetail(row)">详情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-if="total > pageSize"
      class="pagination"
      layout="prev, pager, next, total"
      :total="total"
      :page-size="pageSize"
      v-model:current-page="currentPage"
      @current-change="loadData"
    />

    <el-dialog v-model="detailVisible" title="对话详情" width="700px">
      <div class="conversation-detail" v-if="currentConversation">
        <div class="detail-header">
          <h3>{{ currentConversation.title }}</h3>
          <el-tag :type="statusType(currentConversation.status)">
            {{ statusText(currentConversation.status) }}
          </el-tag>
        </div>
        <div class="detail-info">
          <span>客户: {{ currentConversation.customer_name || '未知' }}</span>
          <span>渠道: {{ channelText(currentConversation.channel) }}</span>
          <span>消息数: {{ currentConversation.message_count }}</span>
          <span>时长: {{ formatDuration(currentConversation.duration) }}</span>
        </div>
        <div class="chat-messages">
          <div v-for="msg in messages" :key="msg.id" :class="['message', msg.sender_type]">
            <div class="sender">{{ msg.sender_name }}</div>
            <div class="content">{{ msg.content }}</div>
            <div class="time">{{ formatTime(msg.created_at) }}</div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Refresh, Search } from '@element-plus/icons-vue'
import { listConversations, getMessages } from '@/api/conversation'
import type { Conversation as ApiConversation, Message as ApiMessage } from '@/api/conversation'

interface Conversation {
  id: string
  title: string
  customer_name: string
  channel: string
  message_count: number
  status: string
  duration?: number
  created_at: string
}

interface Message {
  id: string
  sender_name: string
  sender_type: string
  content: string
  created_at: string
}

const dateRange = ref<[Date, Date]>()
const filters = ref({
  keyword: '',
  channel: '',
  status: ''
})

const stats = ref({
  totalConversations: 0,
  avgMessages: 0,
  avgDuration: 0,
  completionRate: 0
})

const conversations = ref<Conversation[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const detailVisible = ref(false)
const currentConversation = ref<Conversation | null>(null)
const messages = ref<Message[]>([])

const channelText = (channel: string) => {
  const map: Record<string, string> = { web: '网页', wechat: '微信', douyin: '抖音', wecom: '企微' }
  return map[channel] || channel
}

const statusType = (status: string) => {
  const map: Record<string, string> = { active: 'success', waiting: 'warning', closed: 'info' }
  return map[status] || 'info'
}

const statusText = (status: string) => {
  const map: Record<string, string> = { active: '活跃', waiting: '等待', closed: '已关闭' }
  return map[status] || status
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const formatDuration = (seconds?: number) => {
  if (!seconds || seconds < 0) return '-'
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  if (mins > 0) {
    return `${mins}分${secs}秒`
  }
  return `${secs}秒`
}

const resetFilters = () => {
  filters.value = { keyword: '', channel: '', status: '' }
  loadData()
}

const loadData = async () => {
  loading.value = true
  try {
    const params: Record<string, any> = {
      page: currentPage.value,
      page_size: pageSize.value
    }
    if (filters.value.keyword) params.keyword = filters.value.keyword
    if (filters.value.channel) params.channel = filters.value.channel
    if (filters.value.status) params.status = filters.value.status

    const data = await listConversations(params)
    conversations.value = (data.items || []).map((c: ApiConversation) => ({
      id: c.id,
      title: c.customer_name || '未知用户',
      customer_name: c.customer_name || '未知',
      channel: c.channel || 'web',
      message_count: c.message_count || 0,
      status: c.status,
      duration: Math.floor((Date.now() - new Date(c.created_at).getTime()) / 1000),
      created_at: c.created_at
    }))
    total.value = data.total || 0

    const totalMsgs = conversations.value.reduce((sum, c) => sum + c.message_count, 0)
    const totalDur = conversations.value.reduce((sum, c) => sum + (c.duration || 0), 0)
    const count = conversations.value.length

    stats.value.totalConversations = count
    stats.value.avgMessages = count > 0 ? Math.round(totalMsgs / count) : 0
    stats.value.avgDuration = count > 0 ? Math.round(totalDur / 60 / count) : 0
    stats.value.completionRate = count > 0 ? Math.round(conversations.value.filter(c => c.status === 'closed').length / count * 100) : 0
  } catch {
    conversations.value = []
    total.value = 0
    stats.value = { totalConversations: 0, avgMessages: 0, avgDuration: 0, completionRate: 0 }
  } finally {
    loading.value = false
  }
}

const viewDetail = async (conv: Conversation) => {
  currentConversation.value = conv
  detailVisible.value = true
  try {
    const data = await getMessages(conv.id)
    messages.value = (data.items || []).map((m: ApiMessage) => ({
      id: m.id,
      sender_name: m.role === 'user' ? '用户' : '智能体',
      sender_type: m.role,
      content: m.content,
      created_at: m.created_at
    }))
  } catch {
    messages.value = []
  }
}

onMounted(() => {
  loadData()
})
</script>

<style scoped lang="scss">
.analytics-conversations {
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

  .header-actions {
    display: flex;
    gap: 12px;
    align-items: center;
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
  }
}

.filter-card {
  margin-bottom: 20px;

  :deep(.el-card__body) {
    padding: 16px 20px;
  }
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

.conversation-detail {
  .detail-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;

    h3 {
      margin: 0;
      font-size: 18px;
    }
  }

  .detail-info {
    display: flex;
    gap: 24px;
    color: #666;
    margin-bottom: 16px;
    padding-bottom: 16px;
    border-bottom: 1px solid #eee;
    flex-wrap: wrap;
  }

  .chat-messages {
    max-height: 400px;
    overflow-y: auto;
    padding: 16px;
    background: #f5f5f5;
    border-radius: 8px;

    .message {
      margin-bottom: 16px;
      padding: 12px;
      background: white;
      border-radius: 8px;

      &.user {
        border-left: 3px solid #409EFF;
      }

      &.agent {
        border-left: 3px solid #67C23A;
      }

      .sender {
        font-weight: bold;
        margin-bottom: 4px;
      }

      .content {
        color: #333;
        margin-bottom: 4px;
      }

      .time {
        font-size: 12px;
        color: #999;
      }
    }
  }
}
</style>
