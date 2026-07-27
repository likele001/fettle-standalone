<template>
  <div class="agent-list">
    <el-tabs v-model="activeTab" class="page-tabs">
      <el-tab-pane label="智能体列表" name="agents">
        <div class="page-header">
          <h2>智能体管理</h2>
          <el-button type="primary" @click="showCreateDialog">
            <el-icon><Plus /></el-icon>
            创建智能体
          </el-button>
        </div>

    <el-card class="filter-card" shadow="never">
      <el-row :gutter="16" align="middle">
        <el-col :span="6">
          <el-input v-model="filters.keyword" placeholder="搜索智能体名称" clearable @clear="loadAgents">
            <template #prefix><el-icon><Search /></el-icon></template>
          </el-input>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.agent_type" placeholder="类型" clearable @change="loadAgents">
            <el-option label="对话型" value="chat" />
            <el-option label="任务型" value="task" />
            <el-option label="混合型" value="hybrid" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.status" placeholder="状态" clearable @change="loadAgents">
            <el-option label="空闲" value="idle" />
            <el-option label="处理中" value="processing" />
            <el-option label="异常" value="error" />
          </el-select>
        </el-col>
        <el-col :span="4">
          <el-select v-model="filters.model_id" placeholder="模型" clearable @change="loadAgents">
            <el-option label="通义千问" value="default" />
            <el-option label="通义千问 Plus" value="qwen-plus" />
            <el-option label="通义千问 Turbo" value="qwen-turbo" />
            <el-option label="DeepSeek" value="deepseek-chat" />
          </el-select>
        </el-col>
        <el-col :span="6" style="text-align: right">
          <el-button @click="resetFilters">重置</el-button>
        </el-col>
      </el-row>
    </el-card>

    <div v-if="agents.length === 0" class="empty-state">
      <div class="empty-icon">
        <el-icon :size="64" color="#409EFF"><Service /></el-icon>
      </div>
      <h3>还没有智能体</h3>
      <p>选择一个模板，几分钟内创建您的第一个智能体</p>
      <div class="quick-templates">
        <div
          v-for="template in quickTemplates"
          :key="template.id"
          class="quick-template-item"
          @click="createFromTemplate(template)"
        >
          <div class="quick-icon" :style="{ background: template.color + '15', color: template.color }">
            <el-icon :size="24"><component :is="getIcon(template.icon)" /></el-icon>
          </div>
          <span>{{ template.name }}</span>
        </div>
      </div>
      <el-button type="primary" size="large" @click="showCreateDialog" style="margin-top: 16px">
        <el-icon><Plus /></el-icon>
        立即创建
      </el-button>
    </div>

    <el-row v-else :gutter="20" style="margin-top: 20px">
      <el-col :span="8" v-for="agent in agents" :key="agent.id">
        <el-card class="agent-card" shadow="hover">
          <div class="agent-header">
            <el-avatar :size="48" :src="agent.avatar_url">
              {{ agent.name?.charAt(0) }}
            </el-avatar>
            <div class="agent-info">
              <h3>{{ agent.name }}</h3>
              <div class="agent-tags">
                <el-tag :type="statusType(agent.status)" size="small">
                  {{ statusText(agent.status) }}
                </el-tag>
                <el-tag size="small" type="info">{{ typeText(agent.agent_type) }}</el-tag>
              </div>
            </div>
          </div>

          <p class="agent-desc">{{ agent.description || '暂无描述' }}</p>

          <div class="agent-meta">
            <div class="meta-item">
              <el-icon><Cpu /></el-icon>
              <span>{{ modelText(agent.model_id) }}</span>
            </div>
            <div class="meta-item">
              <el-icon><ChatDotRound /></el-icon>
              <span>{{ agent.total_conversations || 0 }} 会话</span>
            </div>
            <div class="meta-item">
              <el-icon><Message /></el-icon>
              <span>{{ agent.total_messages || 0 }} 消息</span>
            </div>
          </div>

          <div class="agent-stats">
            <div class="stat-item">
              <span class="stat-value">{{ agent.current_sessions || 0 }}/{{ agent.max_concurrent }}</span>
              <span class="stat-label">当前并发</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ agent.knowledge_base_ids?.length || 0 }}</span>
              <span class="stat-label">知识库</span>
            </div>
          </div>

          <div class="agent-actions">
            <el-button size="small" @click="editAgent(agent)">
              <el-icon><Edit /></el-icon> 编辑
            </el-button>
            <el-button size="small" @click="testAgent(agent)">
              <el-icon><ChatLineRound /></el-icon> 测试
            </el-button>
            <el-popconfirm title="确定删除此智能体？" @confirm="handleDelete(agent.id)">
              <template #reference>
                <el-button size="small" type="danger">
                  <el-icon><Delete /></el-icon> 删除
                </el-button>
              </template>
            </el-popconfirm>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-pagination
      v-if="total > pageSize"
      class="pagination"
      layout="prev, pager, next, total"
      :total="total"
      :page-size="pageSize"
      v-model:current-page="currentPage"
      @current-change="loadAgents"
    />

    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="760px"
      :close-on-click-modal="false"
    >
      <div class="step-indicator">
        <div :class="['step', { active: currentStep >= 1, done: currentStep > 1 }]">
          <span class="step-num">1</span>
          <span class="step-text">选择模板</span>
        </div>
        <div class="step-line"></div>
        <div :class="['step', { active: currentStep >= 2, done: currentStep > 2 }]">
          <span class="step-num">2</span>
          <span class="step-text">基本设置</span>
        </div>
      </div>

      <div v-if="currentStep === 1" class="step-content">
        <div style="margin-bottom: 16px">
          <h3 style="margin: 0 0 8px 0">选择一个模板开始</h3>
          <p style="margin: 0; color: #999; font-size: 13px">选择预置模板可快速创建智能体，也可以创建空白智能体自定义配置</p>
        </div>
        <AgentTemplateSelector @select="handleTemplateSelect" @clear="handleCreateBlank" />
      </div>

      <div v-else class="step-content">
        <el-form :model="formData" :rules="rules" ref="formRef" label-width="110px">
          <el-form-item label="名称" prop="name">
            <el-input v-model="formData.name" placeholder="请输入智能体名称" maxlength="50" show-word-limit />
          </el-form-item>

          <el-form-item label="描述" prop="description">
            <el-input
              v-model="formData.description"
              type="textarea"
              :rows="2"
              placeholder="描述智能体的用途和功能"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>

          <el-row :gutter="16">
            <el-col :span="12">
              <el-form-item label="AI 模型" prop="model_id">
                <el-select v-model="formData.model_id" placeholder="选择模型" style="width: 100%">
                  <el-option label="通义千问 (推荐)" value="default" />
                  <el-option label="通义千问 Plus" value="qwen-plus" />
                  <el-option label="通义千问 Turbo" value="qwen-turbo" />
                  <el-option label="DeepSeek Chat" value="deepseek-chat" />
                  <el-option label="DeepSeek Coder" value="deepseek-coder" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="行业分类" prop="industry">
                <el-select v-model="formData.industry" placeholder="选择行业" style="width: 100%">
                  <el-option
                    v-for="(label, value) in availableIndustries"
                    :key="value"
                    :label="label"
                    :value="value"
                  />
                </el-select>
                <span v-if="currentPlan?.allowed_industries?.length" style="font-size: 12px; color: #94a3b8; display: block; margin-top: 4px">
                  当前套餐限制：仅可选择 {{ currentPlan.allowed_industries.map(i => industryLabels[i] || i).join('、') }}
                </span>
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

          <el-form-item label="关联知识库">
            <el-select
              v-model="formData.knowledge_base_ids"
              multiple
              placeholder="选择关联的知识库（可选）"
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

          <el-divider />

          <el-collapse v-model="expandAdvanced" accordion>
            <el-collapse-item title="高级配置（可选）" name="advanced">
              <el-row :gutter="16">
                <el-col :span="12">
                  <el-form-item label="性格风格">
                    <el-select v-model="formData.personality" placeholder="选择性格风格" style="width: 100%">
                      <el-option label="专业严谨" value="professional" />
                      <el-option label="亲切友好" value="friendly" />
                      <el-option label="活泼开朗" value="lively" />
                      <el-option label="严谨学术" value="strict" />
                      <el-option label="幽默风趣" value="humorous" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="12">
                  <el-form-item label="温度参数">
                    <el-slider v-model="formData.temperature" :min="0" :max="1" :step="0.1" show-input />
                    <div class="form-tip">值越低回复越确定性，值越高回复越创造性</div>
                  </el-form-item>
                </el-col>
              </el-row>

              <el-form-item label="系统提示词">
                <el-input
                  v-model="formData.system_prompt"
                  type="textarea"
                  :rows="4"
                  placeholder="定义智能体的行为准则、角色定位和回复风格..."
                />
                <div class="form-tip">系统提示词会作为每次对话的上下文发送给 AI 模型</div>
              </el-form-item>

              <el-row :gutter="16">
                <el-col :span="8">
                  <el-form-item label="启用记忆">
                    <el-switch v-model="formData.enable_memory" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="记忆轮数">
                    <el-input-number v-model="formData.memory_turns" :min="0" :max="50" :disabled="!formData.enable_memory" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="敏感词过滤">
                    <el-switch v-model="formData.enable_sensitive_filter" />
                  </el-form-item>
                </el-col>
              </el-row>
            </el-collapse-item>
          </el-collapse>
        </el-form>
      </div>

      <template #footer>
        <el-button v-if="currentStep === 2" @click="currentStep = 1">上一步</el-button>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button v-if="currentStep === 1" type="primary" @click="showBlankForm">
          创建空白智能体
        </el-button>
        <el-button v-else type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
      </el-tab-pane>

      <el-tab-pane name="bundles">
        <template #label>
          <span><el-icon><Grid /></el-icon> 行业套餐</span>
        </template>
        <div class="bundles-section">
          <div class="bundles-header">
            <h2>行业套餐</h2>
            <p class="bundles-desc">选择适合您行业的预设套餐，一键安装智能体和知识库模板</p>
          </div>

          <div v-loading="bundlesLoading">
            <el-row :gutter="20" v-if="bundles.length > 0">
              <el-col :span="8" v-for="bundle in bundles" :key="bundle.industry">
                <el-card class="bundle-card" shadow="hover">
                  <div class="bundle-icon-wrap" :style="{ background: getBundleColor(bundle.industry) + '15', color: getBundleColor(bundle.industry) }">
                    <el-icon :size="32"><component :is="getBundleIcon(bundle.industry)" /></el-icon>
                  </div>
                  <h3 class="bundle-name">{{ bundle.name }}</h3>
                  <p class="bundle-desc">{{ bundle.description }}</p>

                  <div class="bundle-contents">
                    <div class="bundle-sub-section">
                      <h4>包含智能体 ({{ bundle.config?.agents?.length || 0 }})</h4>
                      <ul>
                        <li v-for="agent in bundle.config?.agents" :key="agent.name">{{ agent.name }}</li>
                      </ul>
                    </div>
                    <div class="bundle-sub-section" v-if="bundle.config?.kb_templates?.length">
                      <h4>知识库模板 ({{ bundle.config?.kb_templates?.length }})</h4>
                      <ul>
                        <li v-for="kb in bundle.config?.kb_templates" :key="kb.name">{{ kb.name }}</li>
                      </ul>
                    </div>
                  </div>

                  <el-button
                    type="primary"
                    class="bundle-install-btn"
                    @click="handleInstallBundle(bundle)"
                    :loading="installingBundle === bundle.industry"
                  >
                    <el-icon><Download /></el-icon> 一键安装
                  </el-button>
                </el-card>
              </el-col>
            </el-row>

            <el-empty v-else-if="!bundlesLoading" description="暂无可用行业套餐" />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 安装成功对话框 -->
    <el-dialog v-model="installResultVisible" title="安装完成" width="500px">
      <div class="install-result">
        <el-result icon="success" title="行业套餐安装成功" v-if="installResult.created_agents?.length || installResult.created_kbs?.length">
          <template #sub-title>
            <div class="install-detail">
              <div v-if="installResult.created_agents?.length">
                <h4>已创建智能体：</h4>
                <el-tag v-for="a in installResult.created_agents" :key="a.id" type="success" size="small" style="margin: 2px 4px">{{ a.name }}</el-tag>
              </div>
              <div v-if="installResult.created_kbs?.length" style="margin-top: 12px">
                <h4>已创建知识库：</h4>
                <el-tag v-for="kb in installResult.created_kbs" :key="kb.id" type="info" size="small" style="margin: 2px 4px">{{ kb.name }}</el-tag>
              </div>
              <div v-if="installResult.skipped?.length" style="margin-top: 12px">
                <h4>已跳过（已存在）：</h4>
                <el-tag v-for="s in installResult.skipped" :key="s" type="warning" size="small" style="margin: 2px 4px">{{ s }}</el-tag>
              </div>
            </div>
          </template>
        </el-result>
        <el-result icon="warning" title="部分项目已跳过" v-else>
          <template #sub-title>
            <p>以下内容已存在，已跳过安装：</p>
            <el-tag v-for="s in installResult.skipped" :key="s" type="warning" size="small" style="margin: 2px 4px">{{ s }}</el-tag>
          </template>
        </el-result>
      </div>
      <template #footer>
        <el-button type="primary" @click="installResultVisible = false">确定</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="testDialogVisible" title="测试对话" width="560px">
      <div class="test-chat">
        <div class="chat-header">
          <el-avatar :size="32" :src="currentAgent?.avatar_url">
            {{ currentAgent?.name?.charAt(0) }}
          </el-avatar>
          <span class="chat-agent-name">{{ currentAgent?.name }}</span>
          <el-tag :type="statusType(currentAgent?.status || '')" size="small" style="margin-left: 8px">
            {{ statusText(currentAgent?.status || '') }}
          </el-tag>
        </div>
        <div class="chat-messages" ref="chatMessagesRef">
          <div
            v-for="(msg, index) in testMessages"
            :key="index"
            :class="['message', msg.role]"
          >
            <div class="message-content">{{ msg.content }}</div>
          </div>
          <div v-if="sending" class="message assistant">
            <div class="message-content typing">
              <span class="dot"></span><span class="dot"></span><span class="dot"></span>
            </div>
          </div>
        </div>
        <div class="chat-input">
          <el-input
            v-model="testInput"
            placeholder="输入消息..."
            @keyup.enter="sendTestMessage"
            :disabled="sending"
          >
            <template #append>
              <el-button @click="sendTestMessage" :loading="sending">发送</el-button>
            </template>
          </el-input>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Plus, Search, Edit, Delete, ChatLineRound,
  Cpu, ChatDotRound, Message, Service,
  Headset, EditPen, DataAnalysis, Document,
  TrendCharts, User, Download, Grid,
  ShoppingCart, ForkSpoon, OfficeBuilding, School, FirstAidKit
} from '@element-plus/icons-vue'
import {
  getAgents,
  createAgent,
  updateAgent,
  deleteAgent,
  testChat,
  type Agent,
  type CreateAgentRequest,
  type UpdateAgentRequest
} from '@/api/agent'
import { getKnowledgeBases, type KnowledgeBase } from '@/api/knowledge'
import { getBundles, applyBundle, type IndustryBundle, type ApplyBundleResult } from '@/api/bundle'
import { getSubscription, type Plan } from '@/api/billing'
import AgentTemplateSelector from '@/components/AgentTemplateSelector.vue'
import { agentTemplates, type AgentTemplate } from '@/data/agentTemplates'

