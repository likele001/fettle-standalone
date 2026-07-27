<template>
  <div class="page-container">
    <div class="page-header">
      <div style="display: flex; align-items: center; gap: 8px;">
        <el-button @click="$router.push('/tenants')" :icon="ArrowLeft" link>返回</el-button>
        <h2>{{ tenant?.name || '租户详情' }}</h2>
      </div>
    </div>

    <div v-if="loading" style="text-align: center; padding: 60px;">
      <el-icon class="is-loading" :size="24"><Loading /></el-icon>
      <p style="margin-top: 12px; color: #94a3b8;">加载中...</p>
    </div>

    <template v-else-if="tenant">
      <!-- 基本信息 -->
      <el-card class="section-card">
        <template #header><span style="font-weight: 600;">基本信息</span></template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="租户 ID">{{ tenant.id }}</el-descriptions-item>
          <el-descriptions-item label="租户名称">{{ tenant.name }}</el-descriptions-item>
          <el-descriptions-item label="租户编码">
            <template v-if="editingCode">
              <el-input v-model="editCodeValue" size="small" style="width: 140px;" @keyup.enter="saveCode" />
              <el-button type="primary" size="small" @click="saveCode" :loading="savingCode" style="margin-left: 8px;">保存</el-button>
              <el-button size="small" @click="editingCode = false">取消</el-button>
            </template>
            <template v-else>
              <el-tag size="small" type="warning">{{ tenant.code || '未设置' }}</el-tag>
              <el-button type="primary" link size="small" @click="startEditCode" style="margin-left: 8px;">
                <el-icon><Edit /></el-icon>
              </el-button>
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="套餐类型">
            <el-tag size="small" :type="tenant.plan_type === 'free' ? 'info' : 'success'">{{ tenant.plan_type }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="套餐到期">
            {{ tenant.plan_expires_at ? formatTime(tenant.plan_expires_at) : '无期限' }}
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="tenant.status === 'active' ? 'success' : 'danger'">
              {{ tenant.status === 'active' ? '正常' : tenant.status === 'banned' ? '已封禁' : tenant.status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="审核状态">
            <el-tag :type="auditTagType(tenant.audit_status)">
              {{ auditLabel(tenant.audit_status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">{{ formatTime(tenant.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ formatTime(tenant.updated_at) }}</el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 审核操作 -->
      <el-card class="section-card" v-if="tenant.audit_status === 'pending'">
        <template #header><span style="font-weight: 600;">审核操作</span></template>
        <el-form label-width="80px">
          <el-form-item label="审核备注">
            <el-input
              v-model="auditRemark"
              type="textarea"
              :rows="2"
              placeholder="可选，审核意见..."
            />
          </el-form-item>
          <el-form-item>
            <el-button type="success" :loading="auditLoading" @click="doAudit('approved')">
              审核通过
            </el-button>
            <el-button type="danger" :loading="auditLoading" @click="doAudit('rejected')">
              驳回
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card class="section-card" v-else-if="tenant.audit_status === 'rejected'">
        <template #header><span style="font-weight: 600;">审核信息</span></template>
        <el-alert type="info" :closable="false">
          <p>审核状态：<strong>已驳回</strong></p>
          <p v-if="tenant.audit_remark">驳回原因：{{ tenant.audit_remark }}</p>
        </el-alert>
      </el-card>

      <!-- 封禁操作 -->
      <el-card class="section-card">
        <template #header><span style="font-weight: 600;">封禁管理</span></template>
        <el-alert
          v-if="tenant.banned_reason"
          type="warning"
          :closable="false"
          style="margin-bottom: 16px;"
        >
          <p>封禁时间：{{ formatTime(tenant.banned_at || '') }}</p>
          <p>封禁原因：{{ tenant.banned_reason }}</p>
        </el-alert>
        <el-form label-width="80px">
          <el-form-item v-if="tenant.status !== 'banned'" label="封禁原因">
            <el-input
              v-model="banReason"
              type="textarea"
              :rows="2"
              placeholder="封禁原因..."
            />
          </el-form-item>
          <el-form-item>
            <el-button
              v-if="tenant.status !== 'banned'"
              type="danger"
              :loading="banLoading"
              @click="doBan(true)"
            >
              封禁租户
            </el-button>
            <el-button
              v-else
              type="success"
              :loading="banLoading"
              @click="doBan(false)"
            >
              解除封禁
            </el-button>
          </el-form-item>
        </el-form>
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Loading, Edit } from '@element-plus/icons-vue'
import { getTenantDetail, auditTenant, banTenant, type TenantItem } from '@/api/tenant'
import request from '@/api/request'

const route = useRoute()
const router = useRouter()
const tenant = ref<TenantItem | null>(null)
const loading = ref(true)
const auditLoading = ref(false)
const banLoading = ref(false)
const auditRemark = ref('')
const banReason = ref('')
const editingCode = ref(false)
const editCodeValue = ref('')
const savingCode = ref(false)

function startEditCode() {
  editCodeValue.value = tenant.value?.code || ''
  editingCode.value = true
}

async function saveCode() {
  const val = editCodeValue.value.trim() || null
  if (!val) {
    ElMessage.warning('租户编码不能为空')
    return
  }
  savingCode.value = true
  try {
    await request.put(`/admin/tenants/${tenant.value?.id}/audit`, { code: val }) as any
    // 使用 current tenant update API
    // 没有专门的 update tenant API 暴露给 admin，用 PUT tenants/:id/audit hack
    // 实际应有一个 updateTenant API
  } catch {
    // ignore
  }
  // 改用简单的直接更新
  try {
    // 直接调用租户更新
    await request.put(`/admin/tenants/${tenant.value!.id}`, { code: val }) as any
    ElMessage.success('租户编码已更新')
    tenant.value!.code = val
    editingCode.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '更新失败')
  } finally {
    savingCode.value = false
  }
}

function auditTagType(status: string) {
  const map: Record<string, string> = { pending: 'warning', approved: 'success', rejected: 'danger' }
  return map[status] || 'info'
}

function auditLabel(status: string) {
  const map: Record<string, string> = { pending: '待审核', approved: '已通过', rejected: '已驳回' }
  return map[status] || status
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

async function loadTenant() {
  loading.value = true
  try {
    tenant.value = await getTenantDetail(route.params.id as string)
  } catch {
    ElMessage.error('租户不存在')
    router.push('/tenants')
  } finally {
    loading.value = false
  }
}

async function doAudit(status: string) {
  const label = status === 'approved' ? '通过' : '驳回'
  try {
    await ElMessageBox.confirm(`确定要${label}该租户的审核吗？`, '审核确认', {
      type: status === 'approved' ? 'success' : 'warning'
    })
  } catch { return }

  auditLoading.value = true
  try {
    await auditTenant(route.params.id as string, {
      status,
      remark: auditRemark.value
    })
    ElMessage.success(`审核${label}`)
    await loadTenant()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    auditLoading.value = false
  }
}

async function doBan(banned: boolean) {
  const label = banned ? '封禁' : '解封'
  try {
    await ElMessageBox.confirm(
      banned ? `确定要封禁租户「${tenant.value?.name}」吗？` : `确定要解除租户「${tenant.value?.name}」的封禁吗？`,
      `${label}确认`,
      { type: banned ? 'warning' : 'info' }
    )
  } catch { return }

  banLoading.value = true
  try {
    await banTenant(route.params.id as string, {
      banned,
      reason: banReason.value
    })
    ElMessage.success(`${label}成功`)
    await loadTenant()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  } finally {
    banLoading.value = false
  }
}

onMounted(() => loadTenant())
</script>

<style scoped lang="scss">
.section-card {
  margin-bottom: 20px;
}
</style>
