<template>
  <div class="roles-page">
    <el-card>
      <template #header>
        <div style="display:flex;justify-content:space-between;align-items:center">
          <span>角色管理</span>
          <el-button type="primary" size="small" @click="showCreate">新建角色</el-button>
        </div>
      </template>

      <el-table :data="roles" border stripe v-loading="loading">
        <el-table-column prop="name" label="角色名称" width="140" />
        <el-table-column prop="code" label="角色编码" width="140" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="操作" width="180">
          <template #default="{ row }">
            <el-button size="small" @click="managePerms(row)">权限</el-button>
            <el-button size="small" type="danger" @click="handleDelete(row)" :disabled="row.code === 'super_admin'">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新建角色弹窗 -->
    <el-dialog v-model="createVisible" title="新建角色" width="400px">
      <el-form :model="createForm" label-width="80px">
        <el-form-item label="角色名称"><el-input v-model="createForm.name" /></el-form-item>
        <el-form-item label="角色编码"><el-input v-model="createForm.code" placeholder="英文，如 editor" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="createForm.description" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate" :loading="saving">创建</el-button>
      </template>
    </el-dialog>

    <!-- 权限分配弹窗 -->
    <el-dialog v-model="permVisible" :title="'权限分配 - ' + editingRole?.name" width="640px">
      <el-checkbox-group v-model="permSelected">
        <div v-for="mod in permModules" :key="mod.module" style="margin-bottom:12px">
          <el-divider content-position="left">{{ mod.label }}</el-divider>
          <el-checkbox v-for="p in mod.items" :key="p.code" :value="p.id" :label="p.name" />
        </div>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="permVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSavePerms" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import request from '@/api/request'

interface Role { id: string; name: string; code: string; description: string }
interface Perm { id: string; code: string; name: string; module: string }

const roles = ref<Role[]>([])
const loading = ref(false)
const saving = ref(false)
const createVisible = ref(false)
const createForm = ref({ name: '', code: '', description: '' })

const permVisible = ref(false)
const editingRole = ref<Role | null>(null)
const allPerms = ref<Perm[]>([])
const permSelected = ref<string[]>([])

const permModules = ref<Array<{ module: string; label: string; items: any[] }>>([])

async function loadRoles() {
  loading.value = true
  try {
    const res = await request.get('/roles')
    roles.value = res.data || res || []
  } finally { loading.value = false }
}

async function loadAllPerms() {
  try {
    const res = await request.get('/permissions')
    allPerms.value = res.data || res || []
  } catch { allPerms.value = [] }
}

function groupPerms() {
  const map: Record<string, any[]> = {}
  const labels: Record<string, string> = {
    agents: '智能体', knowledge: '知识库', conversation: '会话',
    analytics: '分析', billing: '计费', team: '团队',
    settings: '系统', workflow: '工作流', skills: '技能', users: '用户'
  }
  allPerms.value.forEach(p => {
    if (!map[p.module]) map[p.module] = []
    map[p.module].push(p)
  })
  permModules.value = Object.keys(map).map(m => ({
    module: m, label: labels[m] || m, items: map[m]
  }))
}

async function managePerms(role: Role) {
  editingRole.value = role
  await loadAllPerms()
  groupPerms()
  try {
    const res = await request.get(`/roles/${role.id}/permissions`)
    const ids = (res.data || res || []).map((p: any) => p.id || p.permission_id)
    permSelected.value = ids
  } catch { permSelected.value = [] }
  permVisible.value = true
}

async function showCreate() {
  createForm.value = { name: '', code: '', description: '' }
  createVisible.value = true
}

async function handleCreate() {
  saving.value = true
  try {
    await request.post('/roles', createForm.value)
    ElMessage.success('创建成功')
    createVisible.value = false
    loadRoles()
  } catch(e: any) { ElMessage.error(e?.message || '创建失败') }
  finally { saving.value = false }
}

async function handleSavePerms() {
  if (!editingRole.value) return
  saving.value = true
  try {
    await request.post(`/roles/${editingRole.value.id}/permissions`, { permission_ids: permSelected.value })
    ElMessage.success('权限已更新')
    permVisible.value = false
  } catch(e: any) { ElMessage.error(e?.message || '保存失败') }
  finally { saving.value = false }
}

async function handleDelete(role: Role) {
  try {
    await request.delete(`/roles/${role.id}`)
    ElMessage.success('已删除')
    loadRoles()
  } catch(e: any) { ElMessage.error(e?.message || '删除失败') }
}

onMounted(() => { loadRoles() })
</script>
