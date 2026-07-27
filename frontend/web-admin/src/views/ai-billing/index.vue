<template>
  <div class="ai-billing-page">
    <div class="page-header">
      <h2>AI 计费管理</h2>
    </div>

    <el-tabs v-model="activeTab" type="border-card">
      <!-- Tab 1: 余额与资源包 -->
      <el-tab-pane label="余额与资源包" name="balance">
        <div v-loading="balanceLoading">
          <!-- 余额卡片 -->
          <el-row :gutter="20" class="balance-section">
            <el-col :span="8">
              <el-card class="balance-card" shadow="hover">
                <div class="balance-label">当前余额</div>
                <div class="balance-value primary">¥{{ formatMoney(balanceInfo?.balance) }}</div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card class="balance-card" shadow="hover">
                <div class="balance-label">累计充值</div>
                <div class="balance-value success">¥{{ formatMoney(balanceInfo?.total_recharged) }}</div>
              </el-card>
            </el-col>
            <el-col :span="8">
              <el-card class="balance-card" shadow="hover">
                <div class="balance-label">累计消费</div>
                <div class="balance-value warning">¥{{ formatMoney(balanceInfo?.total_consumed) }}</div>
              </el-card>
            </el-col>
          </el-row>

          <!-- 活跃资源包 -->
          <el-card style="margin-top: 20px">
            <template #header>
              <div class="card-header">
                <span class="card-title">活跃资源包</span>
                <el-button type="primary" @click="showPackageDialog">购买资源包</el-button>
              </div>
            </template>
            <el-table :data="balanceInfo?.active_packages || []" stripe>
              <el-table-column prop="name" label="资源包名称" min-width="150" />
              <el-table-column label="剩余 Token" min-width="150">
                <template #default="{ row }">
                  {{ formatNumber(row.remaining_tokens) }} / {{ formatNumber(row.total_tokens) }}
                </template>
              </el-table-column>
              <el-table-column label="到期时间" min-width="120">
                <template #default="{ row }">{{ formatDate(row.expires_at) }}</template>
              </el-table-column>
              <el-table-column label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                    {{ row.status === 'active' ? '使用中' : row.status }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-if="!balanceInfo?.active_packages?.length" description="暂无活跃资源包" />
          </el-card>
        </div>

        <!-- 购买资源包对话框 -->
        <el-dialog v-model="packageDialogVisible" title="购买资源包" width="700px">
          <div v-loading="packagesLoading">
            <el-row :gutter="16">
              <el-col :span="12" v-for="pkg in availablePackages" :key="pkg.id" style="margin-bottom: 16px">
                <el-card class="package-card" shadow="hover" :class="{ purchasing: purchasingId === pkg.id }">
                  <div class="package-name">{{ pkg.name }}</div>
                  <div class="package-tokens">{{ formatNumber(pkg.tokens) }} Tokens</div>
                  <div class="package-price">¥{{ formatMoney(pkg.price) }}</div>
                  <div class="package-desc" v-if="pkg.description">{{ pkg.description }}</div>
                  <el-button
                    type="primary"
                    style="width: 100%; margin-top: 12px"
                    :loading="purchasingId === pkg.id"
                    @click="handlePurchase(pkg)"
                  >
                    立即购买
                  </el-button>
                </el-card>
              </el-col>
            </el-row>
            <el-empty v-if="!availablePackages.length && !packagesLoading" description="暂无可购买的资源包" />
          </div>
        </el-dialog>
      </el-tab-pane>

      <!-- Tab 2: 用量明细 -->
      <el-tab-pane label="用量明细" name="usage">
        <el-card>
          <template #header>
            <div class="card-header">
              <span class="card-title">用量明细</span>
              <div class="filter-area">
                <el-date-picker
                  v-model="usageDateRange"
                  type="daterange"
                  range-separator="至"
                  start-placeholder="开始日期"
                  end-placeholder="结束日期"
                  value-format="YYYY-MM-DD"
                  @change="loadUsageDetails"
                  style="width: 280px"
                />
              </div>
            </div>
          </template>
          <el-table :data="usageList" stripe v-loading="usageLoading">
            <el-table-column prop="date" label="日期" width="120">
              <template #default="{ row }">{{ formatDate(row.date) }}</template>
            </el-table-column>
            <el-table-column prop="model_name" label="模型名称" min-width="140" />
            <el-table-column label="输入 Token" width="120" align="right">
              <template #default="{ row }">{{ formatNumber(row.input_tokens) }}</template>
            </el-table-column>
            <el-table-column label="输出 Token" width="120" align="right">
              <template #default="{ row }">{{ formatNumber(row.output_tokens) }}</template>
            </el-table-column>
            <el-table-column label="总 Token" width="120" align="right">
              <template #default="{ row }">{{ formatNumber(row.total_tokens) }}</template>
            </el-table-column>
            <el-table-column label="费用" width="100" align="right">
              <template #default="{ row }">¥{{ formatMoney(row.cost) }}</template>
            </el-table-column>
            <el-table-column label="计费模式" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ billingModeLabel(row.billing_mode) }}</el-tag>
              </template>
            </el-table-column>
          </el-table>
          <div class="pagination-wrapper" v-if="usageTotal > usagePageSize">
            <el-pagination
              layout="total, prev, pager, next"
              :total="usageTotal"
              :page-size="usagePageSize"
              v-model:current-page="usagePage"
              @current-change="loadUsageDetails"
            />
          </div>
        </el-card>
      </el-tab-pane>

      <!-- Tab 3: 模型定价 -->
      <el-tab-pane label="模型定价" name="pricing">
        <el-card>
          <template #header>
            <span class="card-title">平台模型定价</span>
          </template>
          <el-table :data="pricingList" stripe v-loading="pricingLoading">
            <el-table-column prop="model_name" label="模型名称" min-width="200" />
            <el-table-column label="输入价格 (每百万 Token)" min-width="180" align="right">
              <template #default="{ row }">¥{{ formatMoney(row.input_price) }}</template>
            </el-table-column>
            <el-table-column label="输出价格 (每百万 Token)" min-width="180" align="right">
              <template #default="{ row }">¥{{ formatMoney(row.output_price) }}</template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getBalance,
  getPackages,
  purchasePackage,
  getUsageDetails,
  getPlatformPricing,
  type BalanceInfo,
  type AvailablePackage,
  type UsageDetail,
  type ModelPricing
} from '@/api/ai-billing'

