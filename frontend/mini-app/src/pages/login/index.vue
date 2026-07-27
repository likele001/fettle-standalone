<template>
  <view class="login-page">
    <!-- Logo 和标题 -->
    <view class="login-header">
      <image class="logo" src="/static/logo.png" mode="aspectFit" />
      <text class="title">Fettle AI</text>
      <text class="subtitle">AI 智能体平台</text>
    </view>

    <!-- 租户编码输入 + 微信一键登录 -->
    <view class="login-card">
      <view class="card-title">登录</view>

      <view class="form-item">
        <text class="label">租户编码</text>
        <input
          v-model="tenantCode"
          class="input"
          placeholder="请输入租户编码（如 gxkj）"
        />
      </view>

      <!-- 微信一键登录 -->
      <button
        class="btn-primary wx-btn"
        :loading="loading"
        @click="handleWxLogin"
      >
        微信一键登录
      </button>

      <text class="wx-hint">
        首次使用微信登录后需绑定已有账号
      </text>
    </view>

    <!-- 绑定账号（未绑定用户） -->
    <view v-if="showBind" class="login-card bind-card">
      <view class="card-title">绑定已有账号</view>
      <text class="bind-hint">
        您的微信尚未绑定账号，请输入已有账号信息完成绑定
      </text>

      <view class="form-item">
        <text class="label">租户编码</text>
        <input
          v-model="tenantCode"
          class="input"
          placeholder="请输入租户编码"
          disabled
        />
      </view>

      <view class="form-item">
        <text class="label">用户名</text>
        <input
          v-model="bindUsername"
          class="input"
          placeholder="请输入用户名"
        />
      </view>

      <view class="form-item">
        <text class="label">密码</text>
        <input
          v-model="bindPassword"
          class="input"
          password
          placeholder="请输入密码"
        />
      </view>

      <button
        class="btn-primary"
        :loading="binding"
        @click="handleBind"
      >
        绑定并登录
      </button>
    </view>

    <!-- 底部协议 -->
    <view class="login-footer">
      <text class="footer-text">登录即代表同意</text>
      <text class="footer-link">《用户协议》</text>
      <text class="footer-text">和</text>
      <text class="footer-link">《隐私政策》</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useUserStore } from '@/store/user'
import { miniappLogin, bindOpenid, fetchMiniConfig } from '@/api/auth'

const userStore = useUserStore()

const tenantCode = ref('')
const loading = ref(false)
const showBind = ref(false)
const bindUsername = ref('')
const bindPassword = ref('')
const binding = ref(false)
const pendingOpenid = ref('')

onMounted(async () => {
  // 预加载小程序配置
  try {
    const config = await fetchMiniConfig()
    console.log('小程序配置:', config)
  } catch (e) {
    console.error('获取小程序配置失败:', e)
  }

  // 恢复上次输入的租户编码
  const saved = uni.getStorageSync('fettle_tenant_code')
  if (saved) {
    tenantCode.value = saved
  }
})

// 微信一键登录
async function handleWxLogin() {
  if (!tenantCode.value.trim()) {
    uni.showToast({ title: '请输入租户编码', icon: 'none' })
    return
  }

  const code = tenantCode.value.trim().toLowerCase()
  uni.setStorageSync('fettle_tenant_code', code)

  loading.value = true
  try {
    // 微信登录获取 code
    const loginRes = await uni.login({ provider: 'weixin' })
    if (!loginRes.code) {
      throw new Error('获取微信授权失败')
    }

    const res = await miniappLogin(loginRes.code, code)

    if (res.need_bind) {
      // 未绑定，显示绑定表单
      pendingOpenid.value = res.openid
      showBind.value = true
      loading.value = false
      return
    }

    if (res.token) {
      // 登录成功
      await userStore.saveLogin(res.token)
      await userStore.fetchUser()
      goHome()
    }
  } catch (e: any) {
    uni.showToast({
      title: e.message || '登录失败，请重试',
      icon: 'none',
      duration: 2000
    })
  } finally {
    if (!showBind.value) {
      loading.value = false
    }
  }
}

// 绑定已有账号
async function handleBind() {
  if (!bindUsername.value.trim() || !bindPassword.value) {
    uni.showToast({ title: '请输入用户名和密码', icon: 'none' })
    return
  }

  const code = tenantCode.value.trim().toLowerCase()

  binding.value = true
  try {
    const res = await bindOpenid({
      username: bindUsername.value.trim(),
      password: bindPassword.value,
      openid: pendingOpenid.value,
      tenant_code: code,
    })

    if (res.token) {
      await userStore.saveLogin(res.token)
      await userStore.fetchUser()
      goHome()
    }
  } catch (e: any) {
    uni.showToast({
      title: e.message || '绑定失败',
      icon: 'none',
      duration: 2000
    })
  } finally {
    binding.value = false
  }
}

function goHome() {
  uni.reLaunch({ url: '/pages/index/index' })
}
</script>

<style lang="scss" scoped>
.login-page {
  min-height: 100vh;
  padding: 80rpx 40rpx 40rpx;
  background: linear-gradient(180deg, #eff6ff 0%, #f8fafc 100%);
}

.login-header {
  text-align: center;
  margin-bottom: 60rpx;
}

.logo {
  width: 120rpx;
  height: 120rpx;
  margin-bottom: 24rpx;
}

.title {
  display: block;
  font-size: 48rpx;
  font-weight: 700;
  color: #1a1a2e;
  margin-bottom: 8rpx;
}

.subtitle {
  display: block;
  font-size: 26rpx;
  color: #64748b;
}

.login-card {
  background: #fff;
  border-radius: 24rpx;
  padding: 40rpx;
  margin-bottom: 24rpx;
  box-shadow: 0 4rpx 24rpx rgba(0, 0, 0, 0.06);
}

.card-title {
  font-size: 32rpx;
  font-weight: 600;
  color: #1a1a2e;
  margin-bottom: 32rpx;
}

.form-item {
  margin-bottom: 24rpx;
}

.label {
  display: block;
  font-size: 24rpx;
  color: #64748b;
  margin-bottom: 12rpx;
}

.input {
  background: #f1f5f9;
  border-radius: 12rpx;
  padding: 24rpx;
  font-size: 28rpx;
  width: 100%;
  box-sizing: border-box;
}

.btn-primary {
  background: linear-gradient(135deg, #3b82f6, #1d4ed8);
  color: #fff;
  border: none;
  border-radius: 12rpx;
  padding: 24rpx;
  font-size: 30rpx;
  font-weight: 500;
  width: 100%;
}

.wx-btn {
  background: linear-gradient(135deg, #07c160, #05a34f);
  margin-top: 8rpx;
}

.wx-hint {
  display: block;
  font-size: 22rpx;
  color: #94a3b8;
  text-align: center;
  margin-top: 16rpx;
}

.bind-card {
  border: 2rpx solid #e2e8f0;
}

.bind-hint {
  display: block;
  font-size: 24rpx;
  color: #64748b;
  margin-bottom: 24rpx;
  line-height: 1.6;
}

.login-footer {
  text-align: center;
  margin-top: 40rpx;
}

.footer-text {
  font-size: 22rpx;
  color: #94a3b8;
}

.footer-link {
  font-size: 22rpx;
  color: #3b82f6;
}
</style>
