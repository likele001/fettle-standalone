<template>
  <div class="conversation-list">
    <div class="page-header">
      <h2>会话管理</h2>
      <div class="header-actions">
        <el-select v-model="filter.agentId" placeholder="选择智能体" clearable @change="loadConversations">
          <el-option v-for="agent in agents" :key="agent.id" :label="agent.name" :value="agent.id" />
        </el-select>
        <el-select v-model="filter.status" placeholder="状态" clearable @change="loadConversations">
          <el-option label="活跃" value="active" />
          <el-option label="等待" value="waiting" />
          <el-option label="已关闭" value="closed" />
        </el-select>
        <el-button type="success" @click="showExportDialog('conversations')">
          <el-icon><Download /></el-icon> 导出对话
        </el-button>
        <el-button type="warning" @click="showExportDialog('messages')">
          <el-icon><Download /></el-icon> 导出消息
        </el-button>
      </div>
    </div>

    <el-table :data="conversations" stripe v-loading="loading">
      <el-table-column prop="title" label="标题" min-width="200" show-overflow-tooltip />
      <el-table-column prop="customer_name" label="客户" width="120" />
      <el-table-column prop="agent_name" label="智能体" width="120" />
      <el-table-column prop="channel" label="渠道" width="100">
        <template #default="{ row }">
          <el-tag size="small">{{ channelText(row.channel) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="message_count" label="消息数" width="80" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusType(row.status)" size="small">
            {{ statusText(row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="last_message_at" label="最后消息" width="160">
        <template #default="{ row }">
          {{ formatTime(row.last_message_at) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="viewConversation(row)">查看</el-button>
          <el-button size="small" v-if="row.status === 'waiting'" @click="takeOver(row)">接管</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      v-if="total > pageSize"
      class="pagination"
      layout="prev, pager, next"
      :total="total"
      :page-size="pageSize"
      v-model:current-page="currentPage"
      @current-change="loadConversations"
    />

    <!-- 会话详情对话框 -->
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
  
    <!-- 导出日期选择对话框 -->
    <el-dialog v-model="exportDialogVisible" :title="exportType === 'conversations' ? '导出对话' : '导出消息'" width="420px">
      <el-form label-width="80px">
        <el-form-item label="开始日期">
          <el-date-picker
            v-model="exportDateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="exportDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleExport" :loading="exporting">导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { Agent } from '@/api/agent'
import { getAgents } from '@/api/agent'
import { listConversations, getMessages, takeOverConversation } from '@/api/conversation'
import type { Conversation as ApiConversation, Message as ApiMessage } from '@/api/conversation'
import { exportConversations, exportMessages, downloadBlob } from '@/api/export'

interface Conversation {
  id: string
  title: string
  customer_name: string
  agent_name: string
  channel: string
  message_count: number
  status: string
  last_message_at: string
}

interface Message {
  id: string
  sender_name: string
  sender_type: string
  content: string
  created_at: string
}

const conversations = ref<Conversation[]>([])
const agents = ref<Agent[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

const filter = ref({
  agentId: '',
  status: ''
})

const detailVisible = ref(false)
const currentConversation = ref<Conversation | null>(null)
const messages = ref<Message[]>([])

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

const loadConversations = async () => {
  loading.value = true
  try {
    const data = await listConversations({
      page: currentPage.value,
      page_size: pageSize.value,
      status: filter.value.status || undefined,
      agent_id: filter.value.agentId || undefined
    })
    conversations.value = (data.items || []).map((c: ApiConversation) => ({
      id: c.id,
      title: c.customer_name || '未知用户',
      customer_name: c.customer_name || '未知',
      agent_name: '',
      channel: c.channel || 'web',
      message_count: c.message_count || 0,
      status: c.status,
      last_message_at: c.updated_at
    }))
    total.value = data.total || 0
  } catch {
    conversations.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const loadAgents = async () => {
  try {
    const res = await getAgents(1, 100)
    agents.value = res.items
  } catch {
    // ignore
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
  } catch (e: any) {
    ElMessage.error(e.message || '接管失败')
  }
}

// 导出相关
const exportDialogVisible = ref(false)
const exportType = ref<'conversations' | 'messages'>('conversations')
const exportDateRange = ref<string[]>([])
const exporting = ref(false)

const showExportDialog = (type: 'conversations' | 'messages') => {
  exportType.value = type
  exportDateRange.value = []
  exportDialogVisible.value = true
}

const handleExport = async () => {
  if (!exportDateRange.value || exportDateRange.value.length < 2) {
    ElMessage.warning('请选择日期范围')
    return
  }
  exporting.value = true
  try {
    const params = {
      start_date: exportDateRange.value[0],
      end_date: exportDateRange.value[1]
    }
    const blob = exportType.value === 'conversations'
      ? await exportConversations(params) as any
      : await exportMessages(params) as any
    const filename = exportType.value === 'conversations' ? 'conversations_export.csv' : 'messages_export.csv'
    downloadBlob(blob, filename)
    ElMessage.success('导出成功')
    exportDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '导出失败')
  } finally {
    exporting.value = false
  }
}

onMounted(() => {
  loadConversations()
  loadAgents()
})
</script>

<style scoped lang="scss">
.conversation-list {
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
