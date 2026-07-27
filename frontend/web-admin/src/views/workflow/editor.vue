<template>
  <div class="workflow-editor">
    <div class="editor-header">
      <h3>{{ workflow ? '编辑工作流' : '创建工作流' }}</h3>
      <button class="btn btn-secondary" @click="$emit('close')">关闭</button>
    </div>

    <div class="editor-content">
      <div class="left-panel">
        <div class="panel-section">
          <h4>工作流信息</h4>
          <input 
            type="text" 
            v-model="form.name" 
            placeholder="工作流名称"
            class="form-input"
          />
          <textarea 
            v-model="form.description" 
            placeholder="工作流描述"
            class="form-textarea"
          ></textarea>
          <label class="form-checkbox">
            <input type="checkbox" v-model="form.is_public" />
            设为公开
          </label>
        </div>

        <div class="panel-section">
          <h4>节点类型</h4>
          <div class="node-types">
            <div 
              v-for="nodeType in nodeTypes" 
              :key="nodeType.type"
              class="node-type-item"
              @dragstart="onDragStart($event, nodeType)"
              draggable="true"
            >
              <span :class="['node-icon', nodeType.type]">{{ nodeType.icon }}</span>
              <span>{{ nodeType.name }}</span>
            </div>
          </div>
        </div>

        <div class="panel-section">
          <h4>选中节点配置</h4>
          <div v-if="selectedNode" class="node-config">
            <input 
              type="text" 
              v-model="selectedNode.name" 
              placeholder="节点名称"
              class="form-input"
            />
            <div class="config-params">
              <template v-if="selectedNode.type === 'llm'">
                <input 
                  type="text" 
                  v-model="selectedNode.config.model" 
                  placeholder="模型名称"
                  class="form-input"
                />
                <input 
                  type="number" 
                  v-model="selectedNode.config.temperature" 
                  placeholder="温度"
                  class="form-input"
                  min="0" max="1" step="0.1"
                />
                <input 
                  type="number" 
                  v-model="selectedNode.config.max_tokens" 
                  placeholder="最大Token"
                  class="form-input"
                  min="1"
                />
                <textarea 
                  v-model="selectedNode.config.prompt" 
                  placeholder="提示词"
                  class="form-textarea"
                ></textarea>
              </template>
              <template v-else-if="selectedNode.type === 'rag'">
                <input 
                  type="text" 
                  v-model="selectedNode.config.knowledge_base_id" 
                  placeholder="知识库ID"
                  class="form-input"
                />
                <select v-model="selectedNode.config.strategy" class="form-select">
                  <option value="simple">简单检索</option>
                  <option value="multi_query">多查询检索</option>
                  <option value="compression">上下文压缩</option>
                </select>
              </template>
              <template v-else-if="selectedNode.type === 'tool'">
                <select v-model="selectedNode.config.tool_name" class="form-select">
                  <option value="">选择工具</option>
                  <option v-for="tool in availableTools" :key="tool.name" :value="tool.name">
                    {{ tool.name }}
                  </option>
                </select>
                <textarea 
                  v-model="selectedNode.config.parameters" 
                  placeholder="参数JSON"
                  class="form-textarea"
                ></textarea>
              </template>
              <template v-else-if="selectedNode.type === 'condition'">
                <textarea 
                  v-model="selectedNode.config.condition" 
                  placeholder="条件表达式"
                  class="form-textarea"
                ></textarea>
              </template>
              <template v-else-if="selectedNode.type === 'http'">
                <input type="text" v-model="selectedNode.config.url" placeholder="请求URL" class="form-input" />
                <select v-model="selectedNode.config.method" class="form-select">
                  <option value="GET">GET</option>
                  <option value="POST">POST</option>
                  <option value="PUT">PUT</option>
                  <option value="DELETE">DELETE</option>
                </select>
                <textarea v-model="selectedNode.config.body" placeholder='请求体JSON，如 {"key": "{{value}}"}' class="form-textarea"></textarea>
                <input type="text" v-model="selectedNode.config.output_key" placeholder="输出变量名（默认 http_response）" class="form-input" />
              </template>
              <template v-else-if="selectedNode.type === 'cron'">
                <input type="text" v-model="selectedNode.config.cron" placeholder="Cron 表达式，如 0 9 * * *" class="form-input" />
                <input type="text" v-model="selectedNode.config.timezone" placeholder="时区，如 Asia/Shanghai" class="form-input" />
                <p class="config-hint">发布后定时生效，表达式说明：分 时 日 月 周</p>
              </template>
              <template v-else-if="selectedNode.type === 'webhook'">
                <input type="text" v-model="selectedNode.config.path" placeholder="Webhook 路径（可选）" class="form-input" />
                <input type="text" v-model="selectedNode.config.secret" placeholder="签名密钥（可选）" class="form-input" />
              </template>
            </div>
          </div>
          <p v-else class="empty-tip">请选择一个节点进行配置</p>
        </div>
      </div>

      <div class="canvas-container">
        <div 
          class="canvas"
          @dragover.prevent
          @drop="onDrop"
          @click="onCanvasClick"
          @mousemove="onCanvasMouseMove"
          @mouseup="onCanvasMouseUp"
        >
          <svg class="connections">
            <defs>
              <marker id="arrowhead" markerWidth="10" markerHeight="7" refX="9" refY="3.5" orient="auto">
                <polygon points="0 0, 10 3.5, 0 7" fill="#999"/>
              </marker>
            </defs>
            <line 
              v-for="edge in form.edges" 
              :key="edge.id"
              :x1="getNodePosition(edge.source).x + 100"
              :y1="getNodePosition(edge.source).y + 30"
              :x2="getNodePosition(edge.target).x"
              :y2="getNodePosition(edge.target).y + 30"
              stroke="#999"
              stroke-width="2"
              marker-end="url(#arrowhead)"
            />
          </svg>

          <div 
            v-for="node in form.nodes" 
            :key="node.id"
            class="workflow-node"
            :class="[node.type, { selected: selectedNode?.id === node.id }]"
            :style="{ left: node.position.x + 'px', top: node.position.y + 'px' }"
            @mousedown="onNodeMouseDown($event, node)"
            @dblclick="editNode(node)"
          >
            <div class="node-header">
              <span :class="['node-icon', node.type]">{{ getNodeIcon(node.type) }}</span>
              <span class="node-name">{{ node.name }}</span>
            </div>
            <div class="node-ports">
              <div class="port port-input" @click.stop="connectFrom(node)"></div>
              <div class="port port-output" @click.stop="connectTo(node)"></div>
            </div>
          </div>
        </div>

        <div class="canvas-toolbar">
          <button class="btn btn-sm" @click="saveWorkflow">保存</button>
          <button class="btn btn-sm btn-success" @click="validateWorkflow">验证</button>
          <button class="btn btn-sm btn-danger" @click="clearCanvas">清空</button>
        </div>
      </div>
    </div>

    <div v-if="connectionMode" class="connection-mode-hint">
      连接模式: {{ connectionMode === 'from' ? '点击目标节点建立连接' : '点击源节点建立连接' }}
    </div>
  </div>
</template>

<script setup lang="ts">import { ref, reactive, watch } from 'vue';
import { workflowApi, type Workflow, type WorkflowNode, type WorkflowEdge } from '@/api/workflow';
const props = defineProps<{
 workflow: Workflow | null;
}>();
const emit = defineEmits<{
 close: [
 ];
}>();
const form = reactive({
 name: '',
 description: '',
 nodes: [] as WorkflowNode[],
 edges: [] as WorkflowEdge[],
 start_node: '',
 is_public: false
});
const selectedNode = ref<WorkflowNode | null>(null);
const connectionMode = ref<'from' | 'to' | null>(null);
const connectingNode = ref<WorkflowNode | null>(null);
const draggedNodeType = ref<{
    type: string;
    name: string;
    icon: string;
} | null>(null);
const dragState = ref<{
    nodeId: string;
    startX: number;
    startY: number;
    nodeStartX: number;
    nodeStartY: number;
} | null>(null);
const nodeTypes = [
    { type: 'start', name: '开始节点', icon: '▶' },
    { type: 'llm', name: 'LLM调用', icon: '🤖' },
    { type: 'rag', name: '知识库检索', icon: '📚' },
    { type: 'tool', name: '工具调用', icon: '🔧' },
    { type: 'condition', name: '条件判断', icon: '❓' },
    { type: 'http', name: 'HTTP请求', icon: '🌐' },
    { type: 'cron', name: '定时触发', icon: '⏰' },
    { type: 'webhook', name: 'Webhook触发', icon: '🔗' },
    { type: 'end', name: '结束节点', icon: '⏹' }
];
const availableTools = [
 { name: 'web_search', description: '网页搜索' },
 { name: 'http_post', description: 'HTTP POST' },
 { name: 'file_read', description: '读取文件' },
 { name: 'email_send', description: '发送邮件' },
 { name: 'schedule', description: '定时任务' }
];
watch(() => props.workflow, (newWorkflow) => {
 if (newWorkflow) {
 form.name = newWorkflow.name;
 form.description = newWorkflow.description || '';
 form.nodes = JSON.parse(JSON.stringify(newWorkflow.nodes || []));
 form.edges = JSON.parse(JSON.stringify(newWorkflow.edges || []));
 form.start_node = newWorkflow.start_node || '';
 form.is_public = newWorkflow.is_public || false;
 }
}, { immediate: true });
const onDragStart = (event: DragEvent, nodeType: {
 type: string;
 name: string;
 icon: string;
}) => {
 draggedNodeType.value = nodeType;
};
const onDrop = (event: DragEvent) => {
 if (!draggedNodeType.value)
 return;
 const canvas = event.currentTarget as HTMLElement;
 const rect = canvas.getBoundingClientRect();
 const x = event.clientX - rect.left - 100;
 const y = event.clientY - rect.top - 30;
 const newNode: WorkflowNode = {
 id: `node_${Date.now()}`,
 type: draggedNodeType.value.type,
 name: draggedNodeType.value.name,
 config: {},
 position: { x: Math.max(0, x), y: Math.max(0, y) }
 };
 form.nodes.push(newNode);
 draggedNodeType.value = null;
};
const onCanvasClick = () => {
 selectedNode.value = null;
 connectionMode.value = null;
};
const onNodeMouseDown = (event: MouseEvent, node: WorkflowNode) => {
  event.stopPropagation();
  selectedNode.value = node;
  dragState.value = {
    nodeId: node.id,
    startX: event.clientX,
    startY: event.clientY,
    nodeStartX: node.position.x,
    nodeStartY: node.position.y,
  };
};
const onCanvasMouseMove = (event: MouseEvent) => {
  if (!dragState.value) return;
  const ds = dragState.value;
  const dx = event.clientX - ds.startX;
  const dy = event.clientY - ds.startY;
  const node = form.nodes.find(n => n.id === ds.nodeId);
  if (node) {
    node.position.x = Math.max(0, ds.nodeStartX + dx);
    node.position.y = Math.max(0, ds.nodeStartY + dy);
  }
};
const onCanvasMouseUp = () => {
  dragState.value = null;
};
const editNode = (node: WorkflowNode) => {
 selectedNode.value = node;
};
const getNodeIcon = (type: string) => {
 const found = nodeTypes.find(t => t.type === type);
 return found ? found.icon : '📦';
};
const getNodePosition = (nodeId: string) => {
 const node = form.nodes.find(n => n.id === nodeId);
 return node ? node.position : { x: 0, y: 0 };
};
const connectFrom = (node: WorkflowNode) => {
  if (connectionMode.value === 'from' && connectingNode.value) {
    createEdge(connectingNode.value, node);
  }
  connectionMode.value = null;
  connectingNode.value = null;
};
const connectTo = (node: WorkflowNode) => {
  connectingNode.value = node;
  connectionMode.value = 'from';
};
const createEdge = (source: WorkflowNode, target: WorkflowNode) => {
 const existingEdge = form.edges.find(e => e.source === source.id && e.target === target.id);
 if (!existingEdge) {
 const newEdge: WorkflowEdge = {
 id: `edge_${Date.now()}`,
 source: source.id,
 target: target.id
 };
 form.edges.push(newEdge);
 }
};
const saveWorkflow = async () => {
 if (!form.name) {
 alert('请输入工作流名称');
 return;
 }
 try {
 if (props.workflow) {
 await workflowApi.updateWorkflow(props.workflow.id, form);
 }
 else {
 await workflowApi.createWorkflow(form);
 }
 alert('保存成功');
 emit('close');
 }
 catch (error) {
 console.error('保存工作流失败:', error);
 alert('保存失败');
 }
};
const validateWorkflow = async () => {
 try {
 const result = await workflowApi.validateWorkflow({
 nodes: form.nodes,
 edges: form.edges
 });
 alert(`验证结果: ${result.data.valid ? '通过' : '失败'}\n${result.data.message || ''}`);
 }
 catch (error) {
 console.error('验证工作流失败:', error);
 }
};
const clearCanvas = () => {
 if (!confirm('确定清空画布吗？'))
 return;
 form.nodes = [];
 form.edges = [];
 form.start_node = '';
};
</script>

