<template>
  <div class="conversation-monitor">
    <div class="page-header">
      <h2>对话监控</h2>
      <div class="header-actions">
        <el-button type="primary" @click="loadData">
          <el-icon><Refresh /></el-icon>
          刷新数据
        </el-button>
      </div>
    </div>

    <el-row :gutter="16" class="stats-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.activeCount }}</div>
          <div class="stat-label">活跃会话</div>
          <div class="stat-icon active">
            <el-icon :size="24"><CircleCheck /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.waitingCount }}</div>
          <div class="stat-label">等待队列</div>
          <div class="stat-icon waiting">
            <el-icon :size="24"><Clock /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.todayConversations }}</div>
          <div class="stat-label">今日会话</div>
          <div class="stat-icon today">
            <el-icon :size="24"><Calendar /></el-icon>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card">
          <div class="stat-value">{{ stats.todayMessages }}</div>
          <div class="stat-label">今日消息</div>
          <div class="stat-icon messages">
            <el-icon :size="24"><Message /></el-icon>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16">
      <el-col :span="12">
        <el-card>
          <template #header>
            <span>活跃会话</span>
          </template>
          <el-table :data="activeConversations" stripe v-loading="loading">
            <el-table-column prop="title" label="会话" min-width="150" show-overflow-tooltip />
            <el-table-column prop="customer_name" label="客户" width="120" />
            <el-table-column prop="channel" label="渠道" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ channelText(row.channel) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="duration" label="时长" width="100">
              <template #default="{ row }">
                {{ formatDuration(row.duration) }}
              </template>
            </el-table-column>
            <el-table-column prop="message_count" label="消息数" width="80" />
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button size="small" @click="viewConversation(row)">查看</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="activeConversations.length === 0 && !loading" class="empty-state">
            <el-empty description="暂无活跃会话" />
          </div>
        </el-card>
      </el-col>

      <el-col :span="12">
        <el-card>
          <template #header>
            <span>等待队列</span>
          </template>
          <el-table :data="waitingConversations" stripe v-loading="loading">
            <el-table-column prop="title" label="会话" min-width="150" show-overflow-tooltip />
            <el-table-column prop="customer_name" label="客户" width="120" />
            <el-table-column prop="channel" label="渠道" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ channelText(row.channel) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="waiting_time" label="等待时间" width="120">
              <template #default="{ row }">
                <el-tag type="warning" size="small">{{ formatDuration(row.waiting_time) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="100">
              <template #default="{ row }">
                <el-button size="small" type="primary" @click="takeOver(row)">接管</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="waitingConversations.length === 0 && !loading" class="empty-state">
            <el-empty description="暂无等待会话" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="detailVisible" title="会话详情" width="700px">
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
import { ref, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, CircleCheck, Clock, Calendar, Message } from '@element-plus/icons-vue'
import { listConversations, getMessages, takeOverConversation } from '@/api/conversation'
import type { Conversation as ApiConversation, Message as ApiMessage } from '@/api/conversation'

interface Conversation {
  id: string
  title: string
  customer_name: string
  channel: string
  message_count: number
  status: string
  duration?: number
  waiting_time?: number
}

interface Message {
  id: string
  sender_name: string
  sender_type: string
  content: string
  created_at: string
}

const stats = ref({
  activeCount: 0,
  waitingCount: 0,
  todayConversations: 0,
  todayMessages: 0
})

const activeConversations = ref<Conversation[]>([])
const waitingConversations = ref<Conversation[]>([])
const loading = ref(false)

const detailVisible = ref(false)
const currentConversation = ref<Conversation | null>(null)
const messages = ref<Message[]>([])

let refreshInterval: number | null = null

const channelText = (channel: string) => {
  const map: Record<string, string> = {
    web: '网页',
    wechat: '微信',
    douyin: '抖音',
    wecom: '企微'
  }
  return map[channel] || channel
}

const statusType = (status: string) => {
  const map: Record<string, string> = {
    active: 'success',
    waiting: 'warning',
    closed: 'info'
  }
  return map[status] || 'info'
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    active: '活跃',
    waiting: '等待',
    closed: '已关闭'
  }
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

const loadData = async () => {
  loading.value = true
  try {
    const activeData = await listConversations({
      page: 1,
      page_size: 50,
      status: 'active'
    })
    activeConversations.value = (activeData.items || []).map((c: ApiConversation) => ({
      id: c.id,
      title: c.customer_name || '未知用户',
      customer_name: c.customer_name || '未知',
      channel: c.channel || 'web',
      message_count: c.message_count || 0,
      status: c.status,
      duration: Math.floor((Date.now() - new Date(c.created_at).getTime()) / 1000)
    }))

    const waitingData = await listConversations({
      page: 1,
      page_size: 50,
      status: 'waiting'
    })
    waitingConversations.value = (waitingData.items || []).map((c: ApiConversation) => ({
      id: c.id,
      title: c.customer_name || '未知用户',
      customer_name: c.customer_name || '未知',
      channel: c.channel || 'web',
      message_count: c.message_count || 0,
      status: c.status,
      waiting_time: Math.floor((Date.now() - new Date(c.created_at).getTime()) / 1000)
    }))

    stats.value.activeCount = activeConversations.value.length
    stats.value.waitingCount = waitingConversations.value.length
    stats.value.todayConversations = activeConversations.value.length + waitingConversations.value.length

    const totalMessages = activeConversations.value.reduce((sum, c) => sum + c.message_count, 0)
    stats.value.todayMessages = totalMessages
  } catch {
    activeConversations.value = []
    waitingConversations.value = []
    stats.value = { activeCount: 0, waitingCount: 0, todayConversations: 0, todayMessages: 0 }
  } finally {
    loading.value = false
  }
}

const viewConversation = async (conv: Conversation) => {
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

const takeOver = async (conv: Conversation) => {
  try {
    await takeOverConversation(conv.id)
    ElMessage.success('已接管对话')
    loadData()
  } catch (e: any) {
    ElMessage.error(e.message || '接管失败')
  }
}

onMounted(() => {
  loadData()
  refreshInterval = window.setInterval(loadData, 15000)
})

onUnmounted(() => {
  if (refreshInterval) {
    clearInterval(refreshInterval)
  }
})
</script>

<style scoped lang="scss">
.conversation-monitor {
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
    position: relative;

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

    .stat-icon {
      position: absolute;
      top: 16px;
      right: 16px;

      &.active { color: #67C23A; }
      &.waiting { color: #E6A23C; }
      &.today { color: #409EFF; }
      &.messages { color: #909399; }
    }
  }
}

.empty-state {
  padding: 40px 0;
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
