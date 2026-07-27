<template>
  <div class="billing-page">
    <div class="page-header">
      <h2>套餐与计费</h2>
    </div>

    <!-- 支付成功提示 -->
    <el-alert
      v-if="paymentSuccess"
      title="支付成功"
      description="您的套餐已升级生效，感谢您的订阅！"
      type="success"
      show-icon
      closable
      style="margin-bottom: 20px"
      @close="paymentSuccess = false"
    />

    <!-- 当前套餐 & 配额 -->
    <el-row :gutter="20" class="top-section">
      <el-col :span="8">
        <el-card class="current-plan-card" shadow="hover">
          <div class="plan-badge">当前套餐</div>
          <h3 class="plan-name">{{ currentPlan?.name || '免费版' }}</h3>
          <div class="plan-price-row">
            <span class="price">¥{{ currentPlan?.price || 0 }}</span>
            <span class="period">/{{ currentPlan?.period || '月' }}</span>
          </div>
          <div class="plan-status">
            <el-tag :type="subStatusType" size="small">{{ subStatusText }}</el-tag>
            <span class="expire-date" v-if="subscription?.end_date">
              到期：{{ formatDate(subscription.end_date) }}
            </span>
          </div>
          <el-button
            v-if="currentPlan?.id !== 'enterprise'"
            type="primary"
            style="width: 100%; margin-top: 16px"
            @click="showUpgradeDialog"
          >
            升级套餐
          </el-button>
        </el-card>
      </el-col>

      <el-col :span="16">
        <el-card>
          <template #header>
            <span style="font-weight: 600">资源配额</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="8" v-for="item in quotaItems" :key="item.label">
              <div class="quota-item">
                <div class="quota-header">
                  <span class="quota-label">{{ item.label }}</span>
                  <span class="quota-value">{{ item.used }} / {{ item.limit }}</span>
                </div>
                <el-progress
                  :percentage="item.percent"
                  :status="item.percent >= 90 ? 'exception' : item.percent >= 70 ? 'warning' : ''"
                  :stroke-width="10"
                />
                <div class="quota-remaining">剩余 {{ item.remaining }}</div>
              </div>
            </el-col>
          </el-row>
        </el-card>
      </el-col>
    </el-row>

    <!-- 可用套餐 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <span style="font-weight: 600">可用套餐</span>
      </template>
      <el-row :gutter="20">
        <el-col :span="8" v-for="plan in plans" :key="plan.id">
          <el-card
            class="plan-card"
            :class="{ current: plan.id === currentPlan?.id, recommended: plan.id === 'pro' }"
            shadow="hover"
          >
            <div v-if="plan.id === 'pro'" class="recommend-badge">推荐</div>
            <div class="plan-header">
              <h3>{{ plan.name }}</h3>
              <el-tag v-if="plan.id === currentPlan?.id" type="success" size="small">当前</el-tag>
            </div>
            <div class="plan-price-row">
              <span class="price">¥{{ plan.price }}</span>
              <span class="period">/{{ plan.period }}</span>
            </div>
            <ul class="plan-features">
              <li v-for="(f, i) in plan.features" :key="i">
                <el-icon color="#67C23A"><CircleCheck /></el-icon>
                {{ f }}
              </li>
            </ul>
            <div v-if="plan.allowed_industries?.length" class="plan-restrictions">
              <div class="restriction-label">允许行业</div>
              <div class="restriction-tags">
                <span v-for="industry in plan.allowed_industries" :key="industry" class="restriction-tag">
                  {{ industryLabels[industry] || industry }}
                </span>
              </div>
            </div>
            <div v-if="plan.allowed_features?.length" class="plan-restrictions">
              <div class="restriction-label">允许功能</div>
              <div class="restriction-tags">
                <span v-for="feature in plan.allowed_features" :key="feature" class="restriction-tag">
                  {{ featureLabels[feature] || feature }}
                </span>
              </div>
            </div>
            <el-button
              v-if="plan.id !== currentPlan?.id"
              :type="plan.id === 'pro' ? 'primary' : 'default'"
              style="width: 100%"
              :loading="paying === plan.id"
              @click="selectPlan(plan)"
            >
              {{ plan.price > (currentPlan?.price || 0) ? '升级' : '降级' }}
            </el-button>
            <el-button v-else disabled style="width: 100%">当前套餐</el-button>
          </el-card>
        </el-col>
      </el-row>
    </el-card>

    <!-- 支付订单记录 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <div style="display: flex; justify-content: space-between; align-items: center">
          <span style="font-weight: 600">支付记录</span>
          <div style="display: flex; gap: 12px; align-items: center">
            <el-button type="success" size="small" @click="showBillingExport">
              <el-icon><Download /></el-icon> 导出记录
            </el-button>
            <el-radio-group v-model="orderFilter" size="small" @change="loadOrders">
              <el-radio-button value="">全部</el-radio-button>
              <el-radio-button value="paid">已支付</el-radio-button>
              <el-radio-button value="pending">待支付</el-radio-button>
              <el-radio-button value="failed">失败</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>
      <el-table :data="orders" stripe v-loading="ordersLoading">
        <el-table-column prop="trade_order_id" label="订单号" width="180">
          <template #default="{ row }">{{ row.trade_order_id?.slice(0, 16) }}...</template>
        </el-table-column>
        <el-table-column prop="plan_name" label="套餐" width="120" />
        <el-table-column label="金额" width="120">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="orderStatusType(row.status)" size="small">
              {{ orderStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="支付时间" width="160">
          <template #default="{ row }">{{ row.paid_at ? formatDate(row.paid_at) : '-' }}</template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button
              v-if="row.status === 'pending' && row.payment_url"
              size="small"
              type="primary"
              @click="gotoPay(row)"
            >
              去支付
            </el-button>
            <span v-else style="color: #999; font-size: 13px">-</span>
          </template>
        </el-table-column>
      </el-table>
      <el-pagination
        v-if="orderTotal > orderPageSize"
        style="margin-top: 16px; text-align: right"
        layout="prev, pager, next"
        :total="orderTotal"
        :page-size="orderPageSize"
        v-model:current-page="orderPage"
        @current-change="loadOrders"
      />
    </el-card>

    <!-- 升级确认对话框 -->
    <el-dialog v-model="upgradeVisible" title="确认订阅套餐" width="420px">
      <div class="upgrade-confirm">
        <p>确定要订阅 <strong>{{ selectedPlan?.name }}</strong> 套餐吗？</p>
        <div class="upgrade-price">
          <span>费用：</span>
          <span class="price">¥{{ selectedPlan?.price }}</span>
          <span class="period">/{{ selectedPlan?.period }}</span>
        </div>
        <p v-if="selectedPlan && (selectedPlan.price || 0) > 0" style="color: #999; font-size: 13px; margin-top: 12px">
          点击确认后将跳转到支付页面完成付款
        </p>
      </div>
      <template #footer>
        <el-button @click="upgradeVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmUpgrade" :loading="upgrading">
          {{ (selectedPlan?.price || 0) > 0 ? '去支付' : '确认订阅' }}
        </el-button>
      </template>
    </el-dialog>
  
    <!-- 导出记录对话框 -->
    <el-dialog v-model="billingExportVisible" title="导出计费记录" width="420px">
      <el-form label-width="80px">
        <el-form-item label="日期范围">
          <el-date-picker
            v-model="billingExportRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
            style="width: 100%"
          />
        </el-form-item>
        <el-form-item>
          <span style="color: #999; font-size: 13px">不选择日期则导出全部记录</span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="billingExportVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBillingExport" :loading="billingExporting">导出</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { CircleCheck } from '@element-plus/icons-vue'
import {
  getPlans,
  getSubscription,
  subscribe,
  createPayment,
  getQuota,
  getPaymentOrders,
  type Plan,
  type Subscription,
  type Quota
} from '@/api/billing'
import { exportBillingRecords, downloadBlob } from '@/api/export'

const route = useRoute()
const router = useRouter()

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

const featureLabels: Record<string, string> = {
  agent_management: '智能体管理',
  knowledge_base: '知识库管理',
  skill_market: '技能市场',
  conversation: '对话管理',
  analytics: '统计分析',
  channel_config: '渠道配置',
  ai_config: 'AI配置',
  workflow: '自定义工作流',
  api_access: 'API访问',
  team_collaboration: '团队协作'
}

const currentPlan = ref<Plan | null>(null)
const subscription = ref<Subscription | null>(null)
const plans = ref<Plan[]>([])
const quota = ref<Quota | null>(null)

// 支付订单
const orders = ref<any[]>([])
const orderPage = ref(1)
const orderPageSize = ref(10)
const orderTotal = ref(0)
const orderFilter = ref('')
const ordersLoading = ref(false)

// 升级 & 支付
const upgradeVisible = ref(false)
const selectedPlan = ref<Plan | null>(null)
const upgrading = ref(false)
const paying = ref<string | null>(null)
const paymentSuccess = ref(false)

const subStatusType = computed(() => {
  const map: Record<string, string> = { active: 'success', expired: 'danger', cancelled: 'info' }
  return map[subscription.value?.status || ''] || 'info'
})

const subStatusText = computed(() => {
  const map: Record<string, string> = { active: '生效中', expired: '已过期', cancelled: '已取消' }
  return map[subscription.value?.status || ''] || '未知'
})

const quotaItems = computed(() => {
  if (!quota.value || !currentPlan.value) return []
  const items = [
    { label: '智能体数量', used: 0, limit: currentPlan.value.max_agents || 1, remaining: 0 },
    { label: '消息用量', used: quota.value.used || 0, limit: quota.value.limit || currentPlan.value.max_messages || 100, remaining: 0 },
    { label: '存储空间', used: 0, limit: 1024, remaining: 0 }
  ]
  items.forEach(item => {
    item.remaining = Math.max(0, item.limit - item.used)
  })
  return items.map(item => ({
    ...item,
    percent: item.limit > 0 ? Math.round((item.used / item.limit) * 100) : 0
  }))
})

const orderStatusType = (status: string) => {
  const map: Record<string, string> = { paid: 'success', pending: 'warning', failed: 'danger', cancelled: 'info' }
  return map[status] || 'info'
}

const orderStatusText = (status: string) => {
  const map: Record<string, string> = { paid: '已支付', pending: '待支付', failed: '支付失败', cancelled: '已取消' }
  return map[status] || status
}

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' })
}

const loadPlans = async () => {
  try {
    const res = await getPlans()
    const items = res?.items || []
    plans.value = items.map((p: any) => {
      try {
        p.features = typeof p.features === 'string' ? JSON.parse(p.features) : p.features
      } catch { p.features = [] }
      // 兼容后端字段名
      if (p.price === undefined && p.price_monthly !== undefined) {
        p.price = p.price_monthly
      }
      if (p.period === undefined) {
        p.period = '月'
      }
      return p
    })
  } catch {
    plans.value = []
  }
}

const loadSubscription = async () => {
  try {
    const res = await getSubscription()
    const data = res?.data || res
    subscription.value = data
    if (data?.plan_id) {
      currentPlan.value = plans.value.find(p => p.id === data.plan_id) || plans.value[0] || null
    } else {
      currentPlan.value = plans.value[0] || null
    }
  } catch {
    currentPlan.value = plans.value[0] || null
  }
}

const loadQuota = async () => {
  try {
    const res = await getQuota()
    quota.value = res?.data || res
  } catch {
    quota.value = null
  }
}

const loadOrders = async () => {
  ordersLoading.value = true
  try {
    const res = await getPaymentOrders({ page: orderPage.value, page_size: orderPageSize.value })
    const data = res?.data || res || {}
    orders.value = Array.isArray(data) ? data : (data.items || data.records || [])
    orderTotal.value = data.total || orders.value.length || 0
  } catch {
    orders.value = []
    orderTotal.value = 0
  } finally {
    ordersLoading.value = false
  }
}

const showUpgradeDialog = () => {
  const nextPlan = plans.value.find(p => p.price > (currentPlan.value?.price || 0))
  if (nextPlan) {
    selectedPlan.value = nextPlan
    upgradeVisible.value = true
  }
}

const selectPlan = (plan: Plan) => {
  selectedPlan.value = plan
  upgradeVisible.value = true
}

const confirmUpgrade = async () => {
  if (!selectedPlan.value) return
  upgrading.value = true

  try {
    if (selectedPlan.value.price <= 0) {
      // 免费套餐直接订阅
      await subscribe(selectedPlan.value.id)
      ElMessage.success('订阅成功')
      upgradeVisible.value = false
      await loadSubscription()
      await loadQuota()
    } else {
      // 付费套餐 → 创建支付订单
      const res = await createPayment(selectedPlan.value.id)
      const data = res?.data || res
      if (data?.payment_url) {
        upgradeVisible.value = false
        // 跳转到虎皮椒支付页面
        window.location.href = data.payment_url
      } else {
        ElMessage.error('创建支付订单失败：未返回支付链接')
      }
    }
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    upgrading.value = false
  }
}

// 待支付订单 → 跳转支付
const gotoPay = (order: any) => {
  if (order.payment_url) {
    window.location.href = order.payment_url
  } else {
    ElMessage.warning('该订单暂无支付链接')
  }
}

// 导出相关
const billingExportVisible = ref(false)
const billingExportRange = ref<string[]>([])
const billingExporting = ref(false)

const showBillingExport = () => {
  billingExportRange.value = []
  billingExportVisible.value = true
}

const handleBillingExport = async () => {
  billingExporting.value = true
  try {
    const params: { start_date?: string; end_date?: string } = {}
    if (billingExportRange.value && billingExportRange.value.length >= 2) {
      params.start_date = billingExportRange.value[0]
      params.end_date = billingExportRange.value[1]
    }
    const blob = await exportBillingRecords(params) as any
    downloadBlob(blob, 'billing_records_export.csv')
    ElMessage.success('导出成功')
    billingExportVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '导出失败')
  } finally {
    billingExporting.value = false
  }
}

