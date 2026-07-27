<template>
  <div class="mcp-servers">
    <div class="page-header">
      <h2>MCP 服务器管理</h2>
      <button class="btn btn-primary" @click="showAddModal = true">添加 MCP 服务器</button>
    </div>

    <div class="stats-cards">
      <div class="stat-card">
        <div class="stat-value">{{ servers.length }}</div>
        <div class="stat-label">MCP 服务器</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ servers.filter(s => s.status === 'online').length }}</div>
        <div class="stat-label">在线</div>
      </div>
      <div class="stat-card">
        <div class="stat-value">{{ totalTools }}</div>
        <div class="stat-label">已注册工具</div>
      </div>
    </div>

    <table class="data-table">
      <thead>
        <tr>
          <th>名称</th>
          <th>地址</th>
          <th>工具数量</th>
          <th>状态</th>
          <th>租户限制</th>
          <th>创建时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="server in servers" :key="server.id">
          <td>{{ server.name }}</td>
          <td>{{ server.url }}</td>
          <td>{{ server.tool_count || 0 }}</td>
          <td>
            <span :class="['status-tag', server.status]">
              {{ server.status === 'online' ? '在线' : server.status === 'offline' ? '离线' : '未知' }}
            </span>
          </td>
          <td>{{ server.allowed_tenants?.length === 0 ? '全部租户' : `${server.allowed_tenants?.length}个租户` }}</td>
          <td>{{ formatDate(server.created_at) }}</td>
          <td>
            <button class="btn btn-sm" @click="testServer(server)">测试连接</button>
            <button class="btn btn-sm btn-primary" @click="syncTools(server)">同步工具</button>
            <button class="btn btn-sm btn-danger" @click="deleteServer(server)">删除</button>
          </td>
        </tr>
      </tbody>
    </table>

    <div v-if="showAddModal" class="modal-overlay" @click="showAddModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ editingServer ? '编辑 MCP 服务器' : '添加 MCP 服务器' }}</h3>
          <button class="btn btn-sm btn-secondary" @click="showAddModal = false">关闭</button>
        </div>
        <div class="modal-body">
          <form @submit.prevent="saveServer">
            <div class="form-group">
              <label>名称</label>
              <input 
                type="text" 
                v-model="form.name" 
                required
                class="form-input"
                placeholder="MCP 服务器名称"
              />
            </div>
            <div class="form-group">
              <label>地址</label>
              <input 
                type="text" 
                v-model="form.url" 
                required
                class="form-input"
                placeholder="http://localhost:8000"
              />
            </div>
            <div class="form-group">
              <label>API 密钥（可选）</label>
              <input 
                type="text" 
                v-model="form.api_key" 
                class="form-input"
                placeholder="MCP 服务器 API 密钥"
              />
            </div>
            <div class="form-group">
              <label>描述（可选）</label>
              <textarea 
                v-model="form.description" 
                class="form-textarea"
                placeholder="服务器描述"
              ></textarea>
            </div>
            <div class="form-group">
              <label>允许的租户（空表示全部租户）</label>
              <select v-model="form.allowed_tenants" multiple class="form-select">
                <option v-for="tenant in tenants" :key="tenant.id" :value="tenant.id">
                  {{ tenant.name }}
                </option>
              </select>
            </div>
            <div class="form-actions">
              <button type="submit" class="btn btn-primary">保存</button>
              <button type="button" class="btn btn-secondary" @click="showAddModal = false">取消</button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <div v-if="showToolModal" class="modal-overlay" @click="showToolModal = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>{{ currentServer?.name }} - 已注册工具</h3>
          <button class="btn btn-sm btn-secondary" @click="showToolModal = false">关闭</button>
        </div>
        <div class="modal-body">
          <div class="tools-list">
            <div v-for="tool in currentTools" :key="tool.name" class="tool-card">
              <h4>{{ tool.name }}</h4>
              <p>{{ tool.description }}</p>
              <div class="tool-params">
                <span v-for="(param, key) in tool.parameters" :key="key" class="param-tag">
                  {{ key }}: {{ param.type }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

interface MCPServer {
  id: string
  name: string
  url: string
  api_key: string
  description: string
  status: string
  tool_count: number
  allowed_tenants: string[]
  created_at: string
}

interface Tenant {
  id: string
  name: string
}

interface MCPTool {
  name: string
  description: string
  parameters: Record<string, { type: string; description: string }>
}

const servers = ref<MCPServer[]>([])
const tenants = ref<Tenant[]>([])
const showAddModal = ref(false)
const showToolModal = ref(false)
const editingServer = ref<MCPServer | null>(null)
const currentServer = ref<MCPServer | null>(null)
const currentTools = ref<MCPTool[]>([])

const form = ref({
  name: '',
  url: '',
  api_key: '',
  description: '',
  allowed_tenants: [] as string[]
})

const totalTools = computed(() => servers.value.reduce((sum, s) => sum + (s.tool_count || 0), 0))

const loadServers = async () => {
  try {
    const response = await fetch('/api/mcp-servers')
    const data = await response.json()
    servers.value = data.items || []
  } catch (error) {
    console.error('加载 MCP 服务器失败:', error)
  }
}

const loadTenants = async () => {
  try {
    const response = await fetch('/api/tenants')
    const data = await response.json()
    tenants.value = data.items || []
  } catch (error) {
    console.error('加载租户失败:', error)
  }
}

const saveServer = async () => {
  try {
    if (editingServer.value) {
      await fetch(`/api/mcp-servers/${editingServer.value.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form.value)
      })
    } else {
      await fetch('/api/mcp-servers', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form.value)
      })
    }
    showAddModal.value = false
    loadServers()
    resetForm()
  } catch (error) {
    console.error('保存 MCP 服务器失败:', error)
    alert('保存失败')
  }
}

const testServer = async (server: MCPServer) => {
  try {
    const response = await fetch(`/api/mcp-servers/${server.id}/test`)
    const data = await response.json()
    alert(data.success ? '连接成功' : '连接失败: ' + data.message)
    loadServers()
  } catch (error) {
    console.error('测试连接失败:', error)
  }
}

const syncTools = async (server: MCPServer) => {
  try {
    const response = await fetch(`/api/mcp-servers/${server.id}/sync`)
    const data = await response.json()
    alert(`同步完成，新增 ${data.added} 个工具`)
    loadServers()
  } catch (error) {
    console.error('同步工具失败:', error)
  }
}

const deleteServer = async (server: MCPServer) => {
  if (!confirm(`确定删除 MCP 服务器 "${server.name}" 吗？`)) return
  
  try {
    await fetch(`/api/mcp-servers/${server.id}`, { method: 'DELETE' })
    loadServers()
  } catch (error) {
    console.error('删除 MCP 服务器失败:', error)
  }
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const resetForm = () => {
  form.value = {
    name: '',
    url: '',
    api_key: '',
    description: '',
    allowed_tenants: []
  }
  editingServer.value = null
}

onMounted(() => {
  loadServers()
  loadTenants()
})
</script>

<style scoped>
.mcp-servers {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
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

.status-tag.online {
  background: #e8f5e9;
  color: #2e7d32;
}

.status-tag.offline {
  background: #ffebee;
  color: #c62828;
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
  width: 600px;
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
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  margin-bottom: 8px;
  font-size: 14px;
  font-weight: 500;
}

.form-input, .form-select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
}

.form-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-size: 14px;
  min-height: 80px;
  resize: vertical;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.tools-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.tool-card {
  padding: 12px;
  background: #f5f5f5;
  border-radius: 4px;
}

.tool-card h4 {
  margin-bottom: 8px;
  font-size: 14px;
}

.tool-card p {
  margin-bottom: 8px;
  font-size: 13px;
  color: #666;
}

.tool-params {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.param-tag {
  padding: 4px 8px;
  background: white;
  border-radius: 4px;
  font-size: 12px;
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
</style>