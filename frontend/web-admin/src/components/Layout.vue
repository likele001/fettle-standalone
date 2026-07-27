<template>
  <div class="layout">
    <!-- 桌面侧边栏 -->
    <aside v-show="!isNarrow" class="sidebar" :class="{ collapsed }">
      <div class="sidebar-brand">
        <h1 v-show="!collapsed">AI智能体平台</h1>
        <h1 v-show="collapsed" class="brand-mini">AI</h1>
      </div>
      <SidebarMenu :collapsed="collapsed" mode="sidebar" />
    </aside>

    <!-- 移动端抽屉菜单 -->
    <el-drawer v-model="drawerOpen" direction="ltr" size="260px" :with-header="false" append-to-body class="mobile-drawer">
      <div class="drawer-brand">AI智能体平台</div>
      <SidebarMenu :collapsed="false" mode="drawer" />
    </el-drawer>

    <!-- 主区域 -->
    <div class="main-area">
      <header class="main-header">
        <div class="header-left">
          <el-button v-if="isNarrow" text class="menu-btn" @click="drawerOpen = true">
            <el-icon :size="22"><Menu /></el-icon>
          </el-button>
          <el-button v-else text class="menu-btn hidden-mobile" @click="toggleCollapse">
            <el-icon :size="20"><Fold v-if="!collapsed" /><Expand v-else /></el-icon>
          </el-button>
          <el-breadcrumb v-if="!isNarrow" separator="/" class="breadcrumb">
            <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
            <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
              {{ item.title }}
            </el-breadcrumb-item>
          </el-breadcrumb>
          <span v-else class="page-title-mobile">{{ currentTitle }}</span>
        </div>
        <div class="header-right">
          <el-badge :value="notifications" :hidden="notifications === 0" class="notif-badge">
            <el-icon class="header-icon" :size="18"><Bell /></el-icon>
          </el-badge>
          <el-dropdown @command="handleCommand" trigger="click">
            <div class="user-info">
              <el-avatar :size="isNarrow ? 28 : 32" :src="userStore.userInfo?.avatar_url">
                {{ userStore.userInfo?.name?.charAt(0) || 'U' }}
              </el-avatar>
              <span v-show="!isNarrow" class="username">{{ userStore.userInfo?.name || '用户' }}</span>
              <el-icon v-show="!isNarrow" class="el-icon--right"><ArrowDown /></el-icon>
            </div>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人信息</el-dropdown-item>
                <el-dropdown-item command="settings">系统设置</el-dropdown-item>
                <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>
      <main class="main-content">
        <router-view />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Bell, ArrowDown, Fold, Expand, Menu } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'
import SidebarMenu from './SidebarMenu.vue'

const NARROW_QUERY = '(max-width: 1023px)'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const collapsed = ref(false)
const isNarrow = ref(false)
const drawerOpen = ref(false)
const notifications = ref(0)

function toggleCollapse() {
  collapsed.value = !collapsed.value
}

const currentTitle = computed(() => {
  const matched = route.matched.filter(m => m.meta?.title)
  return matched.length > 0 ? matched[matched.length - 1].meta.title as string : ''
})

const breadcrumbs = computed(() => {
  return route.matched.filter(m => m.meta?.title).map(m => ({
    path: m.path,
    title: m.meta.title as string
  }))
})

function handleCommand(cmd: string) {
  switch (cmd) {
    case 'profile': router.push('/settings/profile'); break
    case 'settings': router.push('/settings'); break
    case 'logout':
      userStore.clearUser()
      ElMessage.success('已退出登录')
      router.push('/login')
      break
  }
}

let mq: MediaQueryList | null = null
function syncNarrow() {
  isNarrow.value = mq?.matches ?? false
  if (!isNarrow.value) drawerOpen.value = false
}

watch(() => route.path, () => {
  if (isNarrow.value) drawerOpen.value = false
})

onMounted(() => {
  mq = window.matchMedia(NARROW_QUERY)
  syncNarrow()
  mq.addEventListener('change', syncNarrow)
})

onUnmounted(() => {
  mq?.removeEventListener('change', syncNarrow)
})
</script>

<style scoped lang="scss">
@import '@/styles/variables.scss';

.layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* 桌面侧边栏 */
.sidebar {
  width: 220px;
  background: #001529;
  display: flex;
  flex-direction: column;
  transition: width 0.2s;
  overflow: hidden;
  flex-shrink: 0;

  &.collapsed {
    width: 64px;
  }

  .sidebar-brand {
    height: 56px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: #002140;
    flex-shrink: 0;

    h1 {
      color: #fff;
      font-size: 16px;
      white-space: nowrap;
      margin: 0;
    }

    .brand-mini {
      font-size: 20px;
      font-weight: 700;
    }
  }
}

/* 抽屉品牌 */
.drawer-brand {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #002140;
  color: #fff;
  font-size: 16px;
  font-weight: 600;
}

/* 主区域 */
.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
}

.main-header {
  height: 52px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 12px;
  background: #fff;
  border-bottom: 1px solid #ebeef5;
  flex-shrink: 0;
  gap: 12px;

  @media (min-width: 640px) {
    padding: 0 20px;
  }

  .header-left {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;
    flex: 1;

    .menu-btn {
      flex-shrink: 0;
    }

    .breadcrumb {
      font-size: 13px;
      min-width: 0;
      overflow: hidden;
      white-space: nowrap;
    }

    .page-title-mobile {
      font-size: 14px;
      font-weight: 600;
      color: #303133;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;

    .notif-badge {
      cursor: pointer;

      .header-icon {
        color: #606266;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 6px;
      cursor: pointer;

      .username {
        font-size: 13px;
        color: #303133;
      }
    }
  }
}

.main-content {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  background: $bg-color;

  @media (min-width: 640px) {
    padding: 24px;
  }
}

.hidden-mobile {
  @media (max-width: 1023px) {
    display: none;
  }
}
</style>

<style lang="scss">
/* 抽屉菜单样式修复 */
.mobile-drawer {
  .el-drawer__body {
    padding: 0;
  }

  .el-menu {
    border-right: none;
  }
}
</style>