onMounted(async () => {
  // 检查支付回调
  if (route.query.payment === 'success') {
    paymentSuccess.value = true
    // 清理 URL 参数
    router.replace({ query: { ...route.query, payment: undefined } })
  }

  await loadPlans()
  await loadSubscription()
  await loadQuota()
  await loadOrders()
})
</script>

<style scoped lang="scss">
.billing-page {
  padding: 20px;
}

.page-header {
  margin-bottom: 24px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.top-section {
  margin-bottom: 0;
}

.current-plan-card {
  text-align: center;
  position: relative;

  .plan-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    background: #409eff;
    color: #fff;
    font-size: 12px;
    padding: 2px 8px;
    border-radius: 4px;
  }

  .plan-name {
    margin: 16px 0 8px;
    font-size: 22px;
  }

  .plan-price-row {
    margin-bottom: 12px;

    .price {
      font-size: 36px;
      font-weight: bold;
      color: #409eff;
    }

    .period {
      color: #999;
      font-size: 14px;
    }
  }

  .plan-status {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;

    .expire-date {
      font-size: 13px;
      color: #999;
    }
  }
}

.quota-item {
  .quota-header {
    display: flex;
    justify-content: space-between;
    margin-bottom: 8px;

    .quota-label {
      font-size: 14px;
      color: #333;
    }

    .quota-value {
      font-size: 14px;
      font-weight: 600;
      color: #333;
    }
  }

  .quota-remaining {
    font-size: 12px;
    color: #999;
    margin-top: 4px;
  }
}

