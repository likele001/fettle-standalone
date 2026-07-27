<template>
  <div class="page-container">
    <div class="page-header">
      <h2>平台配置</h2>
      <el-button type="primary" @click="handleSave" :loading="saving">
        保存配置
      </el-button>
    </div>

    <el-form :model="form" label-width="160px" v-loading="loading">
      <!-- 基础配置 -->
      <el-card class="section-card">
        <template #header>
          <span style="font-weight: 600;">基础配置</span>
        </template>
        <el-form-item label="平台名称">
          <el-input v-model="form.site_name" placeholder="请输入平台名称" />
        </el-form-item>
        <el-form-item label="平台 Logo">
          <el-input v-model="form.site_logo" placeholder="请输入 Logo URL" />
        </el-form-item>
        <el-form-item label="联系邮箱">
          <el-input v-model="form.contact_email" placeholder="请输入联系邮箱" />
        </el-form-item>
        <el-form-item label="维护模式">
          <el-switch v-model="form.maintenance_mode" />
          <span style="margin-left: 12px; color: #94a3b8; font-size: 12px;">
            开启后所有租户无法访问平台
          </span>
        </el-form-item>
      </el-card>

      <!-- 微信小程序配置 -->
      <el-card class="section-card">
        <template #header>
          <span style="font-weight: 600;">微信小程序配置</span>
        </template>
        <el-form-item label="小程序 AppID">
          <el-input
            v-model="form.wechat_miniapp_id"
            placeholder="请输入微信小程序 AppID（如 wx1234567890abcdef）"
          />
        </el-form-item>
        <el-form-item label="小程序 AppSecret">
          <el-input
            v-model="form.wechat_miniapp_secret"
            type="password"
            show-password
            placeholder="请输入微信小程序 AppSecret"
          />
        </el-form-item>
        <div style="margin-left: 160px; color: #94a3b8; font-size: 12px; margin-bottom: 12px;">
          配置后小程序端将使用微信一键登录。在
          <a href="https://mp.weixin.qq.com" target="_blank" style="color: #409eff;">微信公众平台</a>
          → 开发管理 → 开发设置 中获取。
        </div>
      </el-card>

      <!-- 上传配置 -->
      <el-card class="section-card">
        <template #header>
          <span style="font-weight: 600;">上传配置</span>
        </template>
        <el-form-item label="最大上传大小">
          <el-input-number v-model="form.max_upload_size" :min="1" :max="100" />
          <span style="margin-left: 12px; color: #94a3b8;">MB</span>
        </el-form-item>
        <el-form-item label="允许文件类型">
          <el-select
            v-model="form.allowed_file_types"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="输入文件类型后回车添加"
            style="width: 100%"
          />
          <div style="margin-top: 8px; color: #94a3b8; font-size: 12px;">
            常用类型：pdf, doc, docx, txt, md, csv, xlsx
          </div>
        </el-form-item>
      </el-card>

      <!-- 短信配置 -->
      <el-card class="section-card">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <span style="font-weight: 600;">短信服务商配置</span>
            <el-button size="small" @click="testSms" :loading="testingSms">
              测试发送
            </el-button>
          </div>
        </template>
        <el-form-item label="短信服务商">
          <el-select v-model="form.sms_provider" placeholder="选择服务商">
            <el-option label="阿里云短信" value="aliyun" />
            <el-option label="腾讯云短信" value="tencent" />
            <el-option label="华为云短信" value="huawei" />
          </el-select>
        </el-form-item>
        <el-form-item label="Access Key">
          <el-input v-model="form.sms_config.access_key" placeholder="请输入 Access Key" />
        </el-form-item>
        <el-form-item label="Secret Key">
          <el-input
            v-model="form.sms_config.secret_key"
            type="password"
            show-password
            placeholder="请输入 Secret Key"
          />
        </el-form-item>
        <el-form-item label="签名">
          <el-input v-model="form.sms_config.sign_name" placeholder="请输入短信签名" />
        </el-form-item>
        <el-form-item label="模板 ID">
          <el-input v-model="form.sms_config.template_id" placeholder="请输入模板 ID" />
        </el-form-item>
        <el-form-item label="测试手机号">
          <el-input v-model="testPhone" placeholder="用于测试发送" style="width: 200px" />
        </el-form-item>
      </el-card>

      <!-- 邮件配置 -->
      <el-card class="section-card">
        <template #header>
          <div style="display: flex; justify-content: space-between; align-items: center;">
            <span style="font-weight: 600;">邮件服务商配置</span>
            <el-button size="small" @click="testEmail" :loading="testingEmail">
              测试发送
            </el-button>
          </div>
        </template>
        <el-form-item label="邮件服务商">
          <el-select v-model="form.email_provider" placeholder="选择服务商">
            <el-option label="SMTP" value="smtp" />
            <el-option label="SendGrid" value="sendgrid" />
            <el-option label="阿里云邮件推送" value="aliyun" />
          </el-select>
        </el-form-item>
        <el-form-item label="SMTP 主机" v-if="form.email_provider === 'smtp'">
          <el-input v-model="form.email_config.host" placeholder="smtp.example.com" />
        </el-form-item>
        <el-form-item label="SMTP 端口" v-if="form.email_provider === 'smtp'">
          <el-input-number v-model="form.email_config.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="发件人邮箱">
          <el-input v-model="form.email_config.from_email" placeholder="noreply@example.com" />
        </el-form-item>
        <el-form-item label="发件人名称">
          <el-input v-model="form.email_config.from_name" placeholder="AI 智能体平台" />
        </el-form-item>
        <el-form-item label="用户名/API Key">
          <el-input v-model="form.email_config.username" placeholder="请输入用户名或 API Key" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input
            v-model="form.email_config.password"
            type="password"
            show-password
            placeholder="请输入密码"
          />
        </el-form-item>
        <el-form-item label="测试邮箱">
          <el-input v-model="testEmailAddr" placeholder="用于测试发送" style="width: 200px" />
        </el-form-item>
      </el-card>
    </el-form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getSettings, updateSettings, testSmsConfig, testEmailConfig, type PlatformSettings } from '@/api/settings'

