<template>
  <div class="agent-create">
    <div class="page-header">
      <h2>创建智能体</h2>
      <el-button @click="goBack">
        <el-icon><ArrowLeft /></el-icon>
        返回列表
      </el-button>
    </div>

    <el-card class="create-card">
      <el-form :model="formData" :rules="rules" ref="formRef" label-width="110px">
        <el-tabs v-model="configTab">
          <el-tab-pane label="基础配置" name="basic">
            <el-form-item label="名称" prop="name">
              <el-input v-model="formData.name" placeholder="请输入智能体名称" maxlength="50" show-word-limit />
            </el-form-item>

            <el-form-item label="描述" prop="description">
              <el-input
                v-model="formData.description"
                type="textarea"
                :rows="3"
                placeholder="描述智能体的用途和功能"
                maxlength="200"
                show-word-limit
              />
            </el-form-item>

            <el-form-item label="头像 URL">
              <el-input v-model="formData.avatar_url" placeholder="输入头像图片地址" />
            </el-form-item>

            <el-row :gutter="16">
              <el-col :span="12">
                <el-form-item label="类型" prop="agent_type">
                  <el-select v-model="formData.agent_type" placeholder="选择类型" style="width: 100%">
                    <el-option label="对话型" value="chat" />
                    <el-option label="任务型" value="task" />
                    <el-option label="混合型" value="hybrid" />
                  </el-select>
                </el-form-item>
              </el-col>
              <el-col :span="12">
                <el-form-item label="最大并发" prop="max_concurrent">
                  <el-input-number v-model="formData.max_concurrent" :min="1" :max="100" style="width: 100%" />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="欢迎语">
              <el-input
                v-model="formData.welcome_message"
                type="textarea"
                :rows="2"
                placeholder="用户首次对话时的欢迎语"
              />
            </el-form-item>
          </el-tab-pane>

          <el-tab-pane label="模型配置" name="model">
            <el-form-item label="AI 模型" prop="model_id">
              <el-select v-model="formData.model_id" placeholder="选择模型" style="width: 100%">
                <el-option label="通义千问 (默认)" value="default" />
                <el-option label="通义千问 Plus" value="qwen-plus" />
                <el-option label="通义千问 Turbo" value="qwen-turbo" />
                <el-option label="通义千问 Max" value="qwen-max" />
                <el-option label="DeepSeek Chat" value="deepseek-chat" />
                <el-option label="DeepSeek Coder" value="deepseek-coder" />
              </el-select>
            </el-form-item>

            <el-form-item label="性格风格">
              <el-select v-model="formData.personality" placeholder="选择性格风格" style="width: 100%">
                <el-option label="专业严谨" value="professional" />
                <el-option label="亲切友好" value="friendly" />
                <el-option label="活泼开朗" value="lively" />
                <el-option label="严谨学术" value="strict" />
                <el-option label="幽默风趣" value="humorous" />
              </el-select>
            </el-form-item>

            <el-form-item label="系统提示词">
              <el-input
                v-model="formData.system_prompt"
                type="textarea"
                :rows="6"
                placeholder="定义智能体的行为准则、角色定位和回复风格。例如：你是一个专业的客服助手，负责解答用户关于产品的问题..."
              />
              <div class="form-tip">系统提示词会作为每次对话的上下文发送给 AI 模型</div>
            </el-form-item>

            <el-form-item label="温度参数">
              <el-slider v-model="formData.temperature" :min="0" :max="1" :step="0.1" show-input />
              <div class="form-tip">值越低回复越确定性，值越高回复越创造性 (0-1)</div>
            </el-form-item>

            <el-form-item label="最大回复长度">
              <el-input-number v-model="formData.max_tokens" :min="100" :max="4096" :step="100" style="width: 100%" />
              <div class="form-tip">AI 单次回复的最大 token 数</div>
            </el-form-item>
          </el-tab-pane>

          <el-tab-pane label="知识库" name="knowledge">
            <el-form-item label="关联知识库">
              <el-select
                v-model="formData.knowledge_base_ids"
                multiple
                placeholder="选择关联的知识库"
                style="width: 100%"
                filterable
              >
                <el-option
                  v-for="kb in knowledgeBases"
                  :key="kb.id"
                  :label="kb.name"
                  :value="kb.id"
                >
                  <span>{{ kb.name }}</span>
                  <span style="color: #999; margin-left: 8px; font-size: 12px">
                    {{ kb.doc_count || 0 }} 篇文档
                  </span>
                </el-option>
              </el-select>
            </el-form-item>

            <el-form-item label="检索策略">
              <el-select v-model="formData.retrieval_strategy" style="width: 100%">
                <el-option label="相似度检索" value="similarity" />
                <el-option label="MMR 多样性" value="mmr" />
                <el-option label="混合检索" value="hybrid" />
              </el-select>
              <div class="form-tip">控制从知识库中检索文档的策略</div>
            </el-form-item>

            <el-form-item label="检索数量">
              <el-input-number v-model="formData.top_k" :min="1" :max="20" style="width: 100%" />
              <div class="form-tip">每次对话从知识库中检索的文档数量</div>
            </el-form-item>
          </el-tab-pane>

          <el-tab-pane label="工作时间" name="hours">
            <el-form-item label="启用工作时间">
              <el-switch v-model="formData.enable_work_hours" />
              <div class="form-tip">开启后，非工作时间将自动回复离线消息</div>
            </el-form-item>

            <template v-if="formData.enable_work_hours">
              <div v-for="day in weekDays" :key="day.value" class="work-hours-row">
                <span class="day-label">{{ day.label }}</span>
                <el-switch v-model="formData.work_hours![day.value].enabled" size="small" />
                <template v-if="formData.work_hours![day.value].enabled">
                  <el-time-picker
                    v-model="formData.work_hours![day.value].start"
                    format="HH:mm"
                    value-format="HH:mm"
                    placeholder="开始"
                    size="small"
                    style="width: 130px; margin-left: 8px"
                  />
                  <span style="margin: 0 4px">-</span>
                  <el-time-picker
                    v-model="formData.work_hours![day.value].end"
                    format="HH:mm"
                    value-format="HH:mm"
                    placeholder="结束"
                    size="small"
                    style="width: 130px"
                  />
                </template>
              </div>

              <el-form-item label="离线自动回复" style="margin-top: 16px">
                <el-input
                  v-model="formData.offline_message"
                  type="textarea"
                  :rows="2"
                  placeholder="非工作时间的自动回复内容"
                />
              </el-form-item>
            </template>
          </el-tab-pane>

          <el-tab-pane label="高级配置" name="advanced">
            <el-form-item label="启用记忆">
              <el-switch v-model="formData.enable_memory" />
              <div class="form-tip">开启后智能体可以记住用户的对话历史</div>
            </el-form-item>

            <el-form-item label="记忆轮数">
              <el-input-number v-model="formData.memory_turns" :min="0" :max="50" :disabled="!formData.enable_memory" style="width: 100%" />
              <div class="form-tip">保留最近多少轮对话作为上下文 (0 表示不限制)</div>
            </el-form-item>

            <el-form-item label="敏感词过滤">
              <el-switch v-model="formData.enable_sensitive_filter" />
              <div class="form-tip">开启后自动过滤敏感内容</div>
            </el-form-item>

            <el-form-item label="回复前缀">
              <el-input v-model="formData.reply_prefix" placeholder="每条回复前自动添加的前缀文本" />
            </el-form-item>

            <el-form-item label="渠道绑定">
              <el-select v-model="formData.bound_channels" multiple placeholder="选择绑定的渠道" style="width: 100%">
                <el-option label="微信" value="wechat" />
                <el-option label="网页" value="web" />
                <el-option label="API" value="api" />
              </el-select>
            </el-form-item>
          </el-tab-pane>
        </el-tabs>

        <div class="form-actions">
          <el-button @click="goBack">取消</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            创建智能体
          </el-button>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { ArrowLeft } from '@element-plus/icons-vue'