.plan-card {
  text-align: center;
  transition: transform 0.2s;
  position: relative;

  &:hover {
    transform: translateY(-4px);
  }

  &.current {
    border-color: #409eff;
    border-width: 2px;
  }

  &.recommended {
    border-color: #e6a23c;
    border-width: 2px;
  }

  .recommend-badge {
    position: absolute;
    top: -1px;
    right: 16px;
    background: #e6a23c;
    color: #fff;
    font-size: 12px;
    padding: 2px 10px;
    border-radius: 0 0 4px 4px;
  }

  .plan-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h3 {
      margin: 0;
      font-size: 18px;
    }
  }

  .plan-price-row {
    margin-bottom: 24px;

    .price {
      font-size: 36px;
      font-weight: bold;
      color: #409eff;
    }

    .period {
      color: #999;
      font-size: 14px;
    }
  }

  .plan-features {
    list-style: none;
    padding: 0;
    margin: 0 0 16px 0;
    text-align: left;

    li {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 8px 0;
      color: #666;
      font-size: 14px;
    }
  }

  .plan-restrictions {
    margin-bottom: 12px;
    padding: 10px;
    background: #faf5ff;
    border-radius: 6px;
    text-align: left;

    .restriction-label {
      font-size: 11px;
      color: #7c3aed;
      font-weight: 500;
      margin-bottom: 6px;
    }

    .restriction-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .restriction-tag {
      background: #ede9fe;
      color: #6d28d9;
      padding: 2px 8px;
      border-radius: 4px;
      font-size: 11px;
    }
  }
}

.upgrade-confirm {
  p {
    font-size: 15px;
    margin-bottom: 16px;
  }

  .upgrade-price {
    background: #f5f7fa;
    padding: 12px 16px;
    border-radius: 8px;
    font-size: 15px;

    .price {
      font-size: 24px;
      font-weight: bold;
      color: #409eff;
    }

    .period {
      color: #999;
    }
  }
}
</style>
