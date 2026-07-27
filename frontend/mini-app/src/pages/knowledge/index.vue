<template>
  <view class="knowledge-page">
    <view class="page-header">
      <text class="page-title">知识库</text>
      <button class="add-btn" @click="showCreateDialog">
        <text>+ 新建</text>
      </button>
    </view>

    <view class="knowledge-list">
      <view
        v-for="kb in knowledgeBases"
        :key="kb.id"
        class="knowledge-card"
        @click="viewDetail(kb)"
      >
        <view class="kb-header">
          <text class="kb-name">{{ kb.name }}</text>
          <view class="kb-status" :class="kb.status">
            <text>{{ statusText(kb.status) }}</text>
          </view>
        </view>
        <text class="kb-desc">{{ kb.description || '暂无描述' }}</text>
        <view class="kb-footer">
          <text class="kb-count">{{ kb.document_count }} 个文档</text>
          <text class="kb-time">{{ formatTime(kb.created_at) }}</text>
        </view>
      </view>

      <view v-if="knowledgeBases.length === 0 && !loading" class="empty-state">
        <image class="empty-icon" src="/static/empty.png" mode="aspectFit" />
        <text class="empty-text">暂无知识库</text>
        <text class="empty-tip">点击右上角"新建"创建您的第一个知识库</text>
      </view>
    </view>

    <view v-if="loading" class="loading-wrapper">
      <text class="loading-text">加载中...</text>
    </view>

    <!-- 创建对话框 -->
    <view v-if="dialogVisible" class="dialog-mask" @click="dialogVisible = false">
      <view class="dialog-content" @click.stop>
        <text class="dialog-title">新建知识库</text>
        <view class="form-item">
          <text class="form-label">名称</text>
          <input
            v-model="formData.name"
            type="text"
            placeholder="请输入知识库名称"
            class="form-input"
          />
        </view>
        <view class="form-item">
          <text class="form-label">描述</text>
          <textarea
            v-model="formData.description"
            placeholder="请输入知识库描述"
            class="form-textarea"
            maxlength="200"
          />
        </view>
        <view class="dialog-footer">
          <button class="dialog-btn cancel" @click="dialogVisible = false">取消</button>
          <button class="dialog-btn confirm" @click="handleCreate">确定</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getKnowledgeBases, createKnowledgeBase, type KnowledgeBase } from '@/api/knowledge'

const knowledgeBases = ref<KnowledgeBase[]>([])
const loading = ref(false)
const dialogVisible = ref(false)
const formData = ref({
  name: '',
  description: ''
})

onMounted(() => {
  loadKnowledgeBases()
})

async function loadKnowledgeBases() {
  loading.value = true
  try {
    const res = await getKnowledgeBases({ page: 1, page_size: 50 })
    knowledgeBases.value = res.items
  } catch (error) {
    uni.showToast({
      title: '加载失败',
      icon: 'none'
    })
  } finally {
    loading.value = false
  }
}

function showCreateDialog() {
  formData.value = { name: '', description: '' }
  dialogVisible.value = true
}

async function handleCreate() {
  if (!formData.value.name.trim()) {
    uni.showToast({
      title: '请输入名称',
      icon: 'none'
    })
    return
  }

  try {
    await createKnowledgeBase(formData.value)
    uni.showToast({
      title: '创建成功',
      icon: 'success'
    })
    dialogVisible.value = false
    await loadKnowledgeBases()
  } catch (error) {
    uni.showToast({
      title: '创建失败',
      icon: 'none'
    })
  }
}

function viewDetail(kb: KnowledgeBase) {
    uni.navigateTo({
      url: "/pages/knowledge/detail?kb_id=" + kb.id + "&kb_name=" + encodeURIComponent(kb.name)
    })
}

function statusText(status: string) {
  const map: Record<string, string> = {
    active: '可用',
    processing: '处理中',
    error: '异常'
  }
  return map[status] || status
}

function formatTime(time: string) {
  const date = new Date(time)
  return `${date.getFullYear()}-${(date.getMonth() + 1).toString().padStart(2, '0')}-${date.getDate().toString().padStart(2, '0')}`
}
</script>

<style lang="scss" scoped>
.knowledge-page {
  min-height: 100vh;
  background: #f5f5f5;
  padding-bottom: 120rpx;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24rpx 32rpx;
  background: #fff;

  .page-title {
    font-size: 36rpx;
    font-weight: 600;
    color: #333;
  }

  .add-btn {
    background: #409eff;
    color: #fff;
    border: none;
    border-radius: 8rpx;
    padding: 12rpx 24rpx;
    font-size: 26rpx;
  }
}

.knowledge-list {
  padding: 20rpx;
}

.knowledge-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);

  &:active {
    background: #f9f9f9;
  }
}

.kb-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12rpx;

  .kb-name {
    font-size: 32rpx;
    font-weight: 600;
    color: #333;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .kb-status {
    padding: 6rpx 16rpx;
    border-radius: 8rpx;
    font-size: 22rpx;
    margin-left: 16rpx;

    &.active {
      background: #e8f5e9;
      color: #67c23a;
    }

    &.processing {
      background: #fff3e0;
      color: #e6a23c;
    }

    &.error {
      background: #ffebee;
      color: #f56c6c;
    }
  }
}

.kb-desc {
  font-size: 26rpx;
  color: #999;
  margin-bottom: 16rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.kb-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .kb-count {
    font-size: 24rpx;
    color: #409eff;
  }

  .kb-time {
    font-size: 24rpx;
    color: #999;
  }
}

.empty-state {
  text-align: center;
  padding: 120rpx 0;

  .empty-icon {
    width: 200rpx;
    height: 200rpx;
    margin-bottom: 24rpx;
  }

  .empty-text {
    display: block;
    font-size: 32rpx;
    color: #666;
    margin-bottom: 12rpx;
  }

  .empty-tip {
    display: block;
    font-size: 26rpx;
    color: #999;
  }
}

.loading-wrapper {
  text-align: center;
  padding: 40rpx 0;

  .loading-text {
    font-size: 28rpx;
    color: #999;
  }
}

.dialog-mask {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.dialog-content {
  width: 80%;
  background: #fff;
  border-radius: 16rpx;
  padding: 40rpx;

  .dialog-title {
    display: block;
    font-size: 36rpx;
    font-weight: 600;
    color: #333;
    margin-bottom: 32rpx;
    text-align: center;
  }
}

.form-item {
  margin-bottom: 24rpx;

  .form-label {
    display: block;
    font-size: 28rpx;
    color: #666;
    margin-bottom: 12rpx;
  }

  .form-input {
    width: 100%;
    height: 80rpx;
    background: #f5f7fa;
    border: 2rpx solid #e4e7ed;
    border-radius: 12rpx;
    padding: 0 24rpx;
    font-size: 28rpx;
    box-sizing: border-box;
  }

  .form-textarea {
    width: 100%;
    height: 160rpx;
    background: #f5f7fa;
    border: 2rpx solid #e4e7ed;
    border-radius: 12rpx;
    padding: 20rpx 24rpx;
    font-size: 28rpx;
    box-sizing: border-box;
  }
}

.dialog-footer {
  display: flex;
  gap: 20rpx;
  margin-top: 32rpx;

  .dialog-btn {
    flex: 1;
    height: 80rpx;
    border-radius: 12rpx;
    font-size: 30rpx;
    font-weight: 500;
    border: none;

    &.cancel {
      background: #f5f7fa;
      color: #666;
    }

    &.confirm {
      background: #409eff;
      color: #fff;
    }
  }
}
</style>
