<template>
  <view class="kb-detail-page">
    <view class="page-header">
      <view class="header-left">
        <text class="back-btn" @click="goBack">{{ '<< ' + '返回' }}</text>
        <text class="header-title">{{ kbName }}</text>
      </view>
      <button class="upload-btn" @click="uploadDocument">{{ '上传文档' }}</button>
    </view>

    <view class="document-list">
      <view v-for="doc in documents" :key="doc.id" class="document-card">
        <view class="doc-icon">
          <text class="doc-icon-text">{{ getFileIcon(doc.type) }}</text>
        </view>
        <view class="doc-info">
          <text class="doc-name">{{ doc.name }}</text>
          <view class="doc-meta">
            <text>{{ getFileTypeLabel(doc.type) }}</text>
            <text class="doc-divider">|</text>
            <text>{{ formatSize(doc.size) }}</text>
            <text class="doc-divider">|</text>
            <text :class="'doc-status ' + doc.status">{{ statusText(doc.status) }}</text>
          </view>
        </view>
        <view class="doc-actions">
          <button class="delete-btn" @click="confirmDelete(doc)">{{ '删除' }}</button>
        </view>
      </view>

      <view v-if="documents.length === 0 && !loading" class="empty-state">
        <text class="empty-text">{{ '暂无文档' }}</text>
        <text class="empty-tip">{{ '点击上方“上传文档”添加知识库内容' }}</text>
      </view>
    </view>

    <view v-if="loading" class="loading-wrapper">
      <text>{{ '加载中...' }}</text>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getDocuments, deleteDocument, uploadDocument, type Document } from '@/api/knowledge'

const kbId = ref('')
const kbName = ref('')
const documents = ref<Document[]>([])
const loading = ref(false)

onMounted(() => {
  const pages = getCurrentPages()
  const cp = pages[pages.length - 1] as any
  const options = cp.$page?.options || cp.options

  kbId.value = options.kb_id || ''
  kbName.value = decodeURIComponent(options.kb_name || '')
  if (!kbId.value) {
    uni.showToast({ title: '参数错误', icon: 'none' })
    setTimeout(() => uni.navigateBack(), 1500)
    return
  }
  loadDocuments()
})

async function loadDocuments() {
  loading.value = true
  try {
    const res = await getDocuments(kbId.value)
    documents.value = res.items || []
  } catch (e) { console.error(e) }
  finally { loading.value = false }
}

async function uploadDocument() {
  uni.chooseImage({
    count: 1,
    success: async (res) => {
      uni.showLoading({ title: '上传中...', mask: true })
      try {
        await uploadDocument(kbId.value, res.tempFilePaths[0])
        uni.hideLoading()
        uni.showToast({ title: '上传成功', icon: 'success' })
        await loadDocuments()
      } catch {
        uni.hideLoading()
        uni.showToast({ title: '上传失败', icon: 'none' })
      }
    }
  })
}

function confirmDelete(doc: Document) {
  uni.showModal({
    title: '提示',
    content: '确定要删除文档“' + doc.name + '”吗？',
    success: async (res) => {
      if (res.confirm) {
        try {
          await deleteDocument(kbId.value, doc.id)
          uni.showToast({ title: '删除成功', icon: 'success' })
          await loadDocuments()
        } catch {
          uni.showToast({ title: '删除失败', icon: 'none' })
        }
      }
    }
  })
}

function goBack() { uni.navigateBack() }
function getFileIcon(t: string): string {
  const m: Record<string,string> = { pdf:'PDF', docx:'DOC', txt:'TXT', md:'MD', image:'IMG' }
  return m[t] || 'FILE'
}
function getFileTypeLabel(t: string): string {
  const m: Record<string,string> = { pdf:'PDF', docx:'Word', txt:'文本', md:'MD', image:'图片' }
  return m[t] || t
}
function statusText(s: string): string {
  const m: Record<string,string> = { pending:'待处理', processing:'处理中', completed:'已完成', error:'失败' }
  return m[s] || s
}
function formatSize(s: number): string {
  if (s < 1024) return s + 'B'
  if (s < 1048576) return (s/1024).toFixed(1) + 'KB'
  return (s/1048576).toFixed(1) + 'MB'
}
</script>

<style scoped>
.kb-detail-page { min-height: 100vh; background: #f5f5f5; padding-bottom: 40rpx; }
.page-header { display: flex; align-items: center; justify-content: space-between; padding: 24rpx 32rpx; background: #fff; border-bottom: 1rpx solid #f0f0f0; }
.header-left { display: flex; align-items: center; flex: 1; min-width: 0; }
.back-btn { font-size: 28rpx; color: #409eff; margin-right: 20rpx; flex-shrink: 0; }
.header-title { font-size: 36rpx; font-weight: 600; color: #333; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.upload-btn { background: #409eff; color: #fff; border: none; border-radius: 8rpx; padding: 12rpx 24rpx; font-size: 26rpx; flex-shrink: 0; }
.document-list { padding: 20rpx; }
.document-card { display: flex; align-items: center; background: #fff; border-radius: 16rpx; padding: 24rpx; margin-bottom: 20rpx; }
.doc-icon { width: 72rpx; height: 72rpx; background: #f0f7ff; border-radius: 12rpx; display: flex; align-items: center; justify-content: center; margin-right: 20rpx; flex-shrink: 0; }
.doc-icon-text { font-size: 24rpx; font-weight: 600; color: #409eff; }
.doc-info { flex: 1; min-width: 0; }
.doc-name { font-size: 30rpx; font-weight: 500; color: #333; margin-bottom: 8rpx; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.doc-meta { font-size: 24rpx; color: #999; display: flex; align-items: center; }
.doc-divider { margin: 0 8rpx; }
.doc-status.completed { color: #67c23a; }
.doc-status.processing { color: #e6a23c; }
.doc-status.error { color: #f56c6c; }
.doc-status.pending { color: #909399; }
.doc-actions { margin-left: 16rpx; flex-shrink: 0; }
.delete-btn { background: #fff; color: #f56c6c; border: 1rpx solid #f56c6c; border-radius: 8rpx; padding: 8rpx 20rpx; font-size: 24rpx; }
.empty-state { text-align: center; padding: 120rpx 0; }
.empty-text { font-size: 32rpx; color: #666; margin-bottom: 12rpx; }
.empty-tip { font-size: 26rpx; color: #999; }
.loading-wrapper { text-align: center; padding: 40rpx 0; color: #999; font-size: 28rpx; }
</style>
