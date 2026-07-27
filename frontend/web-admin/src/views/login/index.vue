<template>
  <div class="auth-page">
    <el-card class="auth-card" shadow="hover">
      <div class="auth-header">
        <h2>AI 智能体平台</h2>
        <p>管理后台登录</p>
      </div>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-position="top"
        @keyup.enter="handleSubmit"
      >
        <el-form-item label="手机号 / 邮箱" prop="account">
          <el-input
            v-model="form.account"
            placeholder="请输入手机号或邮箱"
            :prefix-icon="User"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="请输入密码"
            :prefix-icon="Lock"
            show-password
          />
        </el-form-item>

        <el-form-item label="验证码" prop="captcha_code">
          <div class="captcha-row">
            <el-input
              v-model="form.captcha_code"
              placeholder="请输入验证码"
              maxlength="6"
              style="flex: 1"
            />
            <img
              :src="captchaImage"
              class="captcha-img"
              @click="refreshCaptcha"
              title="点击刷新验证码"
            />
          </div>
        </el-form-item>

        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            @click="handleSubmit"
            style="width: 100%"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>

      <div class="auth-footer">
        <span>还没有账号？</span>
        <el-link type="primary" @click="$router.push('/register')">立即注册</el-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'
import { login, getCaptcha } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { isTokenExpired } from '@/utils/token'
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const userStore = useUserStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const captchaImage = ref('')
const captchaId = ref('')

const form = reactive({
  account: '',
  password: '',
  captcha_code: ''
})

const rules: FormRules = {
  account: [
    { required: true, message: '请输入手机号或邮箱', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' }
  ],
  captcha_code: [
    { required: true, message: '请输入验证码', trigger: 'blur' }
  ]
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

onMounted(async () => {
  refreshCaptcha()
})

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    loading.value = true
    try {
      const res = await login({
        account: form.account,
        password: form.password,
        captcha_id: captchaId.value,
        captcha_code: form.captcha_code
      })
      userStore.setTokens(res.tokens)
      userStore.setUser(res.user)
      ElMessage.success('登录成功')
      router.replace({ name: 'Dashboard' })
    } catch (e: any) {
      refreshCaptcha()
      form.captcha_code = ''
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped lang="scss">
.auth-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1a2e 0%, #2d2d44 100%);
}

.auth-card {
  width: min(92%, 420px);
  border-radius: 12px;
}

.auth-header {
  text-align: center;
  margin-bottom: 24px;

  h2 {
    font-size: 24px;
    color: $text-primary;
    margin-bottom: 8px;
  }

  p {
    color: $text-secondary;
    font-size: 14px;
  }
}

.auth-footer {
  text-align: center;
  margin-top: 16px;
  font-size: 14px;
  color: $text-regular;
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

}
