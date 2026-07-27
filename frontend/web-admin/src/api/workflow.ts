import request from './request'

export interface WorkflowNode {
  id: string
  type: string
  name: string
  config: Record<string, any>
  position: { x: number; y: number }
}

export interface WorkflowEdge {
  id: string
  source: string
  target: string
  sourceHandle?: string
  targetHandle?: string
}

export interface Workflow {
  id: string
  name: string
  description: string
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
  start_node: string
  tenant_id: string
  created_at: string
  updated_at: string
  is_public: boolean
}

export interface WorkflowInstance {
  instance_id: string
  workflow_id: string
  status: string
  input_data: Record<string, any>
  output_data: Record<string, any>
  created_at: string
  updated_at: string
}

export interface CreateWorkflowRequest {
  name: string
  description?: string
  nodes: WorkflowNode[]
  edges: WorkflowEdge[]
  start_node: string
  is_public?: boolean
}

export interface UpdateWorkflowRequest {
  name?: string
  description?: string
  nodes?: WorkflowNode[]
  edges?: WorkflowEdge[]
  start_node?: string
  is_public?: boolean
}

export const workflowApi = {
  listWorkflows(page: number = 1, pageSize: number = 20) {
    return request.get('/workflows', { params: { page, page_size: pageSize } })
  },

  getWorkflow(id: string) {
    return request.get(`/workflows/${id}`)
  },

  createWorkflow(data: CreateWorkflowRequest) {
    return request.post('/workflows', data)
  },

  updateWorkflow(id: string, data: UpdateWorkflowRequest) {
    return request.put(`/workflows/${id}`, data)
  },

  deleteWorkflow(id: string) {
    return request.delete(`/workflows/${id}`)
  },

  executeWorkflow(id: string, inputData: Record<string, any>) {
    return request.post(`/workflows/${id}/execute`, { input_data: inputData })
  },

  streamExecuteWorkflow(id: string, inputData: Record<string, any>) {
    return request.post(`/workflows/${id}/stream`, { input_data: inputData }, { responseType: 'stream' })
  },

  listInstances(workflowId: string, page: number = 1, pageSize: number = 20) {
    return request.get(`/workflows/${workflowId}/instances`, { params: { page, page_size: pageSize } })
  },

  getInstance(instanceId: string) {
    return request.get(`/workflow-instances/${instanceId}`)
  },

  getNodeTypes() {
    return request.get('/workflows/node-types')
  },

  validateWorkflow(data: { nodes: WorkflowNode[]; edges: WorkflowEdge[] }) {
    return request.post('/workflows/validate', data)
  }
}