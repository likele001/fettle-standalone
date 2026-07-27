import { createRouter, createWebHistory } from 'vue-router'
import { setupGuards } from './guards'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('@/views/login/index.vue'),
      meta: { public: true }
    },
    {
      path: '/',
      component: () => import('@/components/Layout.vue'),
      meta: { requiresAuth: true },
      children: [
        {
          path: '',
          name: 'Dashboard',
          component: () => import('@/views/dashboard/index.vue'),
          meta: { title: '平台概览' }
        },
        {
          path: 'tenants',
          name: 'Tenants',
          component: () => import('@/views/tenants/index.vue'),
          meta: { title: '租户管理' }
        },
        {
          path: 'tenants/:id',
          name: 'TenantDetail',
          component: () => import('@/views/tenants/detail.vue'),
          meta: { title: '租户详情' }
        },
        {
          path: 'plans',
          name: 'Plans',
          component: () => import('@/views/plans/index.vue'),
          meta: { title: '套餐管理' }
        },
        {
          path: 'billing',
          name: 'Billing',
          component: () => import('@/views/billing/index.vue'),
          meta: { title: '收入账单' }
        },
        {
          path: 'settings/payment',
          name: 'SettingsPayment',
          component: () => import('@/views/settings/payment.vue'),
          meta: { title: '支付配置' }
        },
        {
          path: 'monitor',
          name: 'Monitor',
          component: () => import('@/views/monitor/index.vue'),
          meta: { title: '系统监控' }
        },
        {
          path: 'settings',
          name: 'Settings',
          component: () => import('@/views/settings/index.vue'),
          meta: { title: '平台配置' }
        },
        {
          path: 'admins',
          name: 'Admins',
          component: () => import('@/views/admins/index.vue'),
          meta: { title: '管理员账号' }
        },
        {
          path: 'ai-models',
          name: 'AIModels',
          component: () => import('@/views/ai-models/index.vue'),
          meta: { title: 'AI 模型管理' }
        },
        {
          path: 'agents',
          name: 'Agents',
          component: () => import('@/views/agents/index.vue'),
          meta: { title: '智能体管理' }
        },
        {
          path: 'skills',
          name: 'Skills',
          component: () => import('@/views/skills/index.vue'),
          meta: { title: '技能管理' }
        },
        {
          path: 'channels',
          name: 'Channels',
          component: () => import('@/views/channels/index.vue'),
          meta: { title: '渠道管理' }
        },
        {
          path: 'ai-pricing',
          name: 'AIPricing',
          component: () => import('@/views/ai-pricing/index.vue'),
          meta: { title: '模型定价' }
        },
        {
          path: 'ai-balances',
          name: 'AIBalances',
          component: () => import('@/views/ai-balances/index.vue'),
          meta: { title: '租户余额' }
        },
        {
          path: 'ai-packages',
          name: 'AIPackages',
          component: () => import('@/views/ai-packages/index.vue'),
          meta: { title: '资源包' }
        },
        {
          path: 'workflows',
          name: 'AdminWorkflowList',
          component: () => import('@/views/workflows/index.vue'),
          meta: { title: '工作流管理' }
        },
        {
          path: 'mcp-servers',
          name: 'MCPServers',
          component: () => import('@/views/mcp-servers/index.vue'),
          meta: { title: 'MCP 服务器管理' }
        }
      ]
    }
  ]
})

setupGuards(router)

export default router
