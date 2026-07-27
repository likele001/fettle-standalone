<template>
  <div class="sidebar">
    <div class="logo">
      <span class="logo-text">Fettle</span>
      <span class="logo-badge">Admin</span>
    </div>
    <el-menu
      :default-active="activeMenu"
      router
      background-color="#1e293b"
      text-color="#94a3b8"
      active-text-color="#fff"
    >
      <el-menu-item index="/">
        <el-icon><DataBoard /></el-icon>
        <span>平台概览</span>
      </el-menu-item>

      <el-sub-menu index="business">
        <template #title>
          <el-icon><OfficeBuilding /></el-icon>
          <span>业务管理</span>
        </template>
        <el-menu-item index="/tenants">租户管理</el-menu-item>
        <el-menu-item index="/plans">套餐管理</el-menu-item>
        <el-menu-item index="/agents">智能体</el-menu-item>
        <el-menu-item index="/skills">技能</el-menu-item>
        <el-menu-item index="/channels">渠道</el-menu-item>
        <el-menu-item index="/billing">收入账单</el-menu-item>
        <el-menu-item index="/settings/payment">支付配置</el-menu-item>
      </el-sub-menu>

      <el-sub-menu index="ai">
        <template #title>
          <el-icon><Cpu /></el-icon>
          <span>AI 管理</span>
        </template>
        <el-menu-item index="/ai-models">模型管理</el-menu-item>
        <el-menu-item index="/ai-pricing">模型定价</el-menu-item>
        <el-menu-item index="/ai-balances">租户余额</el-menu-item>
        <el-menu-item index="/ai-packages">资源包</el-menu-item>
        <el-menu-item index="/workflows">工作流管理</el-menu-item>
        <el-menu-item index="/mcp-servers">MCP 服务器</el-menu-item>
      </el-sub-menu>

      <el-sub-menu index="system">
        <template #title>
          <el-icon><Setting /></el-icon>
          <span>系统管理</span>
        </template>
        <el-menu-item index="/monitor">系统监控</el-menu-item>
        <el-menu-item index="/admins">管理员账号</el-menu-item>
        <el-menu-item index="/settings">平台配置</el-menu-item>
      </el-sub-menu>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { DataBoard, OfficeBuilding, Setting, Cpu } from '@element-plus/icons-vue'

const route = useRoute()
const activeMenu = computed(() => {
  const path = route.path
  if (path.startsWith('/tenants')) return '/tenants'
  if (path.startsWith('/plans')) return '/plans'
  if (path.startsWith('/billing')) return '/billing'
  if (path.startsWith('/settings/payment')) return '/settings/payment'
  if (path.startsWith('/monitor')) return '/monitor'
  if (path.startsWith('/admins')) return '/admins'
  if (path.startsWith('/agents')) return '/agents'
  if (path.startsWith('/skills')) return '/skills'
  if (path.startsWith('/channels')) return '/channels'
  if (path.startsWith('/ai-pricing')) return '/ai-pricing'
  if (path.startsWith('/ai-balances')) return '/ai-balances'
  if (path.startsWith('/ai-packages')) return '/ai-packages'
  if (path.startsWith('/workflows')) return '/workflows'
  if (path.startsWith('/mcp-servers')) return '/mcp-servers'
  if (path.startsWith('/settings')) return '/settings'
  return '/'
})
</script>

<style scoped lang="scss">
.sidebar {
  width: $sidebar-width;
  min-width: $sidebar-width;
  background: $bg-sidebar;
  display: flex;
  flex-direction: column;

  .logo {
    padding: 18px 20px;
    display: flex;
    align-items: center;
    gap: 8px;
    border-bottom: 1px solid #334155;

    .logo-text {
      font-size: 18px;
      font-weight: 700;
      color: #fff;
    }

    .logo-badge {
      font-size: 10px;
      background: $admin-primary;
      color: #fff;
      padding: 2px 6px;
      border-radius: 4px;
      font-weight: 600;
    }
  }

  .el-menu {
    border-right: none;
    flex: 1;
  }

  .el-menu-item {
    &.is-active {
      background-color: $bg-sidebar-active !important;
    }
  }
}
</style>