import { createAgent, type CreateAgentRequest } from '@/api/agent'
import { getKnowledgeBases, type KnowledgeBase } from '@/api/knowledge'

const router = useRouter()

const submitting = ref(false)
const formRef = ref()
const configTab = ref('basic')
const knowledgeBases = ref<KnowledgeBase[]>([])

const weekDays = [
  { label: '周一', value: 'monday' },
  { label: '周二', value: 'tuesday' },
  { label: '周三', value: 'wednesday' },
  { label: '周四', value: 'thursday' },
  { label: '周五', value: 'friday' },
  { label: '周六', value: 'saturday' },
  { label: '周日', value: 'sunday' }
]

const defaultWorkHours = () => {
  const hours: Record<string, { enabled: boolean; start: string; end: string }> = {}
  weekDays.forEach(d => {
    hours[d.value] = { enabled: d.value !== 'saturday' && d.value !== 'sunday', start: '09:00', end: '18:00' }
  })
  return hours
}

const formData = ref<CreateAgentRequest & {
  personality?: string
  system_prompt?: string
  temperature?: number
  max_tokens?: number
  retrieval_strategy?: string
  top_k?: number
  enable_work_hours?: boolean
  work_hours?: Record<string, { enabled: boolean; start: string; end: string }>
  offline_message?: string
  enable_memory?: boolean
  memory_turns?: number
  enable_sensitive_filter?: boolean
  reply_prefix?: string
  bound_channels?: string[]
}>({
  name: '',
  description: '',
  agent_type: 'chat',
  welcome_message: '',
  model_id: 'default',
  personality: 'professional',
  max_concurrent: 10,
  knowledge_base_ids: [],
  system_prompt: '',
  temperature: 0.7,
  max_tokens: 2048,
  retrieval_strategy: 'similarity',
  top_k: 5,
  enable_work_hours: false,
  work_hours: defaultWorkHours(),
  offline_message: '当前不在工作时间，我们将在工作时间内回复您。',
  enable_memory: true,
  memory_turns: 10,
  enable_sensitive_filter: true,
  reply_prefix: '',
  bound_channels: ['web', 'wechat']
})

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