const activeTab = ref('balance')

// ===== 余额与资源包 =====
const balanceLoading = ref(false)
const balanceInfo = ref<BalanceInfo | null>(null)

const loadBalance = async () => {
  balanceLoading.value = true
  try {
    const res = await getBalance()
    balanceInfo.value = (res as any)?.data || res
  } catch {
    balanceInfo.value = null
  } finally {
    balanceLoading.value = false
  }
}

// ===== 购买资源包 =====
const packageDialogVisible = ref(false)
const packagesLoading = ref(false)
const availablePackages = ref<AvailablePackage[]>([])
const purchasingId = ref<string | null>(null)

const showPackageDialog = async () => {
  packageDialogVisible.value = true
  packagesLoading.value = true
  try {
    const res = await getPackages()
    availablePackages.value = Array.isArray(res) ? res : ((res as any)?.data || (res as any)?.items || [])
  } catch {
    availablePackages.value = []
  } finally {
    packagesLoading.value = false
  }
}

const handlePurchase = async (pkg: AvailablePackage) => {
  try {
    await ElMessageBox.confirm(
      `确定购买「${pkg.name}」资源包？费用：¥${formatMoney(pkg.price)}`,
      '确认购买',
      { confirmButtonText: '确认购买', cancelButtonText: '取消', type: 'info' }
    )
  } catch {
    return
  }

  purchasingId.value = pkg.id
  try {
    await purchasePackage(pkg.id)
    ElMessage.success('购买成功')
    packageDialogVisible.value = false
    await loadBalance()
  } catch (e: any) {
    ElMessage.error(e.message || '购买失败')
  } finally {
    purchasingId.value = null
  }
}

// ===== 用量明细 =====
const usageLoading = ref(false)
const usageList = ref<UsageDetail[]>([])
const usagePage = ref(1)
const usagePageSize = ref(20)
const usageTotal = ref(0)
const usageDateRange = ref<string[]>([])

const loadUsageDetails = async () => {
  usageLoading.value = true
  try {
    const params: any = {
      page: usagePage.value,
      page_size: usagePageSize.value
    }
    if (usageDateRange.value && usageDateRange.value.length >= 2) {
      params.start_date = usageDateRange.value[0]
      params.end_date = usageDateRange.value[1]
    }
    const res = await getUsageDetails(params)
    const data = (res as any)?.data || res || {}
    usageList.value = data.items || data.records || (Array.isArray(data) ? data : [])
    usageTotal.value = data.total || usageList.value.length || 0
  } catch {
    usageList.value = []
    usageTotal.value = 0
  } finally {
    usageLoading.value = false
  }
}

const billingModeLabel = (mode: string) => {
  const map: Record<string, string> = {
    token: 'Token 计费',
    package: '资源包抵扣',
    balance: '余额扣费',
    free: '免费'
  }
  return map[mode] || mode || '-'
}

// ===== 模型定价 =====
const pricingLoading = ref(false)
const pricingList = ref<ModelPricing[]>([])

const loadPricing = async () => {
  pricingLoading.value = true
  try {
    const res = await getPlatformPricing()
    pricingList.value = Array.isArray(res) ? res : ((res as any)?.data || (res as any)?.items || [])
  } catch {
    pricingList.value = []
  } finally {
    pricingLoading.value = false
  }
}

// ===== 工具函数 =====
const formatMoney = (val: number | undefined) => {
  if (val === undefined || val === null) return '0.00'
  return Number(val).toFixed(2)
}

const formatNumber = (val: number | undefined) => {
  if (val === undefined || val === null) return '0'
  return Number(val).toLocaleString()
}

const formatDate = (date: string | undefined) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

// ===== 初始化 =====
onMounted(async () => {
  await loadBalance()
  await loadUsageDetails()
  await loadPricing()
})
</script>

<style scoped lang="scss">
.ai-billing-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 20px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.balance-section {
  margin-bottom: 0;
}

.balance-card {
  text-align: center;

  .balance-label {
    font-size: 14px;
    color: #909399;
    margin-bottom: 8px;
  }

  .balance-value {
    font-size: 32px;
    font-weight: bold;

    &.primary {
      color: #409eff;
    }

    &.success {
      color: #67c23a;
    }

    &.warning {
      color: #e6a23c;
    }
  }
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-title {
  font-weight: 600;
  font-size: 16px;
}

.package-card {
  text-align: center;
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-2px);
  }

  .package-name {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 8px;
  }

  .package-tokens {
    font-size: 14px;
    color: #606266;
    margin-bottom: 4px;
  }

  .package-price {
    font-size: 24px;
    font-weight: bold;
    color: #409eff;
    margin-bottom: 4px;
  }

  .package-desc {
    font-size: 12px;
    color: #909399;
  }
}

.filter-area {
  display: flex;
  gap: 12px;
  align-items: center;
}

.pagination-wrapper {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
