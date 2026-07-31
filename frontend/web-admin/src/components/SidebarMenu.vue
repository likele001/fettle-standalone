<template>
  <el-menu
    :default-active="activeMenu"
    :collapse="collapsed && mode === 'sidebar'"
    router
    class="sidebar-menu"
    background-color="#001529"
    text-color="#fff"
    active-text-color="#409EFF"
  >
    <template v-for="item in filteredMenu" :key="item.path || item.group">
      <el-menu-item v-if="item.path" :index="item.path">
        <el-icon><component :is="item.icon" /></el-icon>
        <span>{{ item.title }}</span>
      </el-menu-item>

      <el-sub-menu v-else :index="item.group">
        <template #title>
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.group }}</span>
        </template>
        <el-menu-item
          v-for="child in item.children"
          :key="child.path"
          :index="child.path"
        >
          {{ child.title }}
        </el-menu-item>
      </el-sub-menu>
    </template>
  </el-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  HomeFilled, Cpu, Coin, ChatDotRound, FolderOpened, MagicStick,
  Connection, DataAnalysis, Setting, Operation, UserFilled
} from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const props = withDefaults(defineProps<{
  collapsed?: boolean
  mode?: 'sidebar' | 'drawer'
}>(), {
  collapsed: false,
  mode: 'sidebar'
})

const route = useRoute()
const activeMenu = computed(() => route.path)
const userStore = useUserStore()

interface MenuChild { path: string; title: string; perm: string }
interface MenuItem {
  path?: string
  title?: string
  group?: string
  icon: any
  children?: MenuChild[]
  perm: string
}

const allMenu: MenuItem[] = [
  { path: '/', title: '控制台', icon: HomeFilled, perm: '' },
  { group: '智能体管理', icon: UserFilled, perm: 'agents.read', children: [
    { path: '/agents', title: '智能体列表', perm: 'agents.read' },
    { path: '/agents/create', title: '创建智能体', perm: 'agents.create' },
  ]},
  { group: '会话管理', icon: ChatDotRound, perm: 'conversation.read', children: [
    { path: '/conversations', title: '会话列表', perm: 'conversation.read' },
    { path: '/conversations/monitor', title: '实时监控', perm: 'conversation.manage' },
  ]},
  { group: '知识库', icon: FolderOpened, perm: 'knowledge.read', children: [
    { path: '/knowledge', title: '知识库列表', perm: 'knowledge.read' },
    { path: '/knowledge/upload', title: '上传文档', perm: 'knowledge.create' },
  ]},
  { group: '技能市场', icon: MagicStick, perm: 'skills.read', children: [
    { path: '/skills', title: '技能列表', perm: 'skills.read' },
    { path: '/skills/installed', title: '已安装技能', perm: 'skills.manage' },
  ]},
  { group: '渠道接入', icon: Connection, perm: 'team.manage', children: [
    { path: '/channels', title: '渠道列表', perm: 'team.manage' },
    { path: '/channels/wechat', title: '微信配置', perm: 'team.manage' },
  ]},
  { group: '数据分析', icon: DataAnalysis, perm: 'analytics.read', children: [
    { path: '/analytics/overview', title: '数据概览', perm: 'analytics.read' },
    { path: '/analytics/conversations', title: '会话分析', perm: 'analytics.read' },
  ]},
  { group: '工作流管理', icon: Operation, perm: 'workflow.read', children: [
    { path: '/workflows', title: '工作流列表', perm: 'workflow.read' },
    { path: '/workflows/create', title: '创建工作流', perm: 'workflow.create' },
  ]},
  { path: '/ai-config', title: 'AI 模型配置', icon: Cpu, perm: 'settings.manage' },
  { path: '/ai-billing', title: 'AI 计费', icon: Coin, perm: 'billing.read' },
  { group: '系统设置', icon: Setting, perm: 'settings.read', children: [
    { path: '/settings/team', title: '团队管理', perm: 'team.manage' },
    { path: '/settings/roles', title: '角色管理', perm: 'role.manage' },
    { path: '/settings/profile', title: '个人信息', perm: 'team.read' },
    { path: '/settings/billing', title: '套餐与账单', perm: 'billing.read' },
    { path: '/settings/payment', title: '支付配置', perm: 'billing.manage' },
  ]},
]

function canAccess(perm: string) {
  return !perm || userStore.hasPermission?.(perm)
}

const filteredMenu = computed(() => {
  return allMenu.filter(item => {
    if (item.path) return canAccess(item.perm)
    // Group: show only if user can see at least one child
    const visible = (item.children || []).filter(c => canAccess(c.perm))
    if (visible.length === 0) return false
    item.children = visible
    return true
  })
})
</script>

<style scoped lang="scss">
.sidebar-menu {
  border-right: none;
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1;
}
</style>
