<template>
  <div class="page-container">
    <div class="page-header">
      <h2>支付配置</h2>
      <p>配置虎皮椒支付接口</p>
    </div>

    <el-card>
      <el-form :model="form" label-width="140px" style="max-width: 600px;">
        <el-form-item label="APP ID">
          <el-input v-model="form.app_id" placeholder="虎皮椒 APPID" />
        </el-form-item>

        <el-form-item label="APP Secret">
          <el-input v-model="form.app_secret" type="password" show-password placeholder="虎皮椒 Secret" />
        </el-form-item>

        <el-form-item label="支付网关">
          <el-input v-model="form.gateway_url" placeholder="https://api.xunhupay.com/payment/do.html" />
        </el-form-item>

        <el-form-item label="回调地址">
          <el-input v-model="form.notify_url" placeholder="https://yourdomain.com/api/v1/billing/payment/notify" />
          <div class="form-tip">虎皮椒支付成功后会回调此地址</div>
        </el-form-item>

        <el-form-item label="跳转地址">
          <el-input v-model="form.return_url" placeholder="https://yourdomain.com/billing?payment=success" />
          <div class="form-tip">用户支付成功后跳转的页面</div>
        </el-form-item>

        <el-form-item label="启用支付">
          <el-switch v-model="form.enabled" />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" @click="saveConfig" :loading="saving">保存配置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card style="margin-top: 20px;">
      <template #header>
        <span style="font-weight: 600;">支付订单</span>
      </template>

      <el-table :data="orders" v-loading="loading" stripe>
        <el-table-column prop="trade_order_id" label="订单号" width="200">
          <template #default="{ row }">{{ row.trade_order_id.slice(0, 16) }}...</template>
        </el-table-column>
        <el-table-column prop="plan_name" label="套餐" width="120" />
        <el-table-column label="金额" width="100">
          <template #default="{ row }">¥{{ row.amount }}</template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="statusType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="支付时间" width="170">
          <template #default="{ row }">{{ row.paid_at ? formatTime(row.paid_at) : '-' }}</template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadOrders"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import request from '@/api/request'
import { ElMessage } from 'element-plus'

interface PaymentConfig {
  app_id: string
  app_secret: string
  gateway_url: string
  notify_url: string
  return_url: string
  enabled: boolean
}

interface PaymentOrder {
  trade_order_id: string
  plan_name: string
  amount: number
  status: string
  created_at: string
  paid_at: string | null
}

const form = ref<PaymentConfig>({
  app_id: '',
  app_secret: '',
  gateway_url: 'https://api.xunhupay.com/payment/do.html',
  notify_url: '',
  return_url: '',
  enabled: false
})

const orders = ref<PaymentOrder[]>([])
const loading = ref(false)
const saving = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

async function loadConfig() {
  try {
    const res = await request.get('/admin/payment/config')
    if (res.data) {
      form.value = res.data
    }
  } catch {
    // ignore
  }
}

async function saveConfig() {
  saving.value = true
  try {
    await request.put('/admin/payment/config', form.value)
    ElMessage.success('保存成功')
    loadConfig()
  } catch (err: any) {
    ElMessage.error(err.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function loadOrders() {
  loading.value = true
  try {
    const res: any = await request.get('/admin/payment/orders', {
      params: { page: page.value, page_size: pageSize.value }
    })
    orders.value = res.items || res.data || []
    total.value = res.total || 0
  } catch {
    orders.value = []
  } finally {
    loading.value = false
  }
}

function statusType(status: string) {
  const map: Record<string, string> = { pending: 'warning', paid: 'success', failed: 'danger', cancelled: 'info' }
  return map[status] || 'info'
}

function statusLabel(status: string) {
  const map: Record<string, string> = { pending: '待支付', paid: '已支付', failed: '失败', cancelled: '已取消' }
  return map[status] || status
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(() => {
  loadConfig()
  loadOrders()
})
</script>

<style scoped>
.form-tip {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}
</style>
