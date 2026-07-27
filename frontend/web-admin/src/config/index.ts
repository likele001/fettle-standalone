export const config = {
  // API基础URL
  apiBaseUrl: import.meta.env.VITE_API_BASE_URL || '/api',

  // 应用名称
  appName: 'AI智能体平台',

  // 版本
  version: '1.0.0',

  // 分页默认配置
  pagination: {
    defaultPageSize: 20,
    pageSizes: [10, 20, 50, 100]
  },

  // 文件上传配置
  upload: {
    maxSize: 10 * 1024 * 1024, // 10MB
    accept: ['.pdf', '.doc', '.docx', '.txt', '.xlsx', '.xls']
  },

  // WebSocket配置
  websocket: {
    url: import.meta.env.VITE_WS_URL || 'ws://localhost:20001/ws',
    reconnectInterval: 3000,
    maxReconnectAttempts: 5
  },

  // 角色定义
  roles: {
    SUPER_ADMIN: 'super_admin',
    ADMIN: 'admin',
    OPERATOR: 'operator',
    MEMBER: 'member',
    VIEWER: 'viewer'
  },

  // 套餐类型
  planTypes: {
    FREE: 'free',
    PRO: 'pro',
    ENTERPRISE: 'enterprise'
  },

  // 环境
  env: import.meta.env.MODE || 'development',
  isDev: import.meta.env.DEV,
  isProd: import.meta.env.PROD
}

export default config
