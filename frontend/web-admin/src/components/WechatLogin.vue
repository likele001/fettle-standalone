<template>
  <div class="wechat-login">
    <el-button type="success" @click="showWechatLogin" :loading="loading">
      <el-icon><ChatDotRound /></el-icon>
      微信扫码登录
    </el-button>

    <el-dialog
      v-model="dialogVisible"
      title="微信扫码登录"
      width="400px"
      :close-on-click-modal="false"
    >
      <div class="wechat-qrcode">
        <div v-if="loading" class="loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <p>正在生成二维码...</p>
        </div>
        <div v-else-if="qrcodeUrl" class="qrcode">
          <img :src="qrcodeUrl" alt="微信登录二维码" />
          <p class="tip">请使用微信扫描二维码登录</p>
        </div>
        <div v-else class="error">
          <el-icon><WarningFilled /></el-icon>
          <p>二维码生成失败</p>
          <el-button @click="generateQrcode">重新生成</el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ChatDotRound, Loading, WarningFilled } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()

const dialogVisible = ref(false)
const loading = ref(false)
const qrcodeUrl = ref('')
const checkInterval = ref<number | null>(null)

function showWechatLogin() {
  dialogVisible.value = true
  generateQrcode()
}

async function generateQrcode() {
  loading.value = true
  qrcodeUrl.value = ''

  try {
    // TODO: 调用后端API生成微信登录二维码
    // const response = await api.get('/auth/wechat/qrcode')
    // qrcodeUrl.value = response.qrcode_url

    // 模拟数据
    await new Promise(resolve => setTimeout(resolve, 1000))
    qrcodeUrl.value = 'https://via.placeholder.com/300?text=WeChat+QR+Code'

    // 开始轮询检查登录状态
    startPolling()
  } catch (error) {
    ElMessage.error('生成二维码失败')
  } finally {
    loading.value = false
  }
}

function startPolling() {
  // 每3秒检查一次登录状态
  checkInterval.value = window.setInterval(async () => {
    try {
      // TODO: 调用后端API检查扫码状态
      // const response = await api.get('/auth/wechat/check')
      // if (response.logged_in) {
      //   handleLoginSuccess(response)
      // }

      // 模拟登录成功
      const mockResponse = {
        tokens: {
          access_token: 'mock_access_token',
          refresh_token: 'mock_refresh_token',
          expires_in: 900
        },
        user: {
          id: 'mock_user_id',
          phone: '13800138000',
          name: '微信用户',
          role: 'super_admin',
          tenant_id: 'mock_tenant_id'
        }
      }

      handleLoginSuccess(mockResponse)
    } catch (error) {
      console.error('检查登录状态失败', error)
    }
  }, 3000)
}

function handleLoginSuccess(response: any) {
  if (checkInterval.value) {
    clearInterval(checkInterval.value)
  }

  userStore.setTokens(response.tokens)
  userStore.setUser(response.user)

  dialogVisible.value = false
  ElMessage.success('登录成功')
  router.push('/')
}

// 组件卸载时清理定时器
import { onUnmounted } from 'vue'
onUnmounted(() => {
  if (checkInterval.value) {
    clearInterval(checkInterval.value)
  }
})
</script>

<style scoped lang="scss">
.wechat-login {
  .wechat-qrcode {
    min-height: 350px;
    display: flex;
    align-items: center;
    justify-content: center;

    .loading,
    .error {
      text-align: center;

      .el-icon {
        font-size: 48px;
        color: #909399;
        margin-bottom: 16px;
      }

      p {
        color: #606266;
        margin: 8px 0 16px;
      }
    }

    .qrcode {
      text-align: center;

      img {
        width: 300px;
        height: 300px;
        border: 1px solid #dcdfe6;
        border-radius: 4px;
      }

      .tip {
        color: #909399;
        font-size: 14px;
        margin-top: 16px;
      }
    }
  }
}
</style>
