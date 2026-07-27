<template>
  <div class="workflow-list">
    <div class="page-header">
      <h2>工作流管理</h2>
      <button class="btn btn-primary" @click="createWorkflow">创建工作流</button>
    </div>

    <div class="search-bar">
      <input 
        type="text" 
        v-model="searchQuery" 
        placeholder="搜索工作流名称..."
        class="search-input"
      />
      <select v-model="statusFilter" class="filter-select">
        <option value="">全部状态</option>
        <option value="active">启用</option>
        <option value="disabled">禁用</option>
      </select>
      <button class="btn btn-secondary" @click="loadWorkflows">搜索</button>
    </div>

    <table class="data-table">
      <thead>
        <tr>
          <th>名称</th>
          <th>描述</th>
          <th>节点数量</th>
          <th>状态</th>
          <th>创建时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="workflow in workflows" :key="workflow.id">
          <td>{{ workflow.name }}</td>
          <td>{{ workflow.description || '-' }}</td>
          <td>{{ workflow.nodes?.length || 0 }}</td>
          <td>
            <span :class="['status-tag', workflow.is_public ? 'public' : 'private']">
              {{ workflow.is_public ? '公开' : '私有' }}
            </span>
          </td>
          <td>{{ formatDate(workflow.created_at) }}</td>
          <td>
            <button class="btn btn-sm btn-primary" @click="editWorkflow(workflow)">编辑</button>
            <button class="btn btn-sm btn-success" @click="executeWorkflow(workflow)">测试</button>
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

    <WorkflowEditor 
      v-if="showEditor" 
      :workflow="editingWorkflow"
      @close="showEditor = false; loadWorkflows()"
    />

    <WorkflowTest 
      v-if="showTest" 
      :workflow="testingWorkflow"
      @close="showTest = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { workflowApi, type Workflow } from '@/api/workflow'
import WorkflowEditor from './editor.vue'
import WorkflowTest from './test.vue'

const workflows = ref<Workflow[]>([])
const searchQuery = ref('')
const statusFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const showEditor = ref(false)
const showTest = ref(false)
const editingWorkflow = ref<Workflow | null>(null)
const testingWorkflow = ref<Workflow | null>(null)

const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

const loadWorkflows = async () => {
  try {
    const response = await workflowApi.listWorkflows(currentPage.value, pageSize.value) as unknown as { items: Workflow[]; total: number }
    workflows.value = response.items || []
    total.value = response.total || 0
  } catch (error) {
    console.error('加载工作流失败:', error)
  }
}

const createWorkflow = () => {
  editingWorkflow.value = null
  showEditor.value = true
}

const editWorkflow = (workflow: Workflow) => {
  editingWorkflow.value = workflow
  showEditor.value = true
}

const executeWorkflow = (workflow: Workflow) => {
  testingWorkflow.value = workflow
  showTest.value = true
}

const deleteWorkflow = async (workflow: Workflow) => {
  if (!confirm(`确定删除工作流 "${workflow.name}" 吗？`)) return
  
  try {
    await workflowApi.deleteWorkflow(workflow.id)
    loadWorkflows()
  } catch (error) {
    console.error('删除工作流失败:', error)
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadWorkflows()
})
</script>

<style scoped>
.workflow-list {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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

.btn-success {
  background: #67c23a;
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