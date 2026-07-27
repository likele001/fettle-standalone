<template>
  <div class="page-container">
    <div class="page-header">
      <h2>资源包管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建资源包
      </el-button>
    </div>

    <el-card shadow="hover">
      <el-table :data="packages" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="name" label="资源包名称" width="200" />
        <el-table-column label="Token 额度" width="150">
          <template #default="{ row }">
            {{ formatTokens(row.token_amount) }}
          </template>
        </el-table-column>
        <el-table-column label="价格" width="120">
          <template #default="{ row }">
            <span class="price-value">¥{{ formatMoney(row.price) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip>
          <template #default="{ row }">
            {{ row.description || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '启用' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ row.created_at ? formatTime(row.created_at) : '-' }}
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 创建资源包对话框 -->
    <el-dialog v-model="createDialogVisible" title="创建资源包" width="520px">
      <el-form :model="createForm" :rules="createRules" ref="createFormRef" label-width="110px">
        <el-form-item label="资源包名称" prop="name">
          <el-input v-model="createForm.name" placeholder="如：100万 Token 包" />
        </el-form-item>
        <el-form-item label="Token 额度" prop="token_amount">
          <el-input-number
            v-model="createForm.token_amount"
            :min="1000"
            :step="100000"
            style="width: 100%"
          />
          <span style="margin-left: 8px; color: #909399; font-size: 12px">tokens</span>
        </el-form-item>
        <el-form-item label="价格" prop="price">
          <el-input-number
            v-model="createForm.price"
            :min="0"
            :precision="2"
            :step="10"
            style="width: 100%"
          />
          <span style="margin-left: 8px; color: #909399; font-size: 12px">元</span>
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="createForm.description"
            type="textarea"
            :rows="3"
            placeholder="可选，描述资源包详情"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="creating">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getPackagesList, createPackage, type ResourcePackage } from '@/api/ai-billing'

const packages = ref<ResourcePackage[]>([])
const loading = ref(false)

// 创建对话框
const createDialogVisible = ref(false)
const creating = ref(false)
const createFormRef = ref()
const createForm = ref({
  name: '',
  token_amount: 1000000,
  price: 100,
  description: ''
})
const createRules = {
  name: [{ required: true, message: '请输入资源包名称', trigger: 'blur' }],
  token_amount: [{ required: true, message: '请输入 Token 额度', trigger: 'blur' }],
  price: [{ required: true, message: '请输入价格', trigger: 'blur' }]
}

function formatTokens(amount: number) {
  if (!amount) return '0'
  if (amount >= 1000000) return (amount / 1000000).toFixed(amount % 1000000 === 0 ? 0 : 1) + 'M'
  if (amount >= 1000) return (amount / 1000).toFixed(amount % 1000 === 0 ? 0 : 1) + 'K'
  return amount.toString()
}

function formatMoney(amount: number) {
  if (!amount && amount !== 0) return '0.00'
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

async function loadPackages() {
  loading.value = true
  try {
    const data = await getPackagesList()
    packages.value = data || []
  } catch {
    packages.value = []
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  createForm.value = {
    name: '',
    token_amount: 1000000,
    price: 100,
    description: ''
  }
  createDialogVisible.value = true
}

async function handleCreate() {
  if (!createFormRef.value) return
  await createFormRef.value.validate()

  creating.value = true
  try {
    await createPackage({
      name: createForm.value.name,
      token_amount: createForm.value.token_amount,
      price: createForm.value.price,
      description: createForm.value.description || undefined
    })
    ElMessage.success('资源包已创建')
    createDialogVisible.value = false
    await loadPackages()
  } catch {
    // error handled by request interceptor
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  loadPackages()
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

.price-value {
  font-weight: 600;
  color: #3b82f6;
}
</style>
