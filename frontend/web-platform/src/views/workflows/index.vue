<template>
  <div class="admin-workflows">
    <div class="page-header">
      <h2>工作流管理</h2>
      <div class="header-actions">
        <button class="btn btn-primary" @click="refresh">刷新</button>
      </div>
    </div>

    <div class="stats-cards">
      <div class="stat-card">
        <div class="stat-value">{{ stats.total }}</div>
        <div class="stat-label">总工作流</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.public }}</div>
        <div class="stat-label">公开模板</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ stats.private }}</div>
        <div class="stat-label">私有工作流</div>
      </div>
    </div>

    <div class="search-bar">
      <input 
        type="text" 
        v-model="searchQuery" 
        placeholder="搜索工作流..."
        class="search-input"
      />
      <select v-model="tenantFilter" class="filter-select">
        <option value="">全部租户</option>
        <option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">
          {{ tenant.name }}
        </option>
      </select>
      <button class="btn btn-secondary" @click="loadWorkflows">搜索</button>
    </div>

    <table class="data-table">
      <thead>
        <tr>
          <th>ID</th>
          <th>名称</th>
          <th>租户</th>
          <th>节点数</th>
          <th>状态</th>
          <th>创建时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="workflow in workflows" :key="workflow.id">
          <td>{{ workflow.id }}</td>
          <td>{{ workflow.name }}</td>
          <td>{{ getTenantName(workflow.tenant_id) }}</td>
          <td>{{ workflow.nodes?.length || 0 }}</td>
          <td>
            <span :class="['status-tag', workflow.is_public ? 'public' : 'private']">
              {{ workflow.is_public ? '公开' : '私有' }}
            </span>
          </td>
          <td>{{ formatDate(workflow.created_at) }}</td>
          <td>
            <button class="btn btn-sm btn-primary" @click="viewWorkflow(workflow)">查看</button>
            <button class="btn btn-sm btn-danger" @click="deleteWorkflow(workflow)">删除</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div class="pagination" v-if="total > pageSize">
      <button 
        class="btn btn-sm" 
        :disabled="currentPage <= 1"
        @click="currentPage--; loadWorkflows()"
      >上一页</button>
      <span>{{ currentPage }} / {{ totalPages }}</span>
      <button 
        class="btn btn-sm" 
        :disabled="currentPage >= totalPages"
        @click="currentPage++; loadWorkflows()"
      >下一页</button>
    </div>

    <div v-if="showDetail" class="modal-overlay" @click="showDetail = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ selectedWorkflow?.name }}</h3>
          <button class="btn btn-sm btn-secondary" @click="showDetail = false">关闭</button>
        </div>
        <div class="modal-body">
          <div class="detail-section">
            <h4>基本信息</h4>
            <p><strong>ID:</strong> {{ selectedWorkflow?.id }}</p>
            <p><strong>描述:</strong> {{ selectedWorkflow?.description || '-' }}</p>
            <p><strong>租户:</strong> {{ getTenantName(selectedWorkflow?.tenant_id) }}</p>
            <p><strong>状态:</strong> {{ selectedWorkflow?.is_public ? '公开' : '私有' }}</p>
          </div>
          <div class="detail-section">
            <h4>节点配置</h4>
            <pre>{{ JSON.stringify(selectedWorkflow?.nodes, null, 2) }}</pre>
          </div>
          <div class="detail-section">
            <h4>连接关系</h4>
            <pre>{{ JSON.stringify(selectedWorkflow?.edges, null, 2) }}</pre>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import request from '@/api/request'

interface Workflow {
  id: string
  name: string
  description: string
  tenant_id: string
  nodes: any[]
  edges: any[]
  is_public: boolean
  created_at: string
}

interface Tenant {
  id: string
  name: string
}

const workflows = ref<Workflow[]>([])
const tenants = ref<Tenant[]>([])
const searchQuery = ref('')
const tenantFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showDetail = ref(false)
const selectedWorkflow = ref<Workflow | null>(null)

const stats = computed(() => ({
  total: workflows.value.length,
  public: workflows.value.filter(w => w.is_public).length,
  private: workflows.value.filter(w => !w.is_public).length
}))

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const loadWorkflows = async () => {
  try {
    const response = await request.get('/workflows', { params: { page: currentPage.value, page_size: pageSize.value } }) as { items: Workflow[]; total: number }
    workflows.value = response.items || []
    total.value = response.total || 0
  } catch (error) {
    console.error('加载工作流失败:', error)
  }
}

const loadTenants = async () => {
  try {
    const response = await request.get('/tenants') as { items: Tenant[] }
    tenants.value = response.items || []
  } catch (error) {
    console.error('加载租户失败:', error)
  }
}

const getTenantName = (tenantId: string | undefined) => {
  if (!tenantId) return '-'
  const tenant = tenants.value.find(t => t.id === tenantId)
  return tenant?.name || '-'
}

const viewWorkflow = (workflow: Workflow) => {
  selectedWorkflow.value = workflow
  showDetail.value = true
}

const deleteWorkflow = async (workflow: Workflow) => {
  if (!confirm(`确定删除工作流 "${workflow.name}" 吗？`)) return
  
  try {
    await request.delete(`/workflows/${workflow.id}`)
    loadWorkflows()
  } catch (error) {
    console.error('删除工作流失败:', error)
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const refresh = () => {
  currentPage.value = 1
  loadWorkflows()
}

onMounted(() => {
  loadWorkflows()
  loadTenants()
})
</script>

<style scoped>
.admin-workflows {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.stats-cards {
  display: flex;
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  flex: 1;
  background: white;
  padding: 20px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.stat-value {
  font-size: 28px;
  font-weight: 600;
  color: #409eff;
}

.stat-label {
  font-size: 14px;
  color: #999;
  margin-top: 8px;
}

.search-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 20px;
}

.search-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.filter-select {
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
  border-radius: 8px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.data-table th,
.data-table td {
  padding: 12px 16px;
  text-align: left;
  border-bottom: 1px solid #eee;
}

.data-table th {
  background: #f5f5f5;
  font-weight: 600;
}

.status-tag {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
}

.status-tag.public {
  background: #e8f5e9;
  color: #2e7d32;
}

.status-tag.private {
  background: #f3e5f5;
  color: #7b1fa2;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  margin-top: 20px;
}

.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  justify-content: center;
  align-items: center;
  z-index: 100;
}

.modal-content {
  background: white;
  border-radius: 8px;
  width: 800px;
  max-height: 80vh;
  overflow: hidden;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid #eee;
}

.modal-body {
  padding: 20px;
  overflow-y: auto;
  max-height: calc(80vh - 60px);
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section h4 {
  margin-bottom: 12px;
  font-size: 14px;
  color: #333;
}

.detail-section pre {
  padding: 12px;
  background: #f5f5f5;
  border-radius: 4px;
  font-size: 12px;
  max-height: 200px;
  overflow: auto;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
}

.btn-primary {
  background: #409eff;
  color: white;
}

.btn-secondary {
  background: #606266;
  color: white;
}

.btn-danger {
  background: #f56c6c;
  color: white;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>