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
    <el-menu-item index="/">
      <el-icon><HomeFilled /></el-icon>
      <span>控制台</span>
    </el-menu-item>

    <el-sub-menu index="agents">
      <template #title>
        <el-icon><UserFilled /></el-icon>
        <span>智能体管理</span>
      </template>
      <el-menu-item index="/agents">智能体列表</el-menu-item>
      <el-menu-item index="/agents/create">创建智能体</el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="conversations">
      <template #title>
        <el-icon><ChatDotRound /></el-icon>
        <span>会话管理</span>
      </template>
      <el-menu-item index="/conversations">会话列表</el-menu-item>
      <el-menu-item index="/conversations/monitor">实时监控</el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="knowledge">
      <template #title>
        <el-icon><FolderOpened /></el-icon>
        <span>知识库</span>
      </template>
      <el-menu-item index="/knowledge">知识库列表</el-menu-item>
      <el-menu-item index="/knowledge/upload">上传文档</el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="skills">
      <template #title>
        <el-icon><MagicStick /></el-icon>
        <span>技能市场</span>
      </template>
      <el-menu-item index="/skills">技能列表</el-menu-item>
      <el-menu-item index="/skills/installed">已安装技能</el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="channels">
      <template #title>
        <el-icon><Connection /></el-icon>
        <span>渠道接入</span>
      </template>
      <el-menu-item index="/channels">渠道列表</el-menu-item>
      <el-menu-item index="/channels/wechat">微信配置</el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="analytics">
      <template #title>
        <el-icon><DataAnalysis /></el-icon>
        <span>数据分析</span>
      </template>
      <el-menu-item index="/analytics/overview">数据概览</el-menu-item>
      <el-menu-item index="/analytics/conversations">会话分析</el-menu-item>
    </el-sub-menu>

    <el-menu-item index="/ai-config">
      <el-icon><Cpu /></el-icon>
      <span>AI 模型配置</span>
    </el-menu-item>

    <el-menu-item index="/ai-billing">
      <el-icon><Coin /></el-icon>
      <span>AI 计费</span>
    </el-menu-item>

    <el-sub-menu index="workflows">
      <template #title>
        <el-icon><Operation /></el-icon>
        <span>工作流管理</span>
      </template>
      <el-menu-item index="/workflows">工作流列表</el-menu-item>
      <el-menu-item index="/workflows/create">创建工作流</el-menu-item>
    </el-sub-menu>

    <el-sub-menu index="settings">
      <template #title>
        <el-icon><Setting /></el-icon>
        <span>系统设置</span>
      </template>
      <el-menu-item index="/settings/team">团队管理</el-menu-item>
      <el-menu-item index="/settings/profile">个人信息</el-menu-item>
    </el-sub-menu>
  </el-menu>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import {
  HomeFilled, Cpu, Coin, ChatDotRound, FolderOpened, MagicStick,
  Connection, DataAnalysis, Setting, Operation, UserFilled
} from '@element-plus/icons-vue'

const props = withDefaults(defineProps<{
  collapsed?: boolean
  mode?: 'sidebar' | 'drawer'
}>(), {
  collapsed: false,
  mode: 'sidebar'
})

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style scoped lang="scss">
.sidebar-menu {
  border-right: none;
  overflow-y: auto;
  overflow-x: hidden;
  flex: 1;
}
</style>
