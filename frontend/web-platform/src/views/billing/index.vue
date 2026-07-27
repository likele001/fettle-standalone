<template>
  <div class="page-container">
    <div class="page-header">
      <h2>计费管理</h2>
    </div>

    <!-- 统计卡片 -->
    <el-row :gutter="20" class="stats-row">
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-info">
            <div class="stat-value">{{ subscriptionTotal }}</div>
            <div class="stat-label">总订阅数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-info">
            <div class="stat-value">{{ invoiceTotal }}</div>
            <div class="stat-label">总账单数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="8">
        <el-card shadow="hover">
          <div class="stat-info">
            <div class="stat-value">{{ plans.length }}</div>
            <div class="stat-label">套餐数量</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 订阅列表 -->
    <el-card shadow="hover" style="margin-top: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h3>订阅列表</h3>
      </div>
      <el-table :data="subscriptions" v-loading="subscriptionLoading" stripe style="width: 100%">
        <el-table-column prop="tenant_id" label="租户ID" width="200">
          <template #default="{ row }">
            {{ row.tenant_id.slice(0, 8) }}...
          </template>
        </el-table-column>
        <el-table-column prop="plan_name" label="套餐" width="120" />
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="subStatusTagType(row.status)" size="small">
              {{ subStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="开始时间" width="170">
          <template #default="{ row }">
            {{ row.started_at ? formatTime(row.started_at) : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="到期时间" width="170">
          <template #default="{ row }">
            {{ row.expires_at ? formatTime(row.expires_at) : '-' }}
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
        <el-pagination
          v-model:current-page="subPage"
          v-model:page-size="subPageSize"
          :total="subscriptionTotal"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadSubscriptions"
        />
      </div>
    </el-card>

    <!-- 可用套餐 -->
    <el-card shadow="hover" style="margin-top: 20px;">
      <h3>套餐列表</h3>
      <el-row :gutter="20" style="margin-top: 16px;">
        <el-col :span="8" v-for="plan in plans" :key="plan.id">
          <el-card shadow="hover" :class="{ 'popular-plan': plan.is_popular }">
            <div class="plan-card">
              <h4>{{ plan.name }}</h4>
              <div class="plan-card-price">¥{{ plan.price }}/月</div>
              <p class="plan-card-desc">{{ plan.description }}</p>
              <ul class="plan-card-features">
                <li v-for="(f, i) in plan.features" :key="i">{{ f }}</li>
              </ul>
              <el-tag :type="plan.is_active ? 'success' : 'info'" size="small" style="margin-top: 8px;">
                {{ plan.is_active ? '启用' : '禁用' }}
              </el-tag>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- 账单记录 -->
    <el-card shadow="hover" style="margin-top: 20px;">
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
        <h3>账单记录</h3>
        <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 130px" @change="loadInvoices">
          <el-option label="全部" value="" />
          <el-option label="待支付" value="pending" />
          <el-option label="已支付" value="paid" />
          <el-option label="支付失败" value="failed" />
          <el-option label="已取消" value="cancelled" />
        </el-select>
      </div>
      <el-table :data="invoices" v-loading="invoiceLoading" stripe style="width: 100%">
        <el-table-column prop="id" label="账单ID" width="180">
          <template #default="{ row }">
            {{ row.id.slice(0, 8) }}...
          </template>
        </el-table-column>
        <el-table-column prop="tenant_id" label="租户ID" width="180">
          <template #default="{ row }">
            {{ row.tenant_id?.slice(0, 8) }}...
          </template>
        </el-table-column>
        <el-table-column prop="plan_name" label="套餐" width="120" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">
            <span class="amount">¥{{ formatMoney(row.amount) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="110">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">
              {{ statusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="payment_method" label="支付方式" width="120">
          <template #default="{ row }">
            {{ row.payment_method || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="支付时间" width="170">
          <template #default="{ row }">
            {{ row.paid_at ? formatTime(row.paid_at) : '-' }}
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
        <el-pagination
          v-model:current-page="invoicePage"
          v-model:page-size="invoicePageSize"
          :total="invoiceTotal"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadInvoices"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { adminGetSubscriptions, adminGetPlans, adminGetInvoices, type Subscription, type Plan, type Invoice } from '@/api/billing'

const subscriptions = ref<Subscription[]>([])
const subscriptionTotal = ref(0)
const subscriptionLoading = ref(false)
const subPage = ref(1)
const subPageSize = ref(10)

const plans = ref<Plan[]>([])

const invoices = ref<Invoice[]>([])
const invoiceTotal = ref(0)
const invoiceLoading = ref(false)
const invoicePage = ref(1)
const invoicePageSize = ref(10)
const filterStatus = ref('')

function subStatusTagType(status: string) {
  const map: Record<string, string> = { active: 'success', expired: 'warning', cancelled: 'info' }
  return map[status] || 'info'
}
function subStatusLabel(status: string) {
  const map: Record<string, string> = { active: '正常', expired: '已过期', cancelled: '已取消' }
  return map[status] || '未知'
}
function statusTagType(status: string) {
  const map: Record<string, string> = { pending: 'warning', paid: 'success', failed: 'danger', cancelled: 'info' }
  return map[status] || 'info'
}
function statusLabel(status: string) {
  const map: Record<string, string> = { pending: '待支付', paid: '已支付', failed: '支付失败', cancelled: '已取消' }
  return map[status] || status
}
function formatMoney(amount: number) {
  return amount.toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',')
}
function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

async function loadSubscriptions() {
  subscriptionLoading.value = true
  try {
    const res = await adminGetSubscriptions({
      page: subPage.value,
      page_size: subPageSize.value
    })
    subscriptions.value = res.items || []
    subscriptionTotal.value = res.total || 0
  } catch {
    subscriptions.value = []
    subscriptionTotal.value = 0
  } finally {
    subscriptionLoading.value = false
  }
}

async function loadPlans() {
  try {
    const res = await adminGetPlans()
    plans.value = res.items || res || []
  } catch {
    plans.value = []
  }
}

async function loadInvoices() {
  invoiceLoading.value = true
  try {
    const res = await adminGetInvoices({
      page: invoicePage.value,
      page_size: invoicePageSize.value,
      status: filterStatus.value || undefined
    })
    invoices.value = res.items || []
    invoiceTotal.value = res.total || 0
  } catch {
    invoices.value = []
    invoiceTotal.value = 0
  } finally {
    invoiceLoading.value = false
  }
}

onMounted(() => {
  loadSubscriptions()
  loadPlans()
  loadInvoices()
})
</script>

<style scoped lang="scss">
.stats-row {
  margin-bottom: 20px;
}

.stat-info {
  text-align: center;

  .stat-value {
    font-size: 32px;
    font-weight: 700;
    color: #3b82f6;
    margin-bottom: 4px;
  }

  .stat-label {
    font-size: 14px;
    color: #64748b;
  }
}

.plan-card {
  text-align: center;

  h4 {
    margin: 0 0 8px;
    font-size: 18px;
  }

  .plan-card-price {
    font-size: 24px;
    font-weight: 700;
    color: #3b82f6;
    margin-bottom: 8px;
  }

  .plan-card-desc {
    color: #64748b;
    font-size: 14px;
    margin-bottom: 12px;
  }

  .plan-card-features {
    list-style: none;
    padding: 0;
    margin: 0 0 12px;
    text-align: left;

    li {
      padding: 4px 0;
      font-size: 13px;
      color: #475569;

      &::before {
        content: '✓ ';
        color: #10b981;
        font-weight: bold;
      }
    }
  }
}

.popular-plan {
  border: 2px solid #3b82f6;
}

.amount {
  font-weight: 600;
  color: #3b82f6;
}
</style>
