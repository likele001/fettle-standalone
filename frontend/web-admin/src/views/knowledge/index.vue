<template>
  <div class="knowledge-list">
    <div class="page-header">
      <h2>知识库管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建知识库
      </el-button>
    </div>

    <el-table :data="knowledgeBases" stripe>
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="description" label="描述" show-overflow-tooltip />
      <el-table-column prop="doc_count" label="文档数" width="100" />
      <el-table-column prop="chunk_count" label="分块数" width="100" />
      <el-table-column label="大小" width="120">
        <template #default="{ row }">
          {{ formatSize(row.total_size) }}
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small">
            {{ row.status === 'active' ? '正常' : '处理中' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="viewDocuments(row)">文档</el-button>
          <el-button size="small" @click="editKB(row)">编辑</el-button>
          <el-popconfirm title="确定删除？" @confirm="handleDelete(row.id)">
            <template #reference>
              <el-button size="small" type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <el-empty v-if="knowledgeBases.length === 0" description="暂无知识库" />

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑知识库' : '创建知识库'"
      width="500px"
    >
      <el-form :model="formData" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入知识库名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input
            v-model="formData.description"
            type="textarea"
            :rows="3"
            placeholder="描述知识库内容"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 文档管理对话框 -->
    <el-dialog v-model="docDialogVisible" title="文档管理" width="700px">
      <div class="doc-header">
        <el-upload
          :action="uploadUrl"
          :headers="uploadHeaders"
          :on-success="handleUploadSuccess"
          :before-upload="beforeUpload"
          accept=".pdf,.docx,.txt,.xlsx"
        >
          <el-button type="primary">上传文档</el-button>
          <template #tip>
            <div class="el-upload__tip">支持 PDF、Word、TXT、Excel 格式</div>
          </template>
        </el-upload>
      </div>

      <el-table :data="documents" stripe>
        <el-table-column prop="file_name" label="文件名" />
        <el-table-column prop="file_type" label="类型" width="80" />
        <el-table-column label="大小" width="100">
          <template #default="{ row }">{{ formatSize(row.file_size) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="docStatusType(row.status)" size="small">
              {{ docStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="80">
          <template #default="{ row }">
            <el-popconfirm title="确定删除？" @confirm="handleDeleteDoc(row.id)">
              <template #reference>
                <el-button size="small" type="danger" link>删除</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  getKnowledgeBases,
  createKnowledgeBase,
  updateKnowledgeBase,
  deleteKnowledgeBase,
  getDocuments,
  deleteDocument,
  type KnowledgeBase,
  type KnowledgeDocument
} from '@/api/knowledge'
import { getToken } from '@/utils/token'

const knowledgeBases = ref<KnowledgeBase[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const editingId = ref<string | null>(null)
const submitting = ref(false)
const formRef = ref()

const formData = ref({
  name: '',
  description: ''
})

const rules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
}

// 文档管理
const docDialogVisible = ref(false)
const currentKBId = ref<string | null>(null)
const documents = ref<KnowledgeDocument[]>([])

const uploadUrl = computed(() => {
  if (!currentKBId.value) return ''
  const baseUrl = import.meta.env.VITE_API_BASE_URL || '/api/v1'
  return `${baseUrl}/knowledge/${currentKBId.value}/documents`
})

const uploadHeaders = computed(() => {
  const token = getToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
})

const formatSize = (bytes: number) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

const docStatusType = (status: string) => {
  const map: Record<string, string> = {
    pending: 'info',
    processing: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return map[status] || 'info'
}

const docStatusText = (status: string) => {
  const map: Record<string, string> = {
    pending: '待处理',
    processing: '处理中',
    completed: '已完成',
    failed: '失败'
  }
  return map[status] || status
}

const loadKnowledgeBases = async () => {
  try {
    const res = await getKnowledgeBases()
    knowledgeBases.value = res.items
  } catch (e) {
    ElMessage.error('加载知识库列表失败')
  }
}

const showCreateDialog = () => {
  isEdit.value = false
  editingId.value = null
  formData.value = { name: '', description: '' }
  dialogVisible.value = true
}

const editKB = (kb: KnowledgeBase) => {
  isEdit.value = true
  editingId.value = kb.id
  formData.value = { name: kb.name, description: kb.description }
  dialogVisible.value = true
}

const handleSubmit = async () => {
  if (!formRef.value) return
  await formRef.value.validate()

  submitting.value = true
  try {
    if (isEdit.value && editingId.value) {
      await updateKnowledgeBase(editingId.value, formData.value)
      ElMessage.success('更新成功')
    } else {
      await createKnowledgeBase(formData.value)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadKnowledgeBases()
  } catch (e) {
    ElMessage.error('操作失败')
  } finally {
    submitting.value = false
  }
}

const handleDelete = async (id: string) => {
  try {
    await deleteKnowledgeBase(id)
    ElMessage.success('删除成功')
    loadKnowledgeBases()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

const viewDocuments = async (kb: KnowledgeBase) => {
  currentKBId.value = kb.id
  docDialogVisible.value = true
  await loadDocuments()
}

const loadDocuments = async () => {
  if (!currentKBId.value) return
  try {
    const res = await getDocuments(currentKBId.value)
    documents.value = res.items
  } catch (e) {
    ElMessage.error('加载文档列表失败')
  }
}

const beforeUpload = (file: File) => {
  const allowedTypes = ['application/pdf', 'application/vnd.openxmlformats-officedocument.wordprocessingml.document', 'text/plain', 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet']
  if (!allowedTypes.includes(file.type) && !file.name.match(/\.(pdf|docx|txt|xlsx)$/i)) {
    ElMessage.error('不支持的文件格式')
    return false
  }
  if (file.size > 50 * 1024 * 1024) {
    ElMessage.error('文件大小不能超过 50MB')
    return false
  }
  return true
}

const handleUploadSuccess = () => {
  ElMessage.success('上传成功')
  loadDocuments()
  loadKnowledgeBases()
}

const handleDeleteDoc = async (docId: string) => {
  if (!currentKBId.value) return
  try {
    await deleteDocument(currentKBId.value, docId)
    ElMessage.success('删除成功')
    loadDocuments()
    loadKnowledgeBases()
  } catch (e) {
    ElMessage.error('删除失败')
  }
}

onMounted(() => {
  loadKnowledgeBases()
})
</script>

<style scoped lang="scss">
.knowledge-list {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.doc-header {
  margin-bottom: 16px;
}
</style>
