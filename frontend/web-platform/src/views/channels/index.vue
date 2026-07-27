<template>
  <div class="page-container">
    <div class="page-header">
      <h2>渠道管理</h2>
    </div>

    <!-- Stats Cards -->
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <div class="stat-card">
          <div class="stat-value">{{ channelStats.total || 0 }}</div>
          <div class="stat-label">总渠道数</div>
        </div>
      </el-col>
      <el-col :span="6" v-for="(item, idx) in topChannels" :key="idx">
        <div class="stat-card">
          <div class="stat-value">{{ item.count }}</div>
          <div class="stat-label">{{ channelLabel(item.channel) }}</div>
        </div>
      </el-col>
    </el-row>

    <!-- Channels Table -->
    <el-card>
      <el-table :data="channels" stripe v-loading="loading">
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag size="small">{{ channelLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="tenant_id" label="租户ID" show-overflow-tooltip />
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
          @size-change="loadChannels"
          @current-change="loadChannels"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { getAdminChannels, getAdminChannelStats } from '@/api/admin-channel'

const channels = ref<any[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const channelStats = ref<any>({})

const channelLabels: Record<string, string> = {
  wechat: '微信公众号',
  wecom: '企业微信',
  feishu: '飞书',
  dingtalk: '钉钉',
  douyin: '抖音',
  web: '网页',
}

const channelLabel = (type: string) => channelLabels[type] || type

const topChannels = computed(() => {
  const byType = channelStats.value.by_type || []
  return byType.slice(0, 3)
})

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

async function loadChannels() {
  loading.value = true
  try {
    const res = await getAdminChannels({ page: currentPage.value, page_size: pageSize.value })
    channels.value = res?.items || []
    total.value = res?.total || 0
  } catch {
    channels.value = []
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    channelStats.value = await getAdminChannelStats()
  } catch {
    channelStats.value = {}
  }
}

onMounted(() => {
  loadChannels()
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