const agents = ref<Agent[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(12)

const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<string | null>(null)
const submitting = ref(false)
const formRef = ref()
const currentStep = ref(1)
const expandAdvanced = ref<string[]>([])

const knowledgeBases = ref<KnowledgeBase[]>([])

const quickTemplates = computed(() => agentTemplates.slice(0, 4))

const iconMap: Record<string, any> = {
  'headset': Headset,
  'code': EditPen,
  'pen-tool': EditPen,
  'bar-chart': DataAnalysis,
  'globe': Service,
  'document': Document,
  'trending-up': TrendCharts,
  'users': User
}

const getIcon = (icon: string) => iconMap[icon] || Headset

const industryLabels: Record<string, string> = {
  ecommerce: '电商',
  dining: '餐饮',
  legal: '法律',
  education: '教育',
  medical: '医疗',
  finance: '金融',
  logistics: '物流',
  real_estate: '房产',
  retail: '零售',
  entertainment: '娱乐'
}

const allIndustries: Record<string, string> = {
  ecommerce: '电商',
  dining: '餐饮',
  legal: '法律',
  education: '教育',
  medical: '医疗',
  finance: '金融',
  logistics: '物流',
  real_estate: '房产',
  retail: '零售',
  entertainment: '娱乐'
}

const currentPlan = ref<Plan | null>(null)

const availableIndustries = computed(() => {
  if (!currentPlan.value || !currentPlan.value.allowed_industries.length) {
    return allIndustries
  }
  const allowed: Record<string, string> = {}
  for (const industry of currentPlan.value.allowed_industries) {
    if (allIndustries[industry]) {
      allowed[industry] = allIndustries[industry]
    }
  }
  return allowed
})

const dialogTitle = computed(() => {
  if (isEdit.value) return '编辑智能体'
  if (currentStep.value === 1) return '创建智能体 - 选择模板'
  return '创建智能体 - 基本设置'
})

interface FormDataType {
  name: string
  description: string
  avatar_url: string
  agent_type: string
  welcome_message: string
  model_id: string
  industry: string
  personality: string
  max_concurrent: number
  knowledge_base_ids: string[]
  system_prompt: string
  temperature: number
  max_tokens: number
  enable_memory: boolean
  memory_turns: number
  enable_sensitive_filter: boolean
}

const defaultFormData = (): FormDataType => ({
  name: '',
  description: '',
  avatar_url: '',
  agent_type: 'chat',
  welcome_message: '',
  model_id: 'default',
  industry: '',
  personality: 'professional',
  max_concurrent: 10,
  knowledge_base_ids: [],
  system_prompt: '',
  temperature: 0.7,
  max_tokens: 2048,
  enable_memory: true,
  memory_turns: 10,
  enable_sensitive_filter: true
})

const formData = ref<FormDataType>({ ...defaultFormData() })

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

const filters = ref({
  keyword: '',
  agent_type: '',
  status: '',
  model_id: ''
})

const testDialogVisible = ref(false)
const testAgentId = ref<string | null>(null)
const currentAgent = ref<Agent | null>(null)
const testMessages = ref<{ role: string; content: string }[]>([])
const testInput = ref('')
const sending = ref(false)
const chatMessagesRef = ref<HTMLElement>()

const statusType = (status: string) => {
  const map: Record<string, string> = {
    idle: 'success', processing: 'warning', waiting: 'info', error: 'danger'
  }
  return map[status] || 'info'
}

const statusText = (status: string) => {
  const map: Record<string, string> = {
    idle: '空闲', processing: '处理中', waiting: '等待', error: '异常'
  }
  return map[status] || status
}

const typeText = (type: string) => {
  const map: Record<string, string> = { chat: '对话型', task: '任务型', hybrid: '混合型' }
  return map[type] || type
}

const modelText = (modelId: string) => {
  const map: Record<string, string> = {
    default: '通义千问', 'qwen-plus': 'Qwen Plus', 'qwen-turbo': 'Qwen Turbo',
    'qwen-max': 'Qwen Max', 'deepseek-chat': 'DeepSeek', 'deepseek-coder': 'DeepSeek Coder'
  }
  return map[modelId] || modelId
}

const resetFilters = () => {
  filters.value = { keyword: '', agent_type: '', status: '', model_id: '' }
  loadAgents()
}

const loadAgents = async () => {
  try {
    const filterParams: Record<string, any> = {}
    if (filters.value.keyword) filterParams.keyword = filters.value.keyword
    if (filters.value.agent_type) filterParams.agent_type = filters.value.agent_type
    if (filters.value.status) filterParams.status = filters.value.status
    if (filters.value.model_id) filterParams.model_id = filters.value.model_id

    const res = await getAgents(currentPage.value, pageSize.value, filterParams)
    agents.value = res.items
    total.value = res.total
  } catch (e) {
    ElMessage.error('加载智能体列表失败')
  }
}

const loadKnowledgeBases = async () => {
  try {
    const res = await getKnowledgeBases(1, 100)
    knowledgeBases.value = res.items
  } catch (e) {}
}

const showCreateDialog = () => {
  isEdit.value = false
  editingId.value = null
  currentStep.value = 1
  formData.value = { ...defaultFormData() }
  dialogVisible.value = true
}

const showBlankForm = () => {
  currentStep.value = 2
}

const handleTemplateSelect = (template: AgentTemplate) => {
  formData.value = {
    name: template.name,
    description: template.description,
    avatar_url: '',
    agent_type: template.agent_type,
    welcome_message: template.welcome_message,
    model_id: template.model_id,
    industry: '',
    personality: template.personality_config.style,
    max_concurrent: 10,
    knowledge_base_ids: [],
    system_prompt: template.personality_config.system_prompt,
    temperature: template.personality_config.temperature,
    max_tokens: template.personality_config.max_tokens,
    enable_memory: template.personality_config.enable_memory,
    memory_turns: template.personality_config.memory_turns,
    enable_sensitive_filter: template.personality_config.enable_sensitive_filter
  }
  currentStep.value = 2
}

const handleCreateBlank = () => {
  formData.value = { ...defaultFormData() }
  currentStep.value = 2
}

const createFromTemplate = (template: AgentTemplate) => {
  showCreateDialog()
  setTimeout(() => {
    handleTemplateSelect(template)
  }, 300)
}

const editAgent = (agent: Agent) => {
  isEdit.value = true
  editingId.value = agent.id
  currentStep.value = 2
  formData.value = {
    name: agent.name,
    description: agent.description,
    avatar_url: agent.avatar_url,
    agent_type: agent.agent_type,
    welcome_message: agent.welcome_message,
    model_id: agent.model_id,
    industry: (agent as any).industry || '',
    personality: agent.personality_config?.style || 'professional',
    max_concurrent: agent.max_concurrent,
    knowledge_base_ids: agent.knowledge_base_ids || [],
    system_prompt: agent.personality_config?.system_prompt || '',
    temperature: agent.personality_config?.temperature ?? 0.7,
    max_tokens: agent.personality_config?.max_tokens ?? 2048,
    enable_memory: agent.personality_config?.enable_memory ?? true,
    memory_turns: agent.personality_config?.memory_turns ?? 10,
    enable_sensitive_filter: agent.personality_config?.enable_sensitive_filter ?? true
  }
  dialogVisible.value = true
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
        enable_memory: formData.value.enable_memory,
        memory_turns: formData.value.memory_turns,
        enable_sensitive_filter: formData.value.enable_sensitive_filter
      },
      max_concurrent: formData.value.max_concurrent,
      knowledge_base_ids: formData.value.knowledge_base_ids
    }

    if (isEdit.value && editingId.value) {
      await updateAgent(editingId.value, data as UpdateAgentRequest)
      ElMessage.success('更新成功')
    } else {
      await createAgent(data)
      ElMessage.success('创建成功')
    }

    dialogVisible.value = false
    loadAgents()
  } catch (e) {
    ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id: string) => {
  try {
    await deleteAgent(id)
    ElMessage.success('删除成功')
    loadAgents()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

const testAgent = (agent: Agent) => {
  testAgentId.value = agent.id
  currentAgent.value = agent
  testMessages.value = [
    { role: 'assistant', content: agent.welcome_message || `你好，我是${agent.name}，有什么可以帮你的？` }
  ]
  testDialogVisible.value = true
}

const sendTestMessage = async () => {
  if (!testInput.value.trim() || !testAgentId.value) return

  const userMsg = testInput.value.trim()
  testMessages.value.push({ role: 'user', content: userMsg })
  testInput.value = ''
  sending.value = true

  await nextTick()
  scrollToBottom()

  try {
    const res = await testChat(testAgentId.value, userMsg)
    testMessages.value.push({ role: 'assistant', content: res.reply || '抱歉，我暂时无法回答。' })
  } catch (e) {
    testMessages.value.push({ role: 'assistant', content: '测试请求失败，请检查服务是否正常。' })
  } finally {
    sending.value = false
    await nextTick()
    scrollToBottom()
  }
}

const scrollToBottom = () => {
  if (chatMessagesRef.value) {
    chatMessagesRef.value.scrollTop = chatMessagesRef.value.scrollHeight
  }
}

// 行业套餐相关
const activeTab = ref('agents')
const bundles = ref<IndustryBundle[]>([])
const bundlesLoading = ref(false)
const installingBundle = ref<string | null>(null)
const installResultVisible = ref(false)
const installResult = ref<ApplyBundleResult>({ created_agents: [], created_kbs: [], skipped: [] })

const bundleIconMap: Record<string, any> = {
  ecommerce: ShoppingCart,
  dining: ForkSpoon,
  legal: OfficeBuilding,
  education: School,
  medical: FirstAidKit
}

const bundleColorMap: Record<string, string> = {
  ecommerce: '#FF6B35',
  dining: '#E8432A',
  legal: '#2C5F8A',
  education: '#28A745',
  medical: '#DC3545'
}

const getBundleIcon = (industry: string) => bundleIconMap[industry] || Grid
const getBundleColor = (industry: string) => bundleColorMap[industry] || '#409EFF'

const loadBundles = async () => {
  bundlesLoading.value = true
  try {
    const res = await getBundles()
    bundles.value = Array.isArray(res) ? res : (res as any)?.items || []
  } catch (e) {
    bundles.value = []
  } finally {
    bundlesLoading.value = false
  }
}

const handleInstallBundle = async (bundle: IndustryBundle) => {
  installingBundle.value = bundle.industry
  try {
    const res = await applyBundle(bundle.industry)
    installResult.value = (res as any)?.data || res
    installResultVisible.value = true
    loadAgents()
    loadKnowledgeBases()
  } catch (e: any) {
    ElMessage.error(e.message || '安装失败')
  } finally {
    installingBundle.value = null
  }
}

import { watch } from 'vue'
watch(activeTab, (val) => {
  if (val === 'bundles' && bundles.value.length === 0) {
    loadBundles()
  }
})

const loadCurrentPlan = async () => {
  try {
    const res = await getSubscription()
    const data = (res as any)?.data || res
    currentPlan.value = data?.plan || null
  } catch {
    currentPlan.value = null
  }
}

onMounted(() => {
  loadAgents()
  loadKnowledgeBases()
  loadCurrentPlan()
})
</script>

<style scoped lang="scss">
.agent-list {
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

.filter-card {
  :deep(.el-card__body) {
    padding: 16px 20px;
  }
}

.empty-state {
  text-align: center;
  padding: 60px 20px;

  .empty-icon {
    margin-bottom: 16px;
  }

  h3 {
    margin: 0 0 8px 0;
    font-size: 20px;
    color: #333;
  }

  p {
    margin: 0 0 24px 0;
    color: #999;
    font-size: 14px;
  }

  .quick-templates {
    display: flex;
    justify-content: center;
    gap: 16px;
    margin-bottom: 24px;

    .quick-template-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 8px;
      padding: 16px 24px;
      border: 1px solid #e8e8e8;
      border-radius: 8px;
      cursor: pointer;
      transition: all 0.2s;
      background: #fff;

      &:hover {
        border-color: #409EFF;
        background: #f0f5ff;
      }

      .quick-icon {
        width: 48px;
        height: 48px;
        border-radius: 10px;
        display: flex;
        align-items: center;
        justify-content: center;
      }

      span {
        font-size: 13px;
        color: #666;
      }
    }
  }
}

.agent-card {
  margin-bottom: 20px;

  .agent-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;

    h3 {
      margin: 0 0 4px 0;
      font-size: 16px;
    }

    .agent-tags {
      display: flex;
      gap: 4px;
      margin-top: 4px;
    }
  }

  .agent-desc {
    color: #666;
    font-size: 14px;
    margin-bottom: 12px;
    min-height: 36px;
    line-height: 1.5;
  }

  .agent-meta {
    display: flex;
    gap: 16px;
    margin-bottom: 12px;
    flex-wrap: wrap;

    .meta-item {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 13px;
      color: #888;
    }
  }

  .agent-stats {
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

  .agent-actions {
    display: flex;
    gap: 8px;
  }
}

.pagination {
  margin-top: 20px;
  text-align: center;
}

.step-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;

  .step {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;

    .step-num {
      width: 32px;
      height: 32px;
      border-radius: 50%;
      background: #e8e8e8;
      color: #999;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 14px;
      font-weight: 600;
      transition: all 0.2s;
    }

    .step-text {
      font-size: 13px;
      color: #999;
    }

    &.active {
      .step-num {
        background: #409EFF;
        color: white;
      }
      .step-text {
        color: #409EFF;
      }
    }

    &.done {
      .step-num {
        background: #67C23A;
        color: white;
      }
      .step-text {
        color: #67C23A;
      }
    }
  }

  .step-line {
    width: 120px;
    height: 2px;
    background: #e8e8e8;
    margin: 0 16px;

    .step.active + & {
      background: #409EFF;
    }

    .step.done + & {
      background: #67C23A;
    }
  }
}

.step-content {
  max-height: 500px;
  overflow-y: auto;
}

.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.test-chat {
  .chat-header {
    display: flex;
    align-items: center;
    padding-bottom: 12px;
    border-bottom: 1px solid #eee;
    margin-bottom: 12px;

    .chat-agent-name {
      margin-left: 8px;
      font-weight: 600;
      font-size: 15px;
    }
  }

  .chat-messages {
    height: 360px;
    overflow-y: auto;
    padding: 12px;
    background: #f5f5f5;
    border-radius: 8px;
    margin-bottom: 12px;

    .message {
      margin-bottom: 12px;

      &.user {
        text-align: right;

        .message-content {
          background: #409EFF;
          color: white;
          display: inline-block;
          padding: 8px 12px;
          border-radius: 12px 12px 0 12px;
          max-width: 80%;
        }
      }

      &.assistant {
        .message-content {
          background: white;
          display: inline-block;
          padding: 8px 12px;
          border-radius: 12px 12px 12px 0;
          max-width: 80%;

          &.typing {
            display: flex;
            gap: 4px;
            padding: 12px 16px;

            .dot {
              width: 6px;
              height: 6px;
              background: #999;
              border-radius: 50%;
              animation: bounce 1.4s infinite ease-in-out;

              &:nth-child(2) { animation-delay: 0.2s }
              &:nth-child(3) { animation-delay: 0.4s }
            }
          }
        }
      }
    }
  }
}

@keyframes bounce {
  0%, 80%, 100% { transform: scale(0) }
  40% { transform: scale(1) }
}

.page-tabs {
  :deep(.el-tabs__header) {
    margin-bottom: 20px;
  }
}

.bundles-section {
  .bundles-header {
    margin-bottom: 24px;

    h2 {
      margin: 0 0 8px 0;
      font-size: 20px;
    }

    .bundles-desc {
      margin: 0;
      color: #999;
      font-size: 14px;
    }
  }
}

.bundle-card {
  text-align: center;
  margin-bottom: 20px;
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-4px);
  }

  .bundle-icon-wrap {
    width: 64px;
    height: 64px;
    border-radius: 16px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin: 0 auto 16px;
  }

  .bundle-name {
    margin: 0 0 8px 0;
    font-size: 18px;
  }

  .bundle-desc {
    color: #666;
    font-size: 14px;
    margin-bottom: 16px;
    min-height: 42px;
    line-height: 1.5;
  }

  .bundle-contents {
    text-align: left;
    background: #f8f9fa;
    border-radius: 8px;
    padding: 12px;
    margin-bottom: 16px;

    .bundle-sub-section {
      margin-bottom: 8px;

      &:last-child {
        margin-bottom: 0;
      }

      h4 {
        margin: 0 0 6px 0;
        font-size: 13px;
        color: #333;
      }

      ul {
        margin: 0;
        padding: 0;
        list-style: none;

        li {
          font-size: 12px;
          color: #666;
          padding: 2px 0;

          &::before {
            content: '\2022  ';
            color: #409EFF;
          }
        }
      }
    }
  }

  .bundle-install-btn {
    width: 100%;
  }
}

.install-result {
  .install-detail {
    text-align: left;

    h4 {
      margin: 0 0 8px 0;
      font-size: 14px;
      color: #333;
    }
  }
}
</style>