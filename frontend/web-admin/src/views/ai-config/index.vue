<template>
  <div class="ai-config-page">
    <div class="page-header">
      <h2>AI 模型配置</h2>
      <p>配置您的 AI 模型 API Key 和默认模型</p>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <!-- API Key 管理 -->
      <el-tab-pane label="API Key 管理" name="keys">
        <div class="tab-header">
          <el-button type="primary" @click="showCreateKeyDialog">
            <el-icon><Plus /></el-icon>
            添加 API Key
          </el-button>
        </div>

        <el-table :data="apiKeys" v-loading="loadingKeys" stripe>
          <el-table-column prop="key_name" label="名称" width="200" />
          <el-table-column prop="provider_name" label="厂商" width="150" />
          <el-table-column label="API Key" min-width="200">
            <template #default="{ row }">
              <span class="api-key-masked">{{ maskKey(row.api_key_value || row.api_key) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="custom_base_url" label="API 地址" min-width="180" show-overflow-tooltip>
            <template #default="{ row }">
              {{ row.custom_base_url || '默认' }}
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatTime(row.created_at) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="handleTestKey(row)" :loading="testingKey === row.id">测试</el-button>
              <el-popconfirm title="确定删除此 API Key？" @confirm="handleDeleteKey(row.id)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="apiKeys.length === 0 && !loadingKeys" description="暂无 API Key，请添加" />
      </el-tab-pane>

      <!-- 默认模型配置 -->
      <el-tab-pane label="默认模型配置" name="config">
        <el-form :model="aiConfig" label-width="160px" v-loading="loadingConfig" style="max-width: 650px">
          <el-divider content-position="left">聊天模型（智能体对话）</el-divider>
          <el-form-item label="聊天模型厂商">
            <el-select v-model="aiConfig.default_provider_id" placeholder="选择厂商" @change="handleProviderChange">
              <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="聊天模型">
            <el-select v-model="aiConfig.default_chat_model_id" placeholder="选择聊天模型" filterable>
              <el-option v-for="m in chatModels" :key="m.id" :label="m.model_name + ' (' + m.model_code + ')'" :value="m.id">
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <span>{{ m.model_name }}</span>
                  <span style="color: #999; font-size: 12px;">{{ m.model_code }}</span>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="温度参数">
            <el-slider v-model="aiConfig.temperature" :min="0" :max="2" :step="0.1" show-input />
            <div class="form-tip">值越低回复越确定性，值越高回复越创造性 (0-2)</div>
          </el-form-item>
          <el-form-item label="最大 Token 数">
            <el-input-number v-model="aiConfig.max_tokens" :min="100" :max="32000" :step="100" />
            <div class="form-tip">AI 单次回复的最大 token 数</div>
          </el-form-item>

          <el-divider content-position="left">向量模型（知识库检索）</el-divider>
          <el-form-item label="向量模型">
            <el-select v-model="aiConfig.default_embedding_model_id" placeholder="选择向量模型" filterable>
              <el-option v-for="m in embeddingModels" :key="m.id" :label="m.model_name + ' (' + m.model_code + ')'" :value="m.id">
                <div style="display: flex; justify-content: space-between; align-items: center;">
                  <span>{{ m.model_name }}</span>
                  <span style="color: #999; font-size: 12px;">{{ m.model_code }}</span>
                </div>
              </el-option>
            </el-select>
            <div class="form-tip">用于知识库文档的向量化和语义搜索</div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSaveConfig" :loading="savingConfig">保存配置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- 厂商与模型列表 -->
      <el-tab-pane label="厂商与模型" name="providers">
        <div class="tab-header" style="display: flex; gap: 12px; align-items: center; margin-bottom: 16px;">
          <el-select v-model="filterProviderId" placeholder="全部厂商" clearable style="width: 200px;" @change="loadFilteredModels">
            <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
          <el-select v-model="filterModelType" placeholder="全部类型" clearable style="width: 150px;" @change="loadFilteredModels">
            <el-option label="聊天" value="chat" />
            <el-option label="向量" value="embedding" />
            <el-option label="图片" value="image" />
            <el-option label="视频" value="video" />
            <el-option label="视觉" value="vision" />
            <el-option label="语音合成" value="tts" />
            <el-option label="语音识别" value="stt" />
          </el-select>
          <el-button type="primary" @click="showCreateModelDialog">
            <el-icon><Plus /></el-icon>
            新增模型
          </el-button>
        </div>

        <el-table :data="filteredModels" v-loading="loadingModels" stripe>
          <el-table-column label="厂商" width="140">
            <template #default="{ row }">
              {{ getProviderName(row.provider_id) }}
            </template>
          </el-table-column>
          <el-table-column prop="model_name" label="模型名称" min-width="160" />
          <el-table-column prop="model_code" label="模型 Code" min-width="160" />
          <el-table-column prop="model_type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag :type="modelTypeTag(row.model_type)" size="small">
                {{ modelTypeLabel(row.model_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="输入价格" width="120">
            <template #default="{ row }">
              ¥{{ row.input_price_per_1k }}/1K
            </template>
          </el-table-column>
          <el-table-column label="输出价格" width="120">
            <template #default="{ row }">
              {{ ['embedding','image','tts','stt','video'].includes(row.model_type) ? '-' : '¥' + row.output_price_per_1k + '/1K' }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="showEditModelDialog(row)">编辑</el-button>
              <el-popconfirm title="确定删除此模型？" @confirm="handleDeleteModel(row.id)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 添加 API Key 对话框 -->
    <el-dialog v-model="createKeyDialogVisible" title="添加 API Key" width="500px">
      <el-form :model="keyForm" :rules="keyRules" ref="keyFormRef" label-width="100px">
        <el-form-item label="名称" prop="api_key_name">
          <el-input v-model="keyForm.api_key_name" placeholder="如：生产环境 Key" />
        </el-form-item>
        <el-form-item label="厂商" prop="provider_id">
          <el-select v-model="keyForm.provider_id" placeholder="选择厂商" style="width: 100%">
            <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="API Key" prop="api_key_value">
          <el-input v-model="keyForm.api_key_value" type="password" show-password placeholder="输入您的 API Key" />
        </el-form-item>
        <el-form-item label="API 地址">
          <el-input v-model="keyForm.custom_base_url" placeholder="留空则使用厂商默认地址" />
          <div class="form-tip">如厂商有特殊 API 地址，可在此填写</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createKeyDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreateKey" :loading="creatingKey">创建</el-button>
      </template>
    </el-dialog>

    <!-- 新增/编辑模型对话框 -->
    <el-dialog v-model="modelDialogVisible" :title="editingModel ? '编辑模型' : '新增模型'" width="600px">
      <el-form :model="modelForm" :rules="modelRules" ref="modelFormRef" label-width="120px">
        <el-form-item label="厂商" prop="provider_id">
          <el-select v-model="modelForm.provider_id" placeholder="选择厂商" style="width: 100%" :disabled="!!editingModel">
            <el-option v-for="p in providers" :key="p.id" :label="p.name" :value="p.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="模型 Code" prop="model_code">
          <el-input v-model="modelForm.model_code" placeholder="如 gpt-4o, qwen-max" :disabled="!!editingModel" />
          <div class="form-tip">模型唯一标识，创建后不可修改</div>
        </el-form-item>
        <el-form-item label="模型名称" prop="model_name">
          <el-input v-model="modelForm.model_name" placeholder="如 GPT-4o, 通义千问-Max" />
        </el-form-item>
        <el-form-item label="模型类型" prop="model_type">
          <el-select v-model="modelForm.model_type" placeholder="选择类型" style="width: 100%">
            <el-option label="聊天 (chat)" value="chat" />
            <el-option label="向量 (embedding)" value="embedding" />
            <el-option label="图片生成 (image)" value="image" />
            <el-option label="视频生成 (video)" value="video" />
            <el-option label="视觉理解 (vision)" value="vision" />
            <el-option label="语音合成 (tts)" value="tts" />
            <el-option label="语音识别 (stt)" value="stt" />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="最大输入Token">
              <el-input-number v-model="modelForm.max_input_tokens" :min="0" :step="1024" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最大输出Token">
              <el-input-number v-model="modelForm.max_output_tokens" :min="0" :step="1024" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="输入价格/1K">
              <el-input-number v-model="modelForm.input_price_per_1k" :min="0" :precision="5" :step="0.001" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="输出价格/1K">
              <el-input-number v-model="modelForm.output_price_per_1k" :min="0" :precision="5" :step="0.001" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="默认模型">
          <el-switch v-model="modelForm.is_default" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px;">设为该厂商的默认推荐模型</span>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="modelForm.description" type="textarea" :rows="2" placeholder="模型描述（可选）" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSaveModel" :loading="savingModel">
          {{ editingModel ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getTenantAIConfig,
  updateTenantAIConfig,
  getProviders,
  getModels,
  getAPIKeys,
  createAPIKey,
  deleteAPIKey,
  testAPIKey,
  adminGetModels,
  adminCreateModel,
  adminUpdateModel,
  adminDeleteModel,
  type TenantAIConfig,
  type APIKeyTestResult
} from '@/api/ai-config'

const activeTab = ref('keys')

// API Keys
const apiKeys = ref<any[]>([])
const loadingKeys = ref(false)
const testingKey = ref<string | null>(null)

// AI Config
const aiConfig = ref({
  default_provider_id: '',
  default_chat_model_id: '',
  default_embedding_model_id: '',
  temperature: 0.7,
  max_tokens: 2048
})
const loadingConfig = ref(false)
const savingConfig = ref(false)

// Providers & Models
const providers = ref<any[]>([])
const allModels = ref<any[]>([])
const loadingProviders = ref(false)
const loadingModels = ref(false)

// Filter
const filterProviderId = ref('')
const filterModelType = ref('')

// Create Key Dialog
const createKeyDialogVisible = ref(false)
const creatingKey = ref(false)
const keyFormRef = ref()
const keyForm = ref({
  api_key_name: '',
  provider_id: '',
  api_key_value: '',
  custom_base_url: ''
})
const keyRules = {
  api_key_name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  provider_id: [{ required: true, message: '请选择厂商', trigger: 'change' }],
  api_key_value: [{ required: true, message: '请输入 API Key', trigger: 'blur' }]
}

// Model Dialog
const modelDialogVisible = ref(false)
const editingModel = ref<any>(null)
const savingModel = ref(false)
const modelFormRef = ref()
const modelForm = ref({
  provider_id: '',
  model_code: '',
  model_name: '',
  model_type: 'chat',
  max_input_tokens: 4096,
  max_output_tokens: 2048,
  input_price_per_1k: 0.01,
  output_price_per_1k: 0.03,
  is_default: false,
  description: ''
})
const modelRules = {
  provider_id: [{ required: true, message: '请选择厂商', trigger: 'change' }],
  model_code: [{ required: true, message: '请输入模型 Code', trigger: 'blur' }],
  model_name: [{ required: true, message: '请输入模型名称', trigger: 'blur' }],
  model_type: [{ required: true, message: '请选择模型类型', trigger: 'change' }]
}

// 按类型筛选模型（用于默认模型配置tab）
const chatModels = computed(() => {
  return allModels.value.filter(m =>
    m.model_type === 'chat' &&
    (!aiConfig.value.default_provider_id || m.provider_id === aiConfig.value.default_provider_id)
  )
})

const embeddingModels = computed(() => {
  return allModels.value.filter(m => m.model_type === 'embedding')
})

// 厂商与模型tab的筛选列表
const filteredModels = computed(() => {
  let list = allModels.value
  if (filterProviderId.value) {
    list = list.filter(m => m.provider_id === filterProviderId.value)
  }
  if (filterModelType.value) {
    list = list.filter(m => m.model_type === filterModelType.value)
  }
  return list
})

function maskKey(key: string) {
  if (!key || key.length < 8) return key
  return key.substring(0, 4) + '****' + key.substring(key.length - 4)
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

function getProviderName(providerId: string) {
  const p = providers.value.find(p => p.id === providerId)
  return p ? p.name : '-'
}

const modelTypeMap: Record<string, string> = {
  chat: '聊天',
  embedding: '向量',
  image: '图片',
  video: '视频',
  vision: '视觉',
  tts: '语音合成',
  stt: '语音识别'
}

function modelTypeLabel(type: string) {
  return modelTypeMap[type] || type
}

function modelTypeTag(type: string) {
  const map: Record<string, string> = {
    chat: '',
    embedding: 'warning',
    image: 'success',
    video: 'danger',
    vision: 'info',
    tts: 'success',
    stt: 'info'
  }
  return map[type] || ''
}

async function loadAPIKeys() {
  loadingKeys.value = true
  try {
    const data = await getAPIKeys()
    apiKeys.value = Array.isArray(data) ? data : (data as any)?.items || []
  } catch (e) {
    console.error('加载 API Keys 失败', e)
  } finally {
    loadingKeys.value = false
  }
}

async function loadAIConfig() {
  loadingConfig.value = true
  try {
    const data = await getTenantAIConfig() as TenantAIConfig
    if (data) {
      aiConfig.value = {
        default_provider_id: data.default_provider_id || '',
        default_chat_model_id: data.default_chat_model_id || '',
        default_embedding_model_id: data.default_embedding_model_id || '',
        temperature: data.config?.temperature || 0.7,
        max_tokens: data.config?.max_tokens || 2048
      }
    }
  } catch (e) {
    console.error('加载 AI 配置失败', e)
  } finally {
    loadingConfig.value = false
  }
}

async function loadProviders() {
  loadingProviders.value = true
  try {
    const data = await getProviders()
    providers.value = Array.isArray(data) ? data : (data as any)?.items || []
  } catch (e) {
    console.error('加载厂商列表失败', e)
  } finally {
    loadingProviders.value = false
  }
}

async function loadAllModels() {
  loadingModels.value = true
  try {
    const data = await adminGetModels()
    allModels.value = Array.isArray(data) ? data : (data as any)?.items || []
  } catch (e) {
    // fallback to tenant endpoint
    try {
      const data = await getModels()
      allModels.value = Array.isArray(data) ? data : (data as any)?.items || []
    } catch (e2) {
      console.error('加载模型列表失败', e2)
    }
  } finally {
    loadingModels.value = false
  }
}

async function loadFilteredModels() {
  // filteredModels is computed, no need to reload
}

function handleProviderChange() {
  aiConfig.value.default_chat_model_id = ''
}

async function handleSaveConfig() {
  savingConfig.value = true
  try {
    await updateTenantAIConfig({
      default_provider_id: aiConfig.value.default_provider_id || null,
      default_chat_model_id: aiConfig.value.default_chat_model_id || null,
      default_embedding_model_id: aiConfig.value.default_embedding_model_id || null,
      config: {
        temperature: aiConfig.value.temperature,
        max_tokens: aiConfig.value.max_tokens
      }
    })
    ElMessage.success('配置已保存')
  } catch (e) {
    console.error('保存配置失败', e)
  } finally {
    savingConfig.value = false
  }
}

function showCreateKeyDialog() {
  keyForm.value = { api_key_name: '', provider_id: '', api_key_value: '', custom_base_url: '' }
  createKeyDialogVisible.value = true
}

async function handleCreateKey() {
  if (!keyFormRef.value) return
  await keyFormRef.value.validate()
  creatingKey.value = true
  try {
    await createAPIKey(keyForm.value)
    ElMessage.success('API Key 已创建')
    createKeyDialogVisible.value = false
    await loadAPIKeys()
  } catch (e) {
    console.error('创建 API Key 失败', e)
  } finally {
    creatingKey.value = false
  }
}

async function handleDeleteKey(id: string) {
  try {
    await deleteAPIKey(id)
    ElMessage.success('API Key 已删除')
    await loadAPIKeys()
  } catch (e) {
    console.error('删除 API Key 失败', e)
  }
}

async function handleTestKey(row: any) {
  testingKey.value = row.id
  try {
    const result = await testAPIKey(row.id) as APIKeyTestResult
    if (result.valid || result.success) {
      ElMessage.success('API Key 测试通过')
    } else {
      ElMessage.error(result.error || 'API Key 测试失败')
    }
  } catch (e) {
    console.error('测试 API Key 失败', e)
  } finally {
    testingKey.value = null
  }
}

// ===== 模型 CRUD =====

function showCreateModelDialog() {
  editingModel.value = null
  modelForm.value = {
    provider_id: filterProviderId.value || '',
    model_code: '',
    model_name: '',
    model_type: 'chat',
    max_input_tokens: 4096,
    max_output_tokens: 2048,
    input_price_per_1k: 0.01,
    output_price_per_1k: 0.03,
    is_default: false,
    description: ''
  }
  modelDialogVisible.value = true
}

function showEditModelDialog(row: any) {
  editingModel.value = row
  modelForm.value = {
    provider_id: row.provider_id,
    model_code: row.model_code,
    model_name: row.model_name,
    model_type: row.model_type,
    max_input_tokens: row.max_input_tokens || 4096,
    max_output_tokens: row.max_output_tokens || 2048,
    input_price_per_1k: row.input_price_per_1k || 0.01,
    output_price_per_1k: row.output_price_per_1k || 0.03,
    is_default: row.is_default || false,
    description: row.description || ''
  }
  modelDialogVisible.value = true
}

async function handleSaveModel() {
  if (!modelFormRef.value) return
  await modelFormRef.value.validate()
  savingModel.value = true
  try {
    const payload = {
      provider_id: modelForm.value.provider_id,
      model_code: modelForm.value.model_code,
      model_name: modelForm.value.model_name,
      model_type: modelForm.value.model_type,
      max_input_tokens: modelForm.value.max_input_tokens,
      max_output_tokens: modelForm.value.max_output_tokens,
      input_price_per_1k: modelForm.value.input_price_per_1k,
      output_price_per_1k: modelForm.value.output_price_per_1k,
      is_default: modelForm.value.is_default,
      description: modelForm.value.description,
      status: 'active'
    }

    if (editingModel.value) {
      await adminUpdateModel(editingModel.value.id, payload)
      ElMessage.success('模型已更新')
    } else {
      await adminCreateModel(payload)
      ElMessage.success('模型已创建')
    }
    modelDialogVisible.value = false
    await loadAllModels()
  } catch (e) {
    console.error('保存模型失败', e)
  } finally {
    savingModel.value = false
  }
}

async function handleDeleteModel(id: string) {
  try {
    await adminDeleteModel(id)
    ElMessage.success('模型已删除')
    await loadAllModels()
  } catch (e) {
    console.error('删除模型失败', e)
  }
}

onMounted(() => {
  loadAPIKeys()
  loadAIConfig()
  loadProviders()
  loadAllModels()
})
</script>

<style scoped lang="scss">
.ai-config-page {
  padding: 0;
}
.page-header {
  margin-bottom: 24px;
  h2 { margin: 0 0 8px 0; font-size: 20px; font-weight: 600; }
  p { margin: 0; color: #909399; font-size: 14px; }
}
.tab-header {
  margin-bottom: 16px;
}
.api-key-masked {
  font-family: monospace;
  color: #606266;
}
.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}
</style>
