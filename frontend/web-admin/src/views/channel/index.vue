<template>
  <div class="channel-list">
    <div class="page-header">
      <h2>渠道接入</h2>
    </div>

    <el-row :gutter="20">
      <el-col :span="8" v-for="channel in channels" :key="channel.type">
        <el-card class="channel-card" shadow="hover">
          <div class="channel-header">
            <el-avatar :size="48" :src="channel.icon" shape="square">
              {{ channel.name?.charAt(0) }}
            </el-avatar>
            <div class="channel-info">
              <h3>{{ channel.name }}</h3>
              <el-tag :type="channel.connected ? 'success' : 'info'" size="small">
                {{ channel.connected ? '已连接' : '未连接' }}
              </el-tag>
            </div>
          </div>
          <p class="channel-desc">{{ channel.description }}</p>
          <div class="channel-stats" v-if="channel.connected">
            <div class="stat-item">
              <span class="stat-value">{{ channel.active_sessions || 0 }}</span>
              <span class="stat-label">活跃会话</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ channel.today_messages || 0 }}</span>
              <span class="stat-label">今日消息</span>
            </div>
          </div>
          <div class="channel-actions">
            <el-button type="primary" size="small" @click="configureChannel(channel)">
              {{ channel.connected ? '配置' : '连接' }}
            </el-button>
            <el-button v-if="channel.connected" size="small" @click="testChannel(channel)">测试</el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 配置对话框 -->
    <el-dialog v-model="configVisible" :title="`配置${currentChannel?.name}`" width="500px">
      <el-form :model="configForm" label-width="100px" v-if="currentChannel">
        <template v-if="currentChannel.type === 'wechat'">
          <el-form-item label="AppID" required>
            <el-input v-model="configForm.appId" placeholder="微信公众号 AppID" />
          </el-form-item>
          <el-form-item label="AppSecret" required>
            <el-input v-model="configForm.appSecret" type="password" placeholder="微信公众号 AppSecret" show-password />
          </el-form-item>
          <el-form-item label="Token">
            <el-input v-model="configForm.token" placeholder="消息校验 Token" />
          </el-form-item>
          <el-form-item label="EncodingAESKey">
            <el-input v-model="configForm.aesKey" placeholder="消息加密密钥" />
          </el-form-item>
        </template>
        <template v-else-if="currentChannel.type === 'wecom'">
          <el-form-item label="CorpID" required>
            <el-input v-model="configForm.corpId" placeholder="企业 ID" />
          </el-form-item>
          <el-form-item label="AgentID" required>
            <el-input v-model="configForm.agentId" placeholder="应用 ID" />
          </el-form-item>
          <el-form-item label="Secret" required>
            <el-input v-model="configForm.secret" type="password" placeholder="应用密钥" show-password />
          </el-form-item>
        </template>
        <template v-else-if="currentChannel.type === 'feishu'">
          <el-form-item label="AppID" required>
            <el-input v-model="configForm.appId" placeholder="飞书应用 AppID" />
          </el-form-item>
          <el-form-item label="AppSecret" required>
            <el-input v-model="configForm.appSecret" type="password" placeholder="飞书应用 AppSecret" show-password />
          </el-form-item>
        </template>
        <template v-else-if="currentChannel.type === 'dingtalk'">
          <el-form-item label="AppKey" required>
            <el-input v-model="configForm.appKey" placeholder="钉钉应用 AppKey" />
          </el-form-item>
          <el-form-item label="AppSecret" required>
            <el-input v-model="configForm.appSecret" type="password" placeholder="钉钉应用 AppSecret" show-password />
          </el-form-item>
          <el-form-item label="AgentID" required>
            <el-input v-model="configForm.agentId" placeholder="钉钉应用 AgentID" />
          </el-form-item>
        </template>
        <template v-else-if="currentChannel.type === 'douyin'">
          <el-form-item label="AppID" required>
            <el-input v-model="configForm.appId" placeholder="抖音小程序 AppID" />
          </el-form-item>
          <el-form-item label="AppSecret" required>
            <el-input v-model="configForm.appSecret" type="password" placeholder="抖音小程序 AppSecret" show-password />
          </el-form-item>
        </template>
      </el-form>
      <template #footer>
        <el-button @click="configVisible = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { listChannels, createChannel, saveChannelConfig, testChannel as testChannelApi } from '@/api/channel'
import type { Channel as ApiChannel } from '@/api/channel'

interface Channel {
  id?: string
  type: string
  name: string
  description: string
  icon?: string
  connected: boolean
  active_sessions?: number
  today_messages?: number
}

const channelTemplates: Channel[] = [
  { type: 'wechat', name: '微信公众号', description: '接入微信公众号，自动回复用户消息', connected: false },
  { type: 'wecom', name: '企业微信', description: '接入企业微信，实现内部客服自动化', connected: false },
  { type: 'feishu', name: '飞书', description: '接入飞书机器人，自动回复消息', connected: false },
  { type: 'dingtalk', name: '钉钉', description: '接入钉钉机器人，自动回复消息', connected: false },
  { type: 'douyin', name: '抖音', description: '接入抖音私信，自动回复用户咨询', connected: false }
]

const channels = ref<Channel[]>(channelTemplates)

const configVisible = ref(false)
const currentChannel = ref<Channel | null>(null)
const configForm = ref({
  appId: '',
  appSecret: '',
  appKey: '',
  token: '',
  aesKey: '',
  corpId: '',
  agentId: '',
  secret: ''
})

const configureChannel = (channel: Channel) => {
  currentChannel.value = channel
  configForm.value = {
    appId: '',
    appSecret: '',
    appKey: '',
    token: '',
    aesKey: '',
    corpId: '',
    agentId: '',
    secret: ''
  }
  configVisible.value = true
}

const saveConfig = async () => {
  if (!currentChannel.value) return
  try {
    if (!currentChannel.value.id) {
      const channel = await createChannel(currentChannel.value.type, currentChannel.value.name, configForm.value as any)
      const idx = channels.value.findIndex(c => c.type === currentChannel.value?.type)
      if (idx !== -1) {
        channels.value[idx].id = channel.id
        channels.value[idx].connected = true
      }
    } else {
      await saveChannelConfig(currentChannel.value.id, configForm.value as any)
    }
    ElMessage.success('配置已保存')
    configVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  }
}

const testChannel = async (channel: Channel) => {
  if (!channel.id) {
    ElMessage.info('请先保存配置')
    return
  }
  try {
    const result = await testChannelApi(channel.id)
    if (result.success) {
      ElMessage.success('连接测试通过')
    }
  } catch (e: any) {
    ElMessage.error(e.message || '测试失败')
  }
}

onMounted(async () => {
  try {
    const data = await listChannels()
    const apiChannels = data.items || []
    channels.value = channelTemplates.map(template => {
      const found = apiChannels.find((c: ApiChannel) => c.type === template.type)
      return {
        ...template,
        id: found?.id,
        connected: found?.status === 'active'
      }
    })
  } catch {
    // use defaults
  }
})
</script>

<style scoped lang="scss">
.channel-list {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.channel-card {
  margin-bottom: 20px;

  .channel-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;

    h3 {
      margin: 0 0 4px 0;
      font-size: 16px;
    }
  }

  .channel-desc {
    color: #666;
    font-size: 14px;
    margin-bottom: 16px;
    min-height: 40px;
  }

  .channel-stats {
    display: flex;
    justify-content: space-around;
    padding: 12px 0;
    border-top: 1px solid #eee;
    border-bottom: 1px solid #eee;
    margin-bottom: 12px;

    .stat-item {
      text-align: center;

      .stat-value {
        display: block;
        font-size: 18px;
        font-weight: 600;
        color: #333;
      }

      .stat-label {
        font-size: 12px;
        color: #999;
      }
    }
  }

  .channel-actions {
    display: flex;
    gap: 8px;
  }
}
</style>
