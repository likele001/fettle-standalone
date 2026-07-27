<template>
  <div class="page-container">
    <div class="page-header">
      <h2>模型定价管理</h2>
    </div>

    <el-card shadow="hover">
      <el-table :data="pricingList" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="model_name" label="模型名称" width="200" />
        <el-table-column prop="provider_name" label="厂商" width="150" />
        <el-table-column prop="model_type" label="类型" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ typeLabel(row.model_type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="输入价格 (元/1K tokens)" width="220">
          <template #default="{ row }">
            <div v-if="editingId === row.model_id" style="display: flex; align-items: center; gap: 8px;">
              <el-input-number
                v-model="editForm.input_price_per_1k"
                :min="0"
                :precision="5"
                :step="0.001"
                size="small"
                style="width: 150px"
              />
            </div>
            <span v-else>¥{{ row.input_price_per_1k?.toFixed(5) || '0.00000' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="输出价格 (元/1K tokens)" width="220">
          <template #default="{ row }">
            <div v-if="editingId === row.model_id" style="display: flex; align-items: center; gap: 8px;">
              <el-input-number
                v-model="editForm.output_price_per_1k"
                :min="0"
                :precision="5"
                :step="0.001"
                size="small"
                style="width: 150px"
              />
            </div>
            <span v-else>¥{{ row.output_price_per_1k?.toFixed(5) || '0.00000' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-switch
              :model-value="row.status === 'active'"
              @change="(val: boolean) => handleToggleStatus(row, val)"
              active-value="active"
              inactive-value="inactive"
              inline-prompt
              active-text="启"
              inactive-text="禁"
            />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <template v-if="editingId === row.model_id">
              <el-button size="small" type="primary" @click="handleSave(row)" :loading="saving">保存</el-button>
              <el-button size="small" @click="cancelEdit">取消</el-button>
            </template>
            <template v-else>
              <el-button size="small" @click="startEdit(row)">编辑价格</el-button>
            </template>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { getPricingList, updatePricing, type ModelPricing } from '@/api/ai-billing'

const pricingList = ref<ModelPricing[]>([])
const loading = ref(false)
const saving = ref(false)
const editingId = ref<string | null>(null)
const editForm = ref({
  input_price_per_1k: 0,
  output_price_per_1k: 0
})

const typeMap: Record<string, string> = {
  chat: '聊天',
  embedding: '向量',
  image: '图片',
  video: '视频',
  vision: '视觉',
  tts: '语音合成',
  stt: '语音识别'
}

function typeLabel(type: string) {
  return typeMap[type] || type
}

async function loadPricing() {
  loading.value = true
  try {
    const data = await getPricingList()
    pricingList.value = data || []
  } catch {
    pricingList.value = []
  } finally {
    loading.value = false
  }
}

function startEdit(row: ModelPricing) {
  editingId.value = row.model_id
  editForm.value = {
    input_price_per_1k: row.input_price_per_1k || 0,
    output_price_per_1k: row.output_price_per_1k || 0
  }
}

function cancelEdit() {
  editingId.value = null
}

async function handleSave(row: ModelPricing) {
  saving.value = true
  try {
    await updatePricing(row.model_id, {
      input_price_per_1k: editForm.value.input_price_per_1k,
      output_price_per_1k: editForm.value.output_price_per_1k
    })
    ElMessage.success('价格已更新')
    editingId.value = null
    await loadPricing()
  } catch {
    // error handled by request interceptor
  } finally {
    saving.value = false
  }
}

async function handleToggleStatus(row: ModelPricing, val: boolean) {
  try {
    await updatePricing(row.model_id, {
      status: val ? 'active' : 'inactive'
    })
    ElMessage.success(val ? '已启用' : '已禁用')
    await loadPricing()
  } catch {
    // error handled by request interceptor
  }
}

onMounted(() => {
  loadPricing()
})
</script>

<style scoped lang="scss">
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
}
</style>