const goBack = () => {
  router.push('/agents')
}

const loadKnowledgeBases = async () => {
  try {
    const res = await getKnowledgeBases(1, 100)
    knowledgeBases.value = res.items
  } catch { /* ignore */ }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate()

  submitting.value = true
  try {
    const data: CreateAgentRequest = {
      name: formData.value.name,
      description: formData.value.description,
      avatar_url: formData.value.avatar_url,
      agent_type: formData.value.agent_type,
      welcome_message: formData.value.welcome_message,
      model_id: formData.value.model_id,
      personality_config: {
        style: formData.value.personality,
        system_prompt: formData.value.system_prompt,
        temperature: formData.value.temperature,
        max_tokens: formData.value.max_tokens,
        retrieval_strategy: formData.value.retrieval_strategy,
        top_k: formData.value.top_k,
        enable_memory: formData.value.enable_memory,
        memory_turns: formData.value.memory_turns,
        enable_sensitive_filter: formData.value.enable_sensitive_filter,
        reply_prefix: formData.value.reply_prefix,
        bound_channels: formData.value.bound_channels
      },
      workflow_config: {
        enable_work_hours: formData.value.enable_work_hours,
        work_hours: formData.value.work_hours,
        offline_message: formData.value.offline_message
      },
      max_concurrent: formData.value.max_concurrent,
      knowledge_base_ids: formData.value.knowledge_base_ids
    }

    await createAgent(data)
    ElMessage.success('创建成功')
    router.push('/agents')
  } catch (e) {
    ElMessage.error('创建失败')
  } finally {
    submitting.value = false
  }
}

onMounted(() => {
  loadKnowledgeBases()
})
</script>

<style scoped lang="scss">
.agent-create {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.create-card {
  :deep(.el-card__body) {
    padding: 24px;
  }
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
  line-height: 1.4;
}

.work-hours-row {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid #f5f5f5;

  .day-label {
    width: 50px;
    font-size: 14px;
    color: #333;
  }
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #eee;
}
</style>
