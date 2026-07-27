<template>
  <view class="agent-list-page">
    <view class="search-bar">
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索智能体"
        class="search-input"
        @input="handleSearchInput"
        @confirm="handleSearch"
      />
      <scroll-view scroll-x class="industry-scroll" :show-scrollbar="false">
        <view class="industry-list">
          <view
            v-for="item in industries"
            :key="item"
            class="industry-tag"
            :class="{ active: selectedIndustry === item }"
            @click="handleSelectIndustry(item)"
          >
            {{ item }}
          </view>
        </view>
      </scroll-view>
    </view>

    <view class="agent-list">
      <view
        v-for="agent in agents"
        :key="agent.id"
        class="agent-card"
        @click="goToChat(agent)"
      >
        <image class="agent-avatar" :src="agent.avatar || '/static/default-agent.png'" mode="aspectFill" />
        <view class="agent-info">
          <view class="agent-name">{{ agent.name }}</view>
          <view class="agent-desc">{{ agent.description || '暂无描述' }}</view>
          <view class="agent-meta">
            <text class="meta-item">
              <text class="meta-label">模型：</text>
              <text class="meta-value">{{ agent.model_id }}</text>
            </text>
          </view>
        </view>
        <view class="agent-status">
          <view class="status-dot" :class="agent.status === 'active' ? 'active' : 'inactive'"></view>
          <text class="status-text">{{ agent.status === 'active' ? '可用' : '停用' }}</text>
        </view>
      </view>

      <view v-if="agents.length === 0 && !loading" class="empty-state">
        <image class="empty-icon" src="/static/empty.png" mode="aspectFit" />
        <text class="empty-text">暂无智能体</text>
      </view>
    </view>

    <view v-if="loading" class="loading-wrapper">
      <text class="loading-text">加载中...</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { getAgents, type Agent } from '@/api/agent'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

const agents = ref<Agent[]>([])
const loading = ref(false)
const keyword = ref('')
const page = ref(1)
const pageSize = ref(20)

const industries = ['全部', '电商', '餐饮', '法律', '教育', '医疗', '金融', '物流', '房产', '零售', '娱乐']
const selectedIndustry = ref('全部')

let searchTimer: ReturnType<typeof setTimeout> | null = null

onMounted(() => {
  // 检查登录状态
  if (!userStore.isLoggedIn) {
    uni.redirectTo({
      url: '/pages/login/index'
    })
    return
  }

  loadAgents()
})

onUnmounted(() => {
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
})

// 下拉刷新
uni.$on('onPullDownRefresh', () => {
  page.value = 1
  loadAgents().finally(() => {
    uni.stopPullDownRefresh()
  })
})

async function loadAgents() {
  loading.value = true
  try {
    const res = await getAgents({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      industry: selectedIndustry.value === '全部' ? undefined : selectedIndustry.value,
      status: 'active'
    })
    agents.value = res.items
  } catch (error) {
    uni.showToast({
      title: '加载失败',
      icon: 'none'
    })
  } finally {
    loading.value = false
  }
}

// 输入即搜索（防抖 300ms）
function handleSearchInput() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    loadAgents()
  }, 300)
}

function handleSearch() {
  if (searchTimer) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
  page.value = 1
  loadAgents()
}

function handleSelectIndustry(industry: string) {
  if (selectedIndustry.value === industry) return
  selectedIndustry.value = industry
  page.value = 1
  loadAgents()
}

function goToChat(agent: Agent) {
  if (agent.status !== 'active') {
    uni.showToast({
      title: '该智能体已停用',
      icon: 'none'
    })
    return
  }

  uni.navigateTo({
    url: `/pages/chat/index?agent_id=${agent.id}&agent_name=${agent.name}`
  })
}
</script>

<style lang="scss" scoped>
.agent-list-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.search-bar {
  padding: 20rpx;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 10;

  .search-input {
    width: 100%;
    height: 72rpx;
    background: #f5f7fa;
    border-radius: 36rpx;
    padding: 0 32rpx;
    font-size: 28rpx;
    box-sizing: border-box;
  }
}

.agent-list {
  padding: 20rpx;
}

.agent-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  display: flex;
  align-items: center;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);

  &:active {
    background: #f9f9f9;
  }
}

.agent-avatar {
  width: 96rpx;
  height: 96rpx;
  border-radius: 50%;
  margin-right: 24rpx;
  flex-shrink: 0;
}

.agent-info {
  flex: 1;
  min-width: 0;

  .agent-name {
    font-size: 32rpx;
    font-weight: 600;
    color: #333;
    margin-bottom: 8rpx;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .agent-desc {
    font-size: 26rpx;
    color: #999;
    margin-bottom: 12rpx;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .agent-meta {
    .meta-item {
      font-size: 24rpx;
      color: #666;

      .meta-label {
        color: #999;
      }

      .meta-value {
        color: #409eff;
      }
    }
  }
}

.agent-status {
  display: flex;
  flex-direction: column;
  align-items: center;
  margin-left: 20rpx;

  .status-dot {
    width: 16rpx;
    height: 16rpx;
    border-radius: 50%;
    margin-bottom: 8rpx;

    &.active {
      background: #67c23a;
    }

    &.inactive {
      background: #c0c4cc;
    }
  }

  .status-text {
    font-size: 22rpx;
    color: #999;
  }
}

.empty-state {
  text-align: center;
  padding: 120rpx 0;

  .empty-icon {
    width: 200rpx;
    height: 200rpx;
    margin-bottom: 24rpx;
  }

  .empty-text {
    font-size: 28rpx;
    color: #999;
  }
}

.loading-wrapper {
  text-align: center;
  padding: 40rpx 0;

  .loading-text {
    font-size: 28rpx;
    color: #999;
  }
}
</style>
