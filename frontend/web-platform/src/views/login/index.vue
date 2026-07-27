<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <h1>Fettle Admin</h1>
        <p>平台管理后台</p>
      </div>
      <el-form ref="formRef" :model="form" :rules="rules" size="large" @submit.prevent="handleLogin">
        <el-form-item prop="account">
          <el-input v-model="form.account" placeholder="手机号 / 邮箱" :prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" :prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item prop="captcha_code">
          <div class="captcha-row">
            <el-input v-model="form.captcha_code" placeholder="验证码" maxlength="6" style="flex: 1" />
            <img :src="captchaImage" class="captcha-img" @click="refreshCaptcha" title="点击刷新" />
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="loading" class="login-btn">
            登 录
          </el-button>
        </el-form-item>
      </el-form>
      <div v-if="errorMsg" class="login-error">{{ errorMsg }}</div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import { login as loginApi, getCaptcha } from '@/api/auth'
import type { FormInstance } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const errorMsg = ref('')
const captchaImage = ref('')
const captchaId = ref('')

const form = reactive({
  account: '',
  password: '',
  captcha_code: ''
})

const rules = {
  account: [{ required: true, message: '请输入手机号或邮箱', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  captcha_code: [{ required: true, message: '请输入验证码', trigger: 'blur' }]
}

async function refreshCaptcha() {
  try {
    const res = await getCaptcha()
    captchaId.value = res.captcha_id
    captchaImage.value = res.captcha_image
  } catch (e: any) {
    ElMessage.error('获取验证码失败')
  }
}

onMounted(() => {
  refreshCaptcha()
})

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  errorMsg.value = ''

  try {
    const res = await loginApi({
      account: form.account,
      password: form.password,
      captcha_id: captchaId.value,
      captcha_code: form.captcha_code
    })

    if (res.user.role !== 'super_admin') {
      errorMsg.value = '该账号无权访问管理后台'
      return
    }

    userStore.setTokens(res.tokens)
    userStore.setUser(res.user)
    ElMessage.success('登录成功')
    router.push('/')
  } catch (e: any) {
    errorMsg.value = e.message || '登录失败，请检查账号和密码'
    refreshCaptcha()
    form.captcha_code = ''
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.login-page {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1e293b 0%, #334155 100%);
}

.login-card {
  width: 400px;
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0,0,0,0.3);

  .login-header {
    text-align: center;
    margin-bottom: 32px;

    h1 {
      font-size: 24px;
      font-weight: 700;
      color: #1e293b;
    }

    p {
      font-size: 14px;
      color: #64748b;
      margin-top: 6px;
    }
  }

  .login-btn {
    width: 100%;
  }

  .login-error {
    color: $admin-danger;
    font-size: 13px;
    text-align: center;
    margin-top: 12px;
  }
}

.captcha-row {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 100%;

  .captcha-img {
    height: 40px;
    border-radius: 4px;
    cursor: pointer;
    border: 1px solid #dcdfe6;
  }
}
</style>
