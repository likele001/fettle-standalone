<template>
  <div class="page-container">
    <div class="page-header">
      <h2>租户余额管理</h2>
    </div>

    <el-card shadow="hover">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <el-input
          v-model="keyword"
          placeholder="搜索租户名称"
          clearable
          style="width: 280px"
          @clear="loadBalances"
          @keyup.enter="loadBalances"
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>

      <el-table :data="balances" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="tenant_name" label="租户名称" width="200" />
        <el-table-column prop="tenant_id" label="租户ID" width="200">
          <template #default="{ row }">
            {{ row.tenant_id?.slice(0, 8) }}...
          </template>
        </el-table-column>
        <el-table-column label="当前余额" width="150">
          <template #default="{ row }">
            <span class="balance-value">¥{{ formatMoney(row.balance) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="累计充值" width="150">
          <template #default="{ row }">
            ¥{{ formatMoney(row.total_recharged) }}
          </template>
        </el-table-column>
        <el-table-column label="累计消费" width="150">
          <template #default="{ row }">
            ¥{{ formatMoney(row.total_consumed) }}
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="170">
          <template #default="{ row }">
            {{ row.updated_at ? formatTime(row.updated_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="{ row }">
            <el-button size="small" type="primary" @click="showRechargeDialog(row)">充值</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 充值对话框 -->
    <el-dialog v-model="rechargeDialogVisible" title="租户充值" width="480px">
      <el-form :model="rechargeForm" :rules="rechargeRules" ref="rechargeFormRef" label-width="100px">
        <el-form-item label="租户名称">
          <span>{{ currentTenant?.tenant_name }}</span>
        </el-form-item>
        <el-form-item label="当前余额">
          <span>¥{{ formatMoney(currentTenant?.balance || 0) }}</span>
        </el-form-item>
        <el-form-item label="充值金额" prop="amount">
          <el-input-number
            v-model="rechargeForm.amount"
            :min="1"
            :max="1000000"
            :precision="2"
            :step="100"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="rechargeForm.remark"
            type="textarea"
            :rows="3"
            placeholder="可选，填写充值备注"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rechargeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleRecharge" :loading="recharging">确认充值</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import { getBalancesList, rechargeTenant, type TenantBalance } from '@/api/ai-billing'

const balances = ref<TenantBalance[]>([])
const loading = ref(false)
const keyword = ref('')

// 充值对话框
const rechargeDialogVisible = ref(false)
const currentTenant = ref<TenantBalance | null>(null)
const recharging = ref(false)
const rechargeFormRef = ref()
const rechargeForm = ref({
  amount: 100,
  remark: ''
})
const rechargeRules = {
  amount: [{ required: true, message: '请输入充值金额', trigger: 'blur' }]
}

function formatMoney(amount: number) {
  if (!amount && amount !== 0) return '0.00'
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

async function loadBalances() {
  loading.value = true
  try {
    const data = await getBalancesList({ keyword: keyword.value || undefined })
    balances.value = data || []
  } catch {
    balances.value = []
  } finally {
    loading.value = false
  }
}

function showRechargeDialog(tenant: TenantBalance) {
  currentTenant.value = tenant
  rechargeForm.value = { amount: 100, remark: '' }
  rechargeDialogVisible.value = true
}

async function handleRecharge() {
  if (!rechargeFormRef.value) return
  await rechargeFormRef.value.validate()

  recharging.value = true
  try {
    await rechargeTenant({
      tenant_id: currentTenant.value!.tenant_id,
      amount: rechargeForm.value.amount,
      remark: rechargeForm.value.remark || undefined
    })
    ElMessage.success('充值成功')
    rechargeDialogVisible.value = false
    await loadBalances()
  } catch {
    // error handled by request interceptor
  } finally {
    recharging.value = false
  }
}

onMounted(() => {
  loadBalances()
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

.balance-value {
  font-weight: 600;
  color: #3b82f6;
}
</style>
