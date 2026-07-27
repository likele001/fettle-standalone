<template>
  <view class="profile-page">
    <view class="user-card">
      <image class="user-avatar" :src="userInfo?.avatar || '/static/default-avatar.png'" mode="aspectFill" />
      <view class="user-info">
        <text class="user-name">{{ userInfo?.name || '未登录' }}</text>
        <text class="user-role">{{ userInfo?.tenant_name || '个人用户' }}</text>
      </view>
    </view>

    <view class="menu-list">
      <view class="menu-item" @click="goToConversations">
        <text class="menu-label">我的会话</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @click="goToKnowledge">
        <text class="menu-label">知识库</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @click="showAbout">
        <text class="menu-label">关于我们</text>
        <text class="menu-arrow">></text>
      </view>
      <view class="menu-item" @click="handleLogout">
        <text class="menu-label text-danger">退出登录</text>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()
const userInfo = computed(() => userStore.userInfo)

function goToConversations() {
  uni.showToast({
    title: '功能开发中',
    icon: 'none'
  })
}

function goToKnowledge() {
  uni.switchTab({
    url: '/pages/knowledge/index'
  })
}

function showAbout() {
  uni.showModal({
    title: '关于我们',
    content: 'Fettle AI 智能体平台\n版本：1.0.0\n\n让 AI 为您的业务赋能',
    showCancel: false
  })
}

function handleLogout() {
  uni.showModal({
    title: '提示',
    content: '确定要退出登录吗？',
    success: (res) => {
      if (res.confirm) {
        userStore.logout()
        uni.showToast({
          title: '已退出登录',
          icon: 'success'
        })
      }
    }
  })
}
</script>

<style lang="scss" scoped>
.profile-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.user-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 60rpx 40rpx;
  display: flex;
  align-items: center;

  .user-avatar {
    width: 120rpx;
    height: 120rpx;
    border-radius: 50%;
    margin-right: 24rpx;
    border: 4rpx solid rgba(255, 255, 255, 0.3);
  }

  .user-info {
    flex: 1;

    .user-name {
      display: block;
      font-size: 36rpx;
      font-weight: 600;
      color: #fff;
      margin-bottom: 8rpx;
    }

    .user-role {
      display: block;
      font-size: 26rpx;
      color: rgba(255, 255, 255, 0.8);
    }
  }
}

.menu-list {
  margin-top: 20rpx;
  background: #fff;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 32rpx 40rpx;
  border-bottom: 1rpx solid #f0f0f0;

  &:active {
    background: #f9f9f9;
  }

  &:last-child {
    border-bottom: none;
  }

  .menu-label {
    font-size: 30rpx;
    color: #333;

    &.text-danger {
      color: #f56c6c;
    }
  }

  .menu-arrow {
    font-size: 28rpx;
    color: #999;
  }
}
</style>