const loading = ref(false)
const saving = ref(false)
const testingSms = ref(false)
const testingEmail = ref(false)
const testPhone = ref('')
const testEmailAddr = ref('')

const form = ref<PlatformSettings>({
  site_name: '',
  site_logo: '',
  contact_email: '',
  wechat_miniapp_id: '',
  wechat_miniapp_secret: '',
  sms_provider: 'aliyun',
  sms_config: {
    access_key: '',
    secret_key: '',
    sign_name: '',
    template_id: ''
  },
  email_provider: 'smtp',
  email_config: {
    host: '',
    port: 465,
    from_email: '',
    from_name: '',
    username: '',
    password: ''
  },
  max_upload_size: 10,
  allowed_file_types: ['pdf', 'doc', 'docx', 'txt', 'md'],
  maintenance_mode: false
})

async function loadSettings() {
  loading.value = true
  try {
    const res = await getSettings()
    form.value = { ...form.value, ...res }
  } catch {
    // ignore
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  saving.value = true
  try {
    await updateSettings(form.value)
    ElMessage.success('配置已保存')
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function testSms() {
  if (!testPhone.value) {
    ElMessage.warning('请输入测试手机号')
    return
  }
  testingSms.value = true
  try {
    const res = await testSmsConfig(testPhone.value)
    if (res.success) {
      ElMessage.success('测试短信已发送')
    } else {
      ElMessage.error(res.message || '发送失败')
    }
  } catch (e: any) {
    ElMessage.error(e.message || '发送失败')
  } finally {
    testingSms.value = false
  }
}

async function testEmail() {
  if (!testEmailAddr.value) {
    ElMessage.warning('请输入测试邮箱')
    return
  }
  testingEmail.value = true
  try {
    const res = await testEmailConfig(testEmailAddr.value)
    if (res.success) {
      ElMessage.success('测试邮件已发送')
    } else {
      ElMessage.error(res.message || '发送失败')
    }
  } catch (e: any) {
    ElMessage.error(e.message || '发送失败')
  } finally {
    testingEmail.value = false
  }
}

onMounted(() => loadSettings())
</script>

<style scoped lang="scss">
.section-card {
  margin-bottom: 20px;
}
</style>