<style scoped>
.workflow-editor {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: #f5f5f5;
}

.editor-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: white;
  border-bottom: 1px solid #eee;
}

.editor-content {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.left-panel {
  width: 320px;
  background: white;
  border-right: 1px solid #eee;
  padding: 16px;
  overflow-y: auto;
}

@media (max-width: 1023px) {
  .left-panel {
    width: 240px;
    padding: 10px;
  }
}

@media (max-width: 767px) {
  .editor-content {
    flex-direction: column;
  }
  .left-panel {
    width: 100%;
    max-height: 40vh;
    border-right: none;
    border-bottom: 1px solid #eee;
  }
  .canvas-container {
    min-height: 50vh;
  }
}

.panel-section {
  margin-bottom: 24px;
}

.panel-section h4 {
  margin-bottom: 12px;
  font-size: 14px;
  color: #333;
}

.form-input, .form-select {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 8px;
  font-size: 13px;
}

.form-textarea {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-bottom: 8px;
  font-size: 13px;
  min-height: 80px;
  resize: vertical;
}

.form-checkbox {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
}

.node-types {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.node-type-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f5f5f5;
  border-radius: 4px;
  cursor: grab;
}

.node-type-item:active {
  cursor: grabbing;
}

.node-icon {
  font-size: 18px;
}

.empty-tip {
  color: #999;
  font-size: 13px;
}

.canvas-container {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.canvas {
  flex: 1;
  position: relative;
  background: #fafafa;
  background-image: 
    linear-gradient(#e8e8e8 1px, transparent 1px),
    linear-gradient(90deg, #e8e8e8 1px, transparent 1px);
  background-size: 20px 20px;
  overflow: auto;
}

.canvas svg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 1;
}

.workflow-node {
  position: absolute;
  width: 200px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.1);
  cursor: move;
  z-index: 2;
}

.workflow-node.selected {
  border: 2px solid #409eff;
}

.workflow-node.start {
  border-left: 4px solid #67c23a;
}

.workflow-node.llm {
  border-left: 4px solid #409eff;
}

.workflow-node.rag {
  border-left: 4px solid #e6a23c;
}

.workflow-node.tool {
  border-left: 4px solid #f56c6c;
}

.workflow-node.condition {
  border-left: 4px solid #909399;
}

.workflow-node.http {
  border-left: 4px solid #22c55e;
}

.workflow-node.cron {
  border-left: 4px solid #a855f7;
}

.workflow-node.webhook {
  border-left: 4px solid #f97316;
}

.config-hint {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.workflow-node.end {
  border-left: 4px solid #667eea;
}

.node-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: #f8f9fa;
  border-radius: 8px 8px 0 0;
}

.node-name {
  font-size: 13px;
  font-weight: 500;
}

.node-ports {
  display: flex;
  justify-content: space-between;
  padding: 8px;
}

.port {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #999;
  cursor: pointer;
}

.port:hover {
  background: #409eff;
}

.canvas-toolbar {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  padding: 12px 20px;
  background: white;
  border-top: 1px solid #eee;
}

@media (max-width: 767px) {
  .editor-header {
    padding: 10px 12px;
    flex-wrap: wrap;
    gap: 8px;
  }
  .editor-header h3 {
    font-size: 15px;
  }
  .canvas-toolbar {
    padding: 8px 12px;
    flex-wrap: wrap;
  }
  .node-types {
    flex-direction: row;
    flex-wrap: wrap;
    gap: 6px;
  }
  .node-type-item {
    flex: 0 0 calc(33% - 4px);
    font-size: 12px;
    padding: 6px 8px;
  }
}

.connection-mode-hint {
  position: fixed;
  bottom: 20px;
  left: 50%;
  transform: translateX(-50%);
  padding: 12px 24px;
  background: #409eff;
  color: white;
  border-radius: 4px;
  z-index: 100;
}

.btn {
  padding: 8px 16px;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 14px;
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
</style>