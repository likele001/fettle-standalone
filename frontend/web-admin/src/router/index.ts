import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/components/Layout.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: { public: true, title: '登录' }
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('@/views/register/index.vue'),
      meta: { public: true, title: '注册' }
    },
    {
      path: '/',
      component: Layout,
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/index.vue'),
          meta: { title: '仪表盘' }
        },
        {
          path: 'agents',
          name: 'AgentList',
          component: () => import('@/views/agent/index.vue'),
          meta: { title: '智能体管理' }
        },
        {
          path: 'agents/create',
          name: 'AgentCreate',
          component: () => import('@/views/agent/create.vue'),
          meta: { title: '创建智能体' }
        },
        {
          path: 'knowledge',
          name: 'KnowledgeList',
          component: () => import('@/views/knowledge/index.vue'),
          meta: { title: '知识库管理' }
        },
        {
          path: 'knowledge/upload',
          name: 'KnowledgeUpload',
          component: () => import('@/views/knowledge/index.vue'),
          meta: { title: '上传文档' }
        },
        {
          path: 'conversations',
          name: 'ConversationList',
          component: () => import('@/views/conversation/index.vue'),
          meta: { title: '对话管理' }
        },
        {
          path: 'conversations/monitor',
          name: 'ConversationMonitor',
          component: () => import('@/views/conversation/monitor.vue'),
          meta: { title: '对话监控' }
        },
        {
          path: 'skills',
          name: 'SkillList',
          component: () => import('@/views/skill/index.vue'),
          meta: { title: '技能市场' }
        },
        {
          path: 'skills/installed',
          name: 'SkillInstalled',
          component: () => import('@/views/skill/index.vue'),
          meta: { title: '已安装技能' }
        },
        {
          path: 'channels',
          name: 'ChannelList',
          component: () => import('@/views/channel/index.vue'),
          meta: { title: '渠道管理' }
        },
        {
          path: 'channels/wechat',
          name: 'ChannelWechat',
          component: () => import('@/views/channel/index.vue'),
          meta: { title: '微信渠道' }
        },
        {
          path: 'analytics',
          name: 'Analytics',
          component: () => import('@/views/analytics/index.vue'),
          meta: { title: '统计分析' }
        },
        {
          path: 'analytics/overview',
          name: 'AnalyticsOverview',
          component: () => import('@/views/analytics/index.vue'),
          meta: { title: '概览' }
        },
        {
          path: 'analytics/conversations',
          name: 'AnalyticsConversations',
          component: () => import('@/views/analytics/conversations.vue'),
          meta: { title: '对话分析' }
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '系统设置' }
        },
        {
          path: 'settings/team',
          name: 'SettingsTeam',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '团队管理' }
        },
        {
          path: 'billing',
          name: 'Billing',
          component: () => import('@/views/billing/index.vue'),
          meta: { title: '套餐与计费' }
        },
        {
          path: 'settings/billing',
          name: 'SettingsBilling',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '套餐与账单' }
        },
        {
          path: 'settings/profile',
          name: 'SettingsProfile',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '个人资料' }
        },
        {
          path: 'settings/payment',
          name: 'SettingsPayment',
          component: () => import('@/views/settings/payment.vue'),
          meta: { title: '支付配置', requiredRole: 'admin' }
        },
        {
          path: 'ai-config',
          name: 'AIConfig',
          component: () => import('@/views/ai-config/index.vue'),
          meta: { title: 'AI 模型配置' }
        },
        {
          path: 'ai-billing',
          name: 'AIBilling',
          component: () => import('@/views/ai-billing/index.vue'),
          meta: { title: 'AI 计费' }
        },
        {
          path: 'workflows',
          name: 'WorkflowList',
          component: () => import('@/views/workflow/index.vue'),
          meta: { title: '工作流管理' }
        },
        {
          path: 'workflows/create',
          name: 'WorkflowCreate',
          component: () => import('@/views/workflow/editor.vue'),
          meta: { title: '创建工作流' }
        },
        {
          path: 'workflows/:id',
          name: 'WorkflowEdit',
          component: () => import('@/views/workflow/editor.vue'),
          meta: { title: '编辑工作流' }
        }
      ]
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/'
    }
  ]
})

export default router
