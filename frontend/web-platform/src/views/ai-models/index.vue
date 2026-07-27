<template>
  <div class="ai-models-page">
    <div class="page-header">
      <h2>AI 模型管理</h2>
      <div class="header-actions">
        <el-button @click="handleSeedProviders" :loading="seeding">
          <el-icon><Refresh /></el-icon>
          预置厂商模型
        </el-button>
        <el-button type="primary" @click="showCreateProviderDialog">
          <el-icon><Plus /></el-icon>
          添加厂商
        </el-button>
      </div>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <!-- 厂商管理 -->
      <el-tab-pane label="厂商管理" name="providers">
        <el-table :data="providers" v-loading="loadingProviders" stripe>
          <el-table-column prop="name" label="厂商名称" width="200" />
          <el-table-column prop="code" label="厂商代码" width="150" />
          <el-table-column prop="base_url" label="API 地址" min-width="250" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="模型数" width="100">
            <template #default="{ row }">
              {{ getModelsCount(row.id) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button size="small" @click="showModels(row)">查看模型</el-button>
              <el-button size="small" @click="showEditProviderDialog(row)">编辑</el-button>
              <el-popconfirm title="确定删除此厂商？" @confirm="handleDeleteProvider(row.id)">
                <template #reference>
                  <el-button size="small" type="danger">删除</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 模型管理 -->
      <el-tab-pane label="模型管理" name="models">
        <div class="tab-header">
          <el-select v-model="filterProviderId" placeholder="选择厂商" clearable @change="loadModels" style="width: 200px">
            <el-option
              v-for="p in providers"
              :key="p.id"
              :label="p.name"
              :value="p.id"
            />
          </el-select>
          <el-button type="primary" @click="showCreateModelDialog">
            <el-icon><Plus /></el-icon>
            添加模型
          </el-button>
        </div>

        <el-table :data="filteredModels" v-loading="loadingModels" stripe>
          <el-table-column prop="model_name" label="模型名称" width="200" />
          <el-table-column prop="model_code" label="模型 Code" width="200" />
          <el-table-column label="厂商" width="150">
            <template #default="{ row }">{{ getProviderName(row.provider_id) }}</template>
          </el-table-column>
          <el-table-column prop="model_type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small">{{ modelTypeLabel(row.model_type) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="输入价格" width="120">
            <template #default="{ row }">
              ¥{{ row.input_price_per_1k || 0 }}/1K
            </template>
          </el-table-column>
          <el-table-column label="输出价格" width="120">
            <template #default="{ row }">
              ¥{{ row.output_price_per_1k || 0 }}/1K
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right">
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

    <!-- 创建/编辑厂商对话框 -->
    <el-dialog v-model="providerDialogVisible" :title="isEditProvider ? '编辑厂商' : '添加厂商'" width="600px">
      <el-form :model="providerForm" :rules="providerRules" ref="providerFormRef" label-width="100px">
        <el-form-item label="厂商名称" prop="name">
          <el-input v-model="providerForm.name" placeholder="如：阿里云" />
        </el-form-item>
        <el-form-item label="厂商代码" prop="code">
          <el-input v-model="providerForm.code" placeholder="如：qwen" :disabled="isEditProvider" />
        </el-form-item>
        <el-form-item label="API 地址" prop="base_url">
          <el-input v-model="providerForm.base_url" placeholder="如：https://dashscope.aliyuncs.com/api/v1" />
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="providerForm.status" active-value="active" inactive-value="inactive" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px">
            {{ providerForm.status === 'active' ? '启用' : '禁用' }}
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="providerDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitProvider" :loading="submittingProvider">
          {{ isEditProvider ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 创建/编辑模型对话框 -->
    <el-dialog v-model="modelDialogVisible" :title="isEditModel ? '编辑模型' : '添加模型'" width="600px">
      <el-form :model="modelForm" :rules="modelRules" ref="modelFormRef" label-width="100px">
        <el-form-item label="模型名称" prop="model_name">
          <el-input v-model="modelForm.model_name" placeholder="如：通义千问 Max" />
        </el-form-item>
        <el-form-item label="模型 Code" prop="model_code">
          <el-input v-model="modelForm.model_code" placeholder="如：qwen-max" :disabled="isEditModel" />
        </el-form-item>
        <el-form-item label="所属厂商" prop="provider_id">
          <el-select v-model="modelForm.provider_id" placeholder="选择厂商" style="width: 100%">
            <el-option
              v-for="p in providers"
              :key="p.id"
              :label="p.name"
              :value="p.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模型类型" prop="model_type">
          <el-select v-model="modelForm.model_type" placeholder="选择类型" style="width: 100%">
            <el-option label="对话模型" value="chat" />
            <el-option label="嵌入模型" value="embedding" />
            <el-option label="图像模型" value="image" />
            <el-option label="视频模型" value="video" />
            <el-option label="视觉模型" value="vision" />
            <el-option label="语音合成" value="tts" />
            <el-option label="语音识别" value="stt" />
          </el-select>
        </el-form-item>
        <el-form-item label="输入价格">
          <el-input-number v-model="modelForm.input_price_per_1k" :min="0" :precision="5" :step="0.001" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px">元/1K tokens</span>
        </el-form-item>
        <el-form-item label="输出价格">
          <el-input-number v-model="modelForm.output_price_per_1k" :min="0" :precision="5" :step="0.001" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px">元/1K tokens</span>
        </el-form-item>
        <el-form-item label="状态">
          <el-switch v-model="modelForm.status" active-value="active" inactive-value="inactive" />
          <span style="margin-left: 8px; color: #909399; font-size: 12px">
            {{ modelForm.status === 'active' ? '启用' : '禁用' }}
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="modelDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitModel" :loading="submittingModel">
          {{ isEditModel ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 模型列表对话框 -->
    <el-dialog v-model="modelsDialogVisible" :title="`${currentProvider?.name} - 模型列表`" width="800px">
      <el-table :data="providerModels" stripe>
        <el-table-column prop="model_name" label="模型名称" width="200" />
        <el-table-column prop="model_code" label="模型 Code" width="200" />
        <el-table-column prop="model_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ modelTypeLabel(row.model_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="输入价格" width="120">
          <template #default="{ row }">
            ¥{{ row.input_price_per_1k || 0 }}/1K
          </template>
        </el-table-column>
        <el-table-column label="输出价格" width="120">
          <template #default="{ row }">
            ¥{{ row.output_price_per_1k || 0 }}/1K
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Refresh } from '@element-plus/icons-vue'
import {
  getProviders,
  createProvider,
  updateProvider,
  deleteProvider,
  getModels,
  createModel,
  updateModel,
  deleteModel,
  seedProviders
} from '@/api/ai-model'

const activeTab = ref('providers')

// Providers
const providers = ref<any[]>([])
const loadingProviders = ref(false)

// Models
const allModels = ref<any[]>([])
const loadingModels = ref(false)
const filterProviderId = ref('')

// Provider Dialog
const providerDialogVisible = ref(false)
const isEditProvider = ref(false)
const editingProviderId = ref<string | null>(null)
const submittingProvider = ref(false)
const providerFormRef = ref()
const providerForm = ref({
  name: '',
  code: '',
  base_url: '',
  status: 'active'
})
const providerRules = {
  name: [{ required: true, message: '请输入厂商名称', trigger: 'blur' }],
  code: [{ required: true, message: '请输入厂商代码', trigger: 'blur' }],
  base_url: [{ required: true, message: '请输入 API 地址', trigger: 'blur' }]
}

// Model Dialog
const modelDialogVisible = ref(false)
const isEditModel = ref(false)
const editingModelId = ref<string | null>(null)
const submittingModel = ref(false)
const modelFormRef = ref()
const modelForm = ref({
  model_name: '',
  model_code: '',
  provider_id: '',
  model_type: 'chat',
  input_price_per_1k: 0.01,
  output_price_per_1k: 0.03,
  status: 'active'
})
const modelRules = {
  model_name: [{ required: true, message: '请输入模型名称', trigger: 'blur' }],
  model_code: [{ required: true, message: '请输入模型 Code', trigger: 'blur' }],
  provider_id: [{ required: true, message: '请选择厂商', trigger: 'change' }],
  model_type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

// Models Dialog
const modelsDialogVisible = ref(false)
const currentProvider = ref<any>(null)
const providerModels = computed(() => {
  if (!currentProvider.value) return []
  return allModels.value.filter(m => m.provider_id === currentProvider.value.id)
})

const filteredModels = computed(() => {
  if (!filterProviderId.value) return allModels.value
  return allModels.value.filter(m => m.provider_id === filterProviderId.value)
})

const seeding = ref(false)

function getModelsCount(providerId: string) {
  return allModels.value.filter(m => m.provider_id === providerId).length
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

async function loadProviders() {
  loadingProviders.value = true
  try {
    const data = await getProviders()
    providers.value = data || []
  } catch (e) {
    console.error('加载厂商列表失败', e)
  } finally {
    loadingProviders.value = false
  }
}

async function loadModels() {
  loadingModels.value = true
  try {
    const data = await getModels()
    allModels.value = data || []
  } catch (e) {
    console.error('加载模型列表失败', e)
  } finally {
    loadingModels.value = false
  }
}

function showCreateProviderDialog() {
  isEditProvider.value = false
  editingProviderId.value = null
  providerForm.value = { name: '', code: '', base_url: '', status: 'active' }
  providerDialogVisible.value = true
}

function showEditProviderDialog(provider: any) {
  isEditProvider.value = true
  editingProviderId.value = provider.id
  providerForm.value = {
    name: provider.name,
    code: provider.code,
    base_url: provider.base_url,
    status: provider.status
  }
  providerDialogVisible.value = true
}

async function handleSubmitProvider() {
  if (!providerFormRef.value) return
  await providerFormRef.value.validate()

  submittingProvider.value = true
  try {
    if (isEditProvider.value && editingProviderId.value) {
      await updateProvider(editingProviderId.value, providerForm.value)
      ElMessage.success('厂商已更新')
    } else {
      await createProvider(providerForm.value)
      ElMessage.success('厂商已创建')
    }
    providerDialogVisible.value = false
    await loadProviders()
  } catch (e) {
    console.error('提交厂商失败', e)
  } finally {
    submittingProvider.value = false
  }
}

async function handleDeleteProvider(id: string) {
  try {
    await deleteProvider(id)
    ElMessage.success('厂商已删除')
    await loadProviders()
    await loadModels()
  } catch (e) {
    console.error('删除厂商失败', e)
  }
}

function showCreateModelDialog() {
  isEditModel.value = false
  editingModelId.value = null
  modelForm.value = {
    model_name: '',
    model_code: '',
    provider_id: '',
    model_type: 'chat',
    input_price_per_1k: 0.01,
    output_price_per_1k: 0.03,
    status: 'active'
  }
  modelDialogVisible.value = true
}

function showEditModelDialog(model: any) {
  isEditModel.value = true
  editingModelId.value = model.id
  modelForm.value = {
    model_name: model.model_name,
    model_code: model.model_code,
    provider_id: model.provider_id,
    model_type: model.model_type,
    input_price_per_1k: model.input_price_per_1k || 0.01,
    output_price_per_1k: model.output_price_per_1k || 0.03,
    status: model.status
  }
  modelDialogVisible.value = true
}

async function handleSubmitModel() {
  if (!modelFormRef.value) return
  await modelFormRef.value.validate()

  submittingModel.value = true
  try {
    if (isEditModel.value && editingModelId.value) {
      await updateModel(editingModelId.value, modelForm.value)
      ElMessage.success('模型已更新')
    } else {
      await createModel(modelForm.value)
      ElMessage.success('模型已创建')
    }
    modelDialogVisible.value = false
    await loadModels()
  } catch (e) {
    console.error('提交模型失败', e)
  } finally {
    submittingModel.value = false
  }
}

async function handleDeleteModel(id: string) {
  try {
    await deleteModel(id)
    ElMessage.success('模型已删除')
    await loadModels()
  } catch (e) {
    console.error('删除模型失败', e)
  }
}

function showModels(provider: any) {
  currentProvider.value = provider
  modelsDialogVisible.value = true
}

async function handleSeedProviders() {
  seeding.value = true
  try {
    await seedProviders()
    ElMessage.success('预置厂商和模型已添加')
    await loadProviders()
    await loadModels()
  } catch (e) {
    console.error('预置失败', e)
  } finally {
    seeding.value = false
  }
}

onMounted(() => {
  loadProviders()
  loadModels()
})
</script>

<style scoped lang="scss">
.ai-models-page {
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
    font-weight: 600;
  }

  .header-actions {
    display: flex;
    gap: 12px;
  }
}

.tab-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
</style>
