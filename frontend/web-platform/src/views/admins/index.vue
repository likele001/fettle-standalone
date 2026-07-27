<template>
  <div class="page-container">
    <div class="page-header">
      <h2>管理员账号</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        添加管理员
      </el-button>
    </div>

    <!-- 管理员列表 -->
    <el-card>
      <el-table :data="admins" v-loading="loading" stripe style="width: 100%">
        <el-table-column prop="username" label="用户名" min-width="140" />
        <el-table-column prop="email" label="邮箱" min-width="180" />
        <el-table-column label="角色" width="120">
          <template #default="{ row }">
            <el-tag :type="roleTagType(row.role)" size="small">
              {{ roleLabel(row.role) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'" size="small">
              {{ row.status === 'active' ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="最后登录" width="170">
          <template #default="{ row }">
            {{ row.last_login_at ? formatTime(row.last_login_at) : '从未登录' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="editAdmin(row)">编辑</el-button>
            <el-button type="warning" link size="small" @click="showResetPassword(row)">重置密码</el-button>
            <el-button
              :type="row.status === 'active' ? 'danger' : 'success'"
              link
              size="small"
              @click="toggleStatus(row)"
            >
              {{ row.status === 'active' ? '禁用' : '启用' }}
            </el-button>
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
          @change="loadAdmins"
        />
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑管理员' : '添加管理员'"
      width="500px"
    >
      <el-form :model="formData" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="formData.username" placeholder="请输入用户名" :disabled="isEdit" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="formData.email" placeholder="请输入邮箱" />
        </el-form-item>
        <el-form-item label="密码" prop="password" v-if="!isEdit">
          <el-input
            v-model="formData.password"
            type="password"
            show-password
            placeholder="请输入密码"
          />
        </el-form-item>
        <el-form-item label="角色" prop="role">
          <el-select v-model="formData.role" placeholder="选择角色">
            <el-option label="超级管理员" value="super_admin" />
            <el-option label="管理员" value="admin" />
            <el-option label="运营人员" value="operator" />
          </el-select>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 重置密码对话框 -->
    <el-dialog
      v-model="resetDialogVisible"
      title="重置密码"
      width="400px"
    >
      <el-form :model="resetForm" :rules="resetRules" ref="resetFormRef" label-width="80px">
        <el-form-item label="用户">
          <span>{{ currentAdmin?.username }}</span>
        </el-form-item>
        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="resetForm.newPassword"
            type="password"
            show-password
            placeholder="请输入新密码"
          />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input
            v-model="resetForm.confirmPassword"
            type="password"
            show-password
            placeholder="请再次输入密码"
          />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="resetDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleResetPassword" :loading="resetting">
          确认重置
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getAdmins, createAdmin, updateAdmin, toggleAdminStatus, resetAdminPassword, type AdminUser } from '@/api/admin'
import type { FormInstance, FormRules } from 'element-plus'

const admins = ref<AdminUser[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const editingId = ref('')

const formData = ref({
  username: '',
  email: '',
  password: '',
  role: 'admin'
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' }
  ],
  role: [{ required: true, message: '请选择角色', trigger: 'change' }]
}

const resetDialogVisible = ref(false)
const resetting = ref(false)
const resetFormRef = ref<FormInstance>()
const currentAdmin = ref<AdminUser | null>(null)

const resetForm = ref({
  newPassword: '',
  confirmPassword: ''
})

const resetRules: FormRules = {
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码至少 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== resetForm.value.newPassword) {
          callback(new Error('两次输入的密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ]
}

function roleTagType(role: string) {
  const map: Record<string, string> = { super_admin: 'danger', admin: 'warning', operator: 'info' }
  return map[role] || 'info'
}

function roleLabel(role: string) {
  const map: Record<string, string> = { super_admin: '超级管理员', admin: '管理员', operator: '运营人员' }
  return map[role] || role
}

function formatTime(t: string) {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

async function loadAdmins() {
  loading.value = true
  try {
    const res = await getAdmins({ page: page.value, page_size: pageSize.value })
    admins.value = res.items
    total.value = res.total
  } catch {
    admins.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  isEdit.value = false
  editingId.value = ''
  formData.value = { username: '', email: '', password: '', role: 'admin' }
  dialogVisible.value = true
}

function editAdmin(admin: AdminUser) {
  isEdit.value = true
  editingId.value = admin.id
  formData.value = { username: admin.username, email: admin.email, password: '', role: admin.role }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updateAdmin(editingId.value, { email: formData.value.email, role: formData.value.role as AdminUser['role'] })
        ElMessage.success('管理员已更新')
      } else {
        await createAdmin(formData.value)
        ElMessage.success('管理员已创建')
      }
      dialogVisible.value = false
      await loadAdmins()
    } catch (e: any) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function toggleStatus(admin: AdminUser) {
  const newStatus = admin.status === 'active' ? 'disabled' : 'active'
  const label = newStatus === 'active' ? '启用' : '禁用'
  try {
    await ElMessageBox.confirm(`确定要${label}管理员「${admin.username}」吗？`, '确认操作', {
      type: 'warning'
    })
  } catch { return }

  try {
    await toggleAdminStatus(admin.id, newStatus)
    ElMessage.success(`管理员已${label}`)
    await loadAdmins()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

function showResetPassword(admin: AdminUser) {
  currentAdmin.value = admin
  resetForm.value = { newPassword: '', confirmPassword: '' }
  resetDialogVisible.value = true
}

async function handleResetPassword() {
  if (!resetFormRef.value) return
  await resetFormRef.value.validate(async (valid) => {
    if (!valid) return
    resetting.value = true
    try {
      await resetAdminPassword(currentAdmin.value!.id, resetForm.value.newPassword)
      ElMessage.success('密码已重置')
      resetDialogVisible.value = false
    } catch (e: any) {
      ElMessage.error(e.message || '重置失败')
    } finally {
      resetting.value = false
    }
  })
}

onMounted(() => loadAdmins())
</script>
