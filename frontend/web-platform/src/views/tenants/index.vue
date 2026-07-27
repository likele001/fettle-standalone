<template>
  <div class="page-container">
    <div class="page-header">
      <h2>租户管理</h2>
      <p>管理所有平台的租户，包括审核、封禁等操作</p>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-input
        v-model="searchText"
        placeholder="搜索租户名称或编码..."
        :prefix-icon="Search"
        clearable
        style="width: 280px"
        @clear="loadTenants"
        @keyup.enter="loadTenants"
      />
      <el-select v-model="filterStatus" placeholder="状态" clearable style="width: 130px" @change="loadTenants">
        <el-option label="全部" value="" />
        <el-option label="活跃" value="active" />
        <el-option label="已封禁" value="banned" />
      </el-select>
      <el-select v-model="filterAudit" placeholder="审核状态" clearable style="width: 130px" @change="loadTenants">
        <el-option label="全部" value="" />
        <el-option label="待审核" value="pending" />
        <el-option label="已通过" value="approved" />
        <el-option label="已驳回" value="rejected" />
      </el-select>
      <el-button type="primary" @click="loadTenants">
        <el-icon><Search /></el-icon>
        查询
      </el-button>
    </div>

    <!-- 表格 -->
    <el-card>
      <el-table :data="tenants" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="code" label="租户编码" width="100">
          <template #default="{ row }">
            <el-tag size="small" type="warning">{{ row.code || '-' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="租户名称" min-width="160">
          <template #default="{ row }">
            <el-link type="primary" @click="$router.push(`/tenants/${row.id}`)">{{ row.name }}</el-link>
          </template>
        </el-table-column>
        <el-table-column prop="plan_type" label="套餐" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.plan_type === 'free' ? 'info' : 'success'">{{ row.plan_type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核状态" width="110">
          <template #default="{ row }">
            <el-tag size="small" :type="auditTagType(row.audit_status)">
              {{ auditLabel(row.audit_status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="90">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'active' ? 'success' : 'danger'">
              {{ row.status === 'active' ? '正常' : '已封禁' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="注册时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="$router.push(`/tenants/${row.id}`)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; display: flex; justify-content: flex-end;">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadTenants"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Search } from '@element-plus/icons-vue'
import { getTenants, type TenantItem } from '@/api/tenant'

const tenants = ref<TenantItem[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const searchText = ref('')
const filterStatus = ref('')
const filterAudit = ref('')

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

async function loadTenants() {
  loading.value = true
  try {
    const res = await getTenants({
      page: page.value,
      page_size: pageSize.value,
      search: searchText.value || undefined,
      status: filterStatus.value || undefined,
      audit_status: filterAudit.value || undefined
    })
    tenants.value = res.items
    total.value = res.total
  } catch {
    tenants.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

onMounted(() => loadTenants())
</script>
