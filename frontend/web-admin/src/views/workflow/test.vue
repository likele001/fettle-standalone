<template>
  <div class="workflow-test" v-if="workflow">
    <div class="test-header">
      <h3>测试工作流: {{ workflow.name }}</h3>
      <button class="btn btn-secondary" @click="$emit('close')">关闭</button>
    </div>

    <div class="test-content">
      <div class="input-section">
        <h4>输入参数</h4>
        <textarea 
          v-model="inputData" 
          placeholder='{"input": "你的问题"}'
          class="json-input"
        ></textarea>
        <button class="btn btn-primary" @click="execute" :disabled="isExecuting">
          {{ isExecuting ? '执行中...' : '执行' }}
        </button>
      </div>

      <div class="output-section">
        <h4>执行结果</h4>
        <div class="result-container">
          <div v-if="result" class="result-content">
            <pre class="json-output">{{ JSON.stringify(result, null, 2) }}</pre>
          </div>
          <div v-else-if="isExecuting" class="loading">
            <span class="spinner"></span>
            <span>执行中...</span>
          </div>
          <div v-else class="empty">
            点击执行按钮开始测试
          </div>
        </div>
      </div>

      <div class="logs-section">
        <h4>执行日志</h4>
        <div class="logs-container">
          <div 
            v-for="(log, index) in logs" 
            :key="index"
            class="log-item"
            :class="log.level"
          >
            <span class="log-time">{{ log.time }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { workflowApi, type Workflow } from '@/api/workflow'

const props = defineProps<{
  workflow: Workflow | null
}>()

defineEmits<{
  close: []
}>()

const inputData = ref('{"input": "你的问题"}')
const result = ref<Record<string, any> | null>(null)
const logs = ref<{ time: string; level: string; message: string }[]>([])
const isExecuting = ref(false)

const execute = async () => {
  isExecuting.value = true
  result.value = null
  logs.value = []

  try {
    const parsedInput = JSON.parse(inputData.value)
    
    if (!props.workflow) return

    logs.value.push({
      time: new Date().toLocaleTimeString(),
      level: 'info',
      message: `开始执行工作流: ${props.workflow.name}`
    })

    const response = await workflowApi.executeWorkflow(props.workflow.id, parsedInput)
    result.value = response.data

    logs.value.push({
      time: new Date().toLocaleTimeString(),
      level: 'success',
      message: '执行完成'
    })
  } catch (error: any) {
    logs.value.push({
      time: new Date().toLocaleTimeString(),
      level: 'error',
      message: error.message || '执行失败'
    })
    console.error('执行工作流失败:', error)
  } finally {
    isExecuting.value = false
  }
}
</script>

<style scoped>
.workflow-test {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f5f5;
}

.test-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: white;
  border-bottom: 1px solid #eee;
}

.test-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 20px;
  gap: 20px;
  overflow: auto;
}

.input-section, .output-section, .logs-section {
  background: white;
  border-radius: 8px;
  padding: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
}

.input-section h4, .output-section h4, .logs-section h4 {
  margin-bottom: 12px;
  font-size: 14px;
  color: #333;
}

.json-input {
  width: 100%;
  min-height: 100px;
  padding: 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  font-family: monospace;
  font-size: 13px;
  margin-bottom: 12px;
  resize: vertical;
}

.result-container {
  min-height: 150px;
  border: 1px solid #ddd;
  border-radius: 4px;
  overflow: auto;
}

.json-output {
  padding: 12px;
  font-family: monospace;
  font-size: 13px;
  color: #333;
  white-space: pre-wrap;
}

.loading {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 10px;
  height: 150px;
  color: #999;
}

.spinner {
  width: 20px;
  height: 20px;
  border: 2px solid #409eff;
  border-top-color: transparent;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.empty {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 150px;
  color: #999;
}

.logs-container {
  max-height: 200px;
  overflow-y: auto;
  border: 1px solid #ddd;
  border-radius: 4px;
}

.log-item {
  display: flex;
  gap: 10px;
  padding: 8px 12px;
  font-size: 12px;
  border-bottom: 1px solid #eee;
}

.log-item:last-child {
  border-bottom: none;
}

.log-item.info {
  background: #f0f9ff;
}

.log-item.success {
  background: #f0fdf4;
}

.log-item.error {
  background: #fef2f2;
}

.log-time {
  color: #999;
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

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>