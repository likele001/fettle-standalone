<template>
  <div class="team-page">
    <div class="page-header">
      <h2>团队管理</h2>
      <el-button type="primary" @click="showInviteDialog">
        <el-icon><Plus /></el-icon> 邀请成员
      </el-button>
    </div>

    <!-- 成员列表 -->
    <el-card>
      <template #header>
        <div class="card-header">
          <span>成员列表</span>
          <el-radio-group v-model="memberFilter" size="small" @change="loadMembers">
            <el-radio-button label="">全部</el-radio-button>
            <el-radio-button label="active">在职</el-radio-button>
            <el-radio-button label="resigned">离职</el-radio-button>
          </el-radio-group>
        </div>
      </template>

      <el-table :data="members" stripe v-loading="loading">
        <el-table-column prop="user_name" label="姓名" width="150" />
        <el-table-column prop="user_phone" label="手机" width="130" />
        <el-table-column prop="user_email" label="邮箱" width="180" show-overflow-tooltip />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag :type="roleTag(row.role)" size="small">{{ roleLabel(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
              {{ row.status === 'active' ? '在职' : '离职' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="joined_at" label="加入时间" width="180">
          <template #default="{ row }">{{ formatDate(row.joined_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <template v-if="row.status === 'active' && row.role !== 'owner'">
              <el-dropdown trigger="click" @command="(cmd: string) => handleRoleChange(row, cmd)">
                <el-button size="small" text>改角色</el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="admin" :disabled="row.role === 'admin'">管理员</el-dropdown-item>
                    <el-dropdown-item command="member" :disabled="row.role === 'member'">成员</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
              <el-popconfirm title="确定将该成员设为离职？" @confirm="handleResign(row)">
                <template #reference>
                  <el-button size="small" text type="warning">离职</el-button>
                </template>
              </el-popconfirm>
            </template>
            <span v-else style="color: #999; font-size: 12px">-</span>
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; display: flex; justify-content: flex-end">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadMembers"
          @current-change="loadMembers"
        />
      </div>
    </el-card>

    <!-- 邀请记录 -->
    <el-card style="margin-top: 20px">
      <template #header>
        <span>邀请记录</span>
      </template>
      <el-table :data="invitations" stripe>
        <el-table-column prop="code" label="邀请码" width="280">
          <template #default="{ row }">
            <code>{{ row.code }}</code>
            <el-button size="small" text @click="copyCode(row.code)">复制</el-button>
          </template>
        </el-table-column>
        <el-table-column prop="invitee_phone" label="受邀手机" width="130" />
        <el-table-column prop="invitee_email" label="受邀邮箱" width="180" />
        <el-table-column prop="role" label="角色" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ roleLabel(row.role) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="used" label="状态" width="80">
          <template #default="{ row }">
            <el-tag :type="row.used ? 'info' : 'success'" size="small">{{ row.used ? '已使用' : '待使用' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">{{ formatDate(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-popconfirm v-if="!row.used" title="确定取消此邀请？" @confirm="handleCancelInvite(row.id)">
              <template #reference>
                <el-button size="small" text type="danger">取消</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 邀请对话框 -->
    <el-dialog v-model="inviteDialogVisible" title="邀请成员" width="480px">
      <el-form :model="inviteForm" label-width="80px">
        <el-form-item label="手机号">
          <el-input v-model="inviteForm.invitee_phone" placeholder="选填" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="inviteForm.invitee_email" placeholder="选填" />
        </el-form-item>
        <el-form-item label="角色">
          <el-radio-group v-model="inviteForm.role">
            <el-radio value="member">成员</el-radio>
            <el-radio value="admin">管理员</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="inviteDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleInvite" :loading="inviting">生成邀请链接</el-button>
      </template>
    </el-dialog>

    <!-- 邀请结果 -->
    <el-dialog v-model="inviteResultVisible" title="邀请链接已生成" width="480px">
      <p style="color: #666; margin-bottom: 12px">将以下链接发送给受邀人即可：</p>
      <el-input :model-value="inviteResultUrl" readonly>
        <template #append>
          <el-button @click="copyCode(inviteResultUrl)">复制</el-button>
        </template>
      </el-input>
      <template #footer>
        <el-button type="primary" @click="inviteResultVisible = false">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getTeamMembers, createInvitation, getInvitations,
  cancelInvitation, updateMemberRole, resignMember,
  type TeamMember, type Invitation
} from '@/api/team'

const members = ref<TeamMember[]>([])
const invitations = ref<Invitation[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const memberFilter = ref('')

const inviteDialogVisible = ref(false)
const inviteResultVisible = ref(false)
const inviteResultUrl = ref('')
const inviting = ref(false)
const inviteForm = ref({
  invitee_phone: '',
  invitee_email: '',
  role: 'member'
})

const roleLabel = (role: string) => {
  const map: Record<string, string> = { owner: '所有者', admin: '管理员', member: '成员' }
  return map[role] || role
}

const roleTag = (role: string) => {
  const map: Record<string, string> = { owner: 'danger', admin: 'warning', member: 'primary' }
  return map[role] || 'info'
}

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const copyCode = (text: string) => {
  navigator.clipboard.writeText(text).then(() => {
    ElMessage.success('已复制到剪贴板')
  })
}

async function loadMembers() {
  loading.value = true
  try {
    const res = await getTeamMembers({ status: memberFilter.value, page: currentPage.value, page_size: pageSize.value }) as any
    members.value = res?.items || []
    total.value = res?.total || 0
  } catch { members.value = [] }
  finally { loading.value = false }
}

async function loadInvitations() {
  try {
    const res = await getInvitations() as any
    invitations.value = res || []
  } catch { invitations.value = [] }
}

function showInviteDialog() {
  inviteForm.value = { invitee_phone: '', invitee_email: '', role: 'member' }
  inviteDialogVisible.value = true
}

async function handleInvite() {
  inviting.value = true
  try {
    const res = await createInvitation(inviteForm.value) as any
    const origin = window.location.origin
    inviteResultUrl.value = `${origin}/invite?code=${res.code}`
    inviteDialogVisible.value = false
    inviteResultVisible.value = true
    loadInvitations()
  } catch (e: any) {
    ElMessage.error(e.message || '邀请失败')
  } finally { inviting.value = false }
}

async function handleRoleChange(member: TeamMember, role: string) {
  try {
    await updateMemberRole(member.id, role)
    ElMessage.success('角色已更新')
    loadMembers()
  } catch (e: any) { ElMessage.error(e.message || '操作失败') }
}

async function handleResign(member: TeamMember) {
  try {
    await resignMember(member.id)
    ElMessage.success('已设为离职')
    loadMembers()
  } catch (e: any) { ElMessage.error(e.message || '操作失败') }
}

async function handleCancelInvite(id: string) {
  try {
    await cancelInvitation(id)
    ElMessage.success('邀请已取消')
    loadInvitations()
  } catch (e: any) { ElMessage.error(e.message || '操作失败') }
}

onMounted(() => { loadMembers(); loadInvitations() })
</script>

<style scoped lang="scss">
.team-page { padding: 0; }
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;
  h2 { margin: 0; font-size: 20px; }
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
