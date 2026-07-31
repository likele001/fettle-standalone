import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/components/Layout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/login', name: 'Login', component: () => import('@/views/login/index.vue'), meta: { public: true, title: '登录' } },
    { path: '/register', name: 'Register', component: () => import('@/views/register/index.vue'), meta: { public: true, title: '注册' } },
    {
      path: '/', component: Layout, meta: { requiresAuth: true }, children: [
        { path: '', name: 'Dashboard', component: () => import('@/views/dashboard/index.vue'), meta: { title: '仪表盘', permission: 'analytics.read' } },
        { path: 'agents', name: 'AgentList', component: () => import('@/views/agent/index.vue'), meta: { title: '智能体管理', permission: 'agents.read' } },
        { path: 'agents/create', name: 'AgentCreate', component: () => import('@/views/agent/create.vue'), meta: { title: '创建智能体', permission: 'agents.create' } },
        { path: 'knowledge', name: 'KnowledgeList', component: () => import('@/views/knowledge/index.vue'), meta: { title: '知识库管理', permission: 'knowledge.read' } },
        { path: 'knowledge/upload', name: 'KnowledgeUpload', component: () => import('@/views/knowledge/index.vue'), meta: { title: '上传文档', permission: 'knowledge.create' } },
        { path: 'conversations', name: 'ConversationList', component: () => import('@/views/conversation/index.vue'), meta: { title: '会话管理', permission: 'conversation.read' } },
        { path: 'conversations/monitor', name: 'ConversationMonitor', component: () => import('@/views/conversation/monitor.vue'), meta: { title: '实时监控', permission: 'conversation.manage' } },
        { path: 'skills', name: 'SkillList', component: () => import('@/views/skill/index.vue'), meta: { title: '技能市场', permission: 'skills.read' } },
        { path: 'channels', name: 'ChannelList', component: () => import('@/views/channel/index.vue'), meta: { title: '渠道管理', permission: 'team.manage' } },
        { path: 'analytics', name: 'Analytics', component: () => import('@/views/analytics/index.vue'), meta: { title: '数据概览', permission: 'analytics.read' } },
        { path: 'analytics/conversations', name: 'AnalyticsConversations', component: () => import('@/views/analytics/conversations.vue'), meta: { title: '会话分析', permission: 'analytics.read' } },
        { path: 'ai-config', name: 'AIConfig', component: () => import('@/views/ai-config/index.vue'), meta: { title: 'AI 模型配置', permission: 'settings.manage' } },
        { path: 'ai-billing', name: 'AIBilling', component: () => import('@/views/ai-billing/index.vue'), meta: { title: 'AI 计费', permission: 'billing.read' } },
        { path: 'workflows', name: 'WorkflowList', component: () => import('@/views/workflow/index.vue'), meta: { title: '工作流列表', permission: 'workflow.read' } },
        { path: 'workflows/create', name: 'WorkflowCreate', component: () => import('@/views/workflow/editor.vue'), meta: { title: '创建工作流', permission: 'workflow.create' } },
        { path: 'workflows/:id', name: 'WorkflowEdit', component: () => import('@/views/workflow/editor.vue'), meta: { title: '编辑工作流', permission: 'workflow.manage' } },
        { path: 'billing', name: 'Billing', component: () => import('@/views/billing/index.vue'), meta: { title: '套餐与计费', permission: 'billing.read' } },
        { path: 'settings/team', name: 'SettingsTeam', component: () => import('@/views/settings/index.vue'), meta: { title: '团队管理', permission: 'team.manage' } },
        { path: 'settings/roles', name: 'SettingsRoles', component: () => import('@/views/settings/roles.vue'), meta: { title: '角色管理', permission: 'role.manage' } },
        { path: 'settings/billing', name: 'SettingsBilling', component: () => import('@/views/settings/index.vue'), meta: { title: '套餐与账单', permission: 'billing.read' } },
        { path: 'settings/profile', name: 'SettingsProfile', component: () => import('@/views/settings/index.vue'), meta: { title: '个人资料', permission: 'team.read' } },
        { path: 'settings/payment', name: 'SettingsPayment', component: () => import('@/views/settings/payment.vue'), meta: { title: '支付配置', permission: 'billing.manage' } },
      ]
    },
    { path: '/:pathMatch(.*)*', redirect: '/' }
  ]
})
export default router
