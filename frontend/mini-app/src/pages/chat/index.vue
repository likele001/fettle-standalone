<template>
  <view class="chat-page">
    <view class="chat-header">
      <text class="agent-name">{{ agentName }}</text>
    </view>

    <scroll-view
      class="chat-messages"
      scroll-y
      :scroll-top="scrollTop"
      :scroll-with-animation="scrollWithAnimation"
      :upper-threshold="50"
      @scrolltoupper="loadMoreMessages"
    >
      <view v-if="loadingMore" class="loading-more-wrapper">
        <text class="loading-more-text">正在加载更多...</text>
      </view>
      <view v-else-if="!hasMore && messages.length > 0" class="no-more-wrapper">
        <text class="no-more-text">没有更多消息了</text>
      </view>

      <view v-if="loading" class="loading-wrapper">
        <text class="loading-text">加载中...</text>
      </view>

      <view
        v-for="message in messages"
        :key="message.id"
        class="message-item"
        :class="message.role"
      >
        <view class="message-avatar">
          <image
            v-if="message.role === 'user'"
            src="/static/user-avatar.png"
            mode="aspectFill"
          />
          <image
            v-else
            src="/static/ai-avatar.png"
            mode="aspectFill"
          />
        </view>
        <view class="message-content">
          <text class="message-text">{{ message.content }}</text>
          <text class="message-time">{{ formatTime(message.created_at) }}</text>
        </view>
      </view>

      <view v-if="messages.length === 0 && !loading" class="empty-chat">
        <text class="empty-text">开始对话吧</text>
      </view>
    </scroll-view>

    <view class="chat-input-bar">
      <input
        v-model="inputMessage"
        type="text"
        placeholder="输入消息..."
        class="chat-input"
        :adjust-position="true"
        @confirm="sendMessage"
      />
      <button class="send-btn" :disabled="!inputMessage.trim()" @click="sendMessage">
        发送
      </button>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue'
import { getMessages, sendMessage as sendMessageApi, createConversation, type Message } from '@/api/chat'

const agentId = ref('')
const agentName = ref('')
const conversationId = ref('')
const messages = ref<Message[]>([])
const inputMessage = ref('')
const loading = ref(false)
const scrollTop = ref(0)
const scrollWithAnimation = ref(true)

// 分页加载相关状态
const PAGE_SIZE = 20
const currentPage = ref(1)
const hasMore = ref(true)
const loadingMore = ref(false)

onMounted(() => {
  // 获取页面参数
  const pages = getCurrentPages()
  const currentPage = pages[pages.length - 1] as any
  const options = currentPage.$page?.options || currentPage.options

  agentId.value = options.agent_id || ''
  agentName.value = decodeURIComponent(options.agent_name || 'AI 助手')
  conversationId.value = options.conversation_id || ''

  if (!agentId.value) {
    uni.showToast({
      title: '参数错误',
      icon: 'none'
    })
    setTimeout(() => {
      uni.navigateBack()
    }, 1500)
    return
  }

  initConversation()
})

async function initConversation() {
  try {
    if (conversationId.value) {
      // 已有会话ID，直接加载历史消息（续聊）
      await loadMessages()
    } else {
      // 无会话ID，创建新会话
      const conversation = await createConversation(agentId.value)
      conversationId.value = conversation.id
      await loadMessages()
    }
  } catch (error) {
    uni.showToast({
      title: '初始化失败',
      icon: 'none'
    })
  }
}

async function loadMessages() {
  if (!conversationId.value) return

  loading.value = true
  try {
    currentPage.value = 1
    hasMore.value = true
    const res = await getMessages(conversationId.value, currentPage.value, PAGE_SIZE)
    messages.value = res.items.reverse() // 最新的消息在底部
    hasMore.value = res.items.length >= PAGE_SIZE && messages.value.length < res.total
    await scrollToBottom()
  } catch (error) {
    console.error('加载消息失败', error)
  } finally {
    loading.value = false
  }
}

async function loadMoreMessages() {
  if (!conversationId.value || loadingMore.value || !hasMore.value || loading.value) return

  loadingMore.value = true
  scrollWithAnimation.value = false

  try {
    // 记录加载前的内容高度，用于保持滚动位置
    const oldScrollHeight = await getScrollHeight()

    currentPage.value += 1
    const res = await getMessages(conversationId.value, currentPage.value, PAGE_SIZE)

    if (res.items.length === 0) {
      hasMore.value = false
      return
    }

    // 将更早的消息插入到顶部（API 返回最新在前，reverse 后按时间正序）
    const olderMessages = res.items.reverse()
    messages.value = [...olderMessages, ...messages.value]
    hasMore.value = res.items.length >= PAGE_SIZE && messages.value.length < res.total

    // 等待 DOM 更新完成
    await nextTick()
    await new Promise(resolve => setTimeout(resolve, 50))

    // 计算新的滚动位置：保持用户看到的内容位置不变
    const newScrollHeight = await getScrollHeight()
    const scrollOffset = newScrollHeight - oldScrollHeight

    // 先重置为 0 再设置目标值，确保 scroll-view 能响应 scrollTop 的变化
    scrollTop.value = 0
    await nextTick()
    scrollTop.value = scrollOffset
  } catch (error) {
    console.error('加载更多消息失败', error)
    currentPage.value -= 1
  } finally {
    loadingMore.value = false
    scrollWithAnimation.value = true
  }
}

function getScrollHeight(): Promise<number> {
  return new Promise((resolve) => {
    uni.createSelectorQuery()
      .select('.chat-messages')
      .fields({ scrollHeight: true }, (res) => {
        resolve((res as any)?.scrollHeight || 0)
      })
      .exec()
  })
}

async function sendMessage() {
  const content = inputMessage.value.trim()
  if (!content || !conversationId.value) return

  // 添加用户消息到列表
  const userMessage: Message = {
    id: Date.now().toString(),
    conversation_id: conversationId.value,
    role: 'user',
    content,
    created_at: new Date().toISOString()
  }
  messages.value.push(userMessage)
  inputMessage.value = ''
  await scrollToBottom()

  try {
    // 发送消息到服务器
    const response = await sendMessageApi({
      conversation_id: conversationId.value,
      content
    })

    // 添加 AI 回复到列表
    messages.value.push(response)
    await scrollToBottom()
  } catch (error) {
    uni.showToast({
      title: '发送失败',
      icon: 'none'
    })
  }
}

async function scrollToBottom() {
  await nextTick()
  scrollTop.value = 999999
}

function formatTime(time: string) {
  const date = new Date(time)
  const hours = date.getHours().toString().padStart(2, '0')
  const minutes = date.getMinutes().toString().padStart(2, '0')
  return `${hours}:${minutes}`
}
</script>

<style lang="scss" scoped>
.chat-page {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f5f5;
}

.chat-header {
  height: 88rpx;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1rpx solid #e4e7ed;

  .agent-name {
    font-size: 32rpx;
    font-weight: 600;
    color: #333;
  }
}

.chat-messages {
  flex: 1;
  padding: 20rpx;
  overflow-y: auto;
}

.message-item {
  display: flex;
  margin-bottom: 24rpx;

  &.user {
    flex-direction: row-reverse;

    .message-content {
      align-items: flex-end;

      .message-text {
        background: #409eff;
        color: #fff;
        border-radius: 16rpx 4rpx 16rpx 16rpx;
      }
    }
  }

  &.assistant {
    .message-content {
      align-items: flex-start;

      .message-text {
        background: #fff;
        color: #333;
        border-radius: 4rpx 16rpx 16rpx 16rpx;
      }
    }
  }
}

.message-avatar {
  width: 72rpx;
  height: 72rpx;
  border-radius: 50%;
  margin: 0 16rpx;
  flex-shrink: 0;

  image {
    width: 100%;
    height: 100%;
    border-radius: 50%;
  }
}

.message-content {
  display: flex;
  flex-direction: column;
  max-width: 70%;

  .message-text {
    padding: 20rpx 24rpx;
    font-size: 28rpx;
    line-height: 1.6;
    word-break: break-all;
  }

  .message-time {
    font-size: 22rpx;
    color: #999;
    margin-top: 8rpx;
  }
}

.empty-chat {
  text-align: center;
  padding: 120rpx 0;

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

.loading-more-wrapper {
  text-align: center;
  padding: 24rpx 0;

  .loading-more-text {
    font-size: 26rpx;
    color: #999;
  }
}

.no-more-wrapper {
  text-align: center;
  padding: 24rpx 0;

  .no-more-text {
    font-size: 26rpx;
    color: #c0c4cc;
  }
}

.chat-input-bar {
  display: flex;
  align-items: center;
  padding: 20rpx;
  background: #fff;
  border-top: 1rpx solid #e4e7ed;

  .chat-input {
    flex: 1;
    height: 72rpx;
    background: #f5f7fa;
    border-radius: 36rpx;
    padding: 0 32rpx;
    font-size: 28rpx;
    margin-right: 20rpx;
  }

  .send-btn {
    width: 140rpx;
    height: 72rpx;
    background: #409eff;
    color: #fff;
    border: none;
    border-radius: 36rpx;
    font-size: 28rpx;
    font-weight: 500;
    padding: 0;

    &[disabled] {
      background: #c0c4cc;
    }
  }
}
</style>
