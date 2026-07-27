<template>
  <view class="history-page">
    <view class="conversation-list">
      <view
        v-for="item in conversations"
        :key="item.id"
        class="conversation-card"
        @click="goToChat(item)"
      >
        <view class="conversation-info">
          <view class="conversation-header">
            <text class="agent-name">{{ item.agent_name }}</text>
            <text class="conversation-time">{{ formatTime(item.updated_at) }}</text>
          </view>
          <view class="conversation-summary">{{ item.last_message || '暂无消息' }}</view>
        </view>
      </view>

      <view v-if="conversations.length === 0 && !loading" class="empty-state">
        <image class="empty-icon" src="/static/empty.png" mode="aspectFit" />
        <text class="empty-text">暂无对话记录</text>
        <text class="empty-sub-text">快去和智能体开始对话吧</text>
      </view>
    </view>

    <view v-if="loading" class="loading-wrapper">
      <text class="loading-text">加载中...</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { onPullDownRefresh } from '@dcloudio/uni-app'
import { getConversations, type Conversation } from '@/api/chat'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

const conversations = ref<Conversation[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = ref(20)

onMounted(() => {
  // 检查登录状态
  if (!userStore.isLoggedIn) {
    uni.redirectTo({
      url: '/pages/login/index'
    })
    return
  }

  loadConversations()
})

// 下拉刷新
onPullDownRefresh(() => {
  page.value = 1
  loadConversations().finally(() => {
    uni.stopPullDownRefresh()
  })
})

async function loadConversations() {
  loading.value = true
  try {
    const res = await getConversations({
      page: page.value,
      page_size: pageSize.value
    })
    // 按时间倒序排列
    conversations.value = res.items.sort((a, b) => {
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
    })
  } catch (error) {
    uni.showToast({
      title: '加载失败',
      icon: 'none'
    })
  } finally {
    loading.value = false
  }
}

function goToChat(item: Conversation) {
  uni.navigateTo({
    url: `/pages/chat/index?agent_id=${item.agent_id}&agent_name=${encodeURIComponent(item.agent_name)}&conversation_id=${item.id}`
  })
}

function formatTime(time: string) {
  const date = new Date(time)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes}分钟前`
  if (hours < 24) return `${hours}小时前`
  if (days < 2) return '昨天'
  if (days < 7) return `${days}天前`

  const month = (date.getMonth() + 1).toString().padStart(2, '0')
  const day = date.getDate().toString().padStart(2, '0')
  if (date.getFullYear() === now.getFullYear()) {
    return `${month}-${day}`
  }
  return `${date.getFullYear()}-${month}-${day}`
}
</script>

<style lang="scss" scoped>
.history-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 40rpx;
}

.conversation-list {
  padding: 20rpx;
}

.conversation-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);

  &:active {
    background: #f9f9f9;
  }
}

.conversation-info {
  .conversation-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 12rpx;

    .agent-name {
      font-size: 32rpx;
      font-weight: 600;
      color: #333;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      flex: 1;
      min-width: 0;
    }

    .conversation-time {
      font-size: 24rpx;
      color: #999;
      flex-shrink: 0;
      margin-left: 20rpx;
    }
  }

  .conversation-summary {
    font-size: 26rpx;
    color: #999;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
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
    display: block;
    font-size: 28rpx;
    color: #999;
    margin-bottom: 12rpx;
  }

  .empty-sub-text {
    display: block;
    font-size: 24rpx;
    color: #bbb;
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
