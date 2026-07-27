<template>
  <el-header class="header">
    <div class="header-left">
      <el-breadcrumb separator="/">
        <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item v-for="item in breadcrumbs" :key="item.path">
          {{ item.title }}
        </el-breadcrumb-item>
      </el-breadcrumb>
    </div>

    <div class="header-right">
      <el-badge :value="notifications" :hidden="notifications === 0" class="notification-badge">
        <el-icon class="header-icon"><Bell /></el-icon>
      </el-badge>

      <el-dropdown @command="handleCommand">
        <div class="user-info">
          <el-avatar :size="32" :src="userStore.userInfo?.avatar_url">
            {{ userStore.userInfo?.name?.charAt(0) || 'U' }}
          </el-avatar>
          <span class="username">{{ userStore.userInfo?.name || '用户' }}</span>
          <el-icon class="el-icon--right"><arrow-down /></el-icon>
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
  </el-header>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { Bell, ArrowDown } from '@element-plus/icons-vue'
import { useUserStore } from '@/stores/user'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const notifications = ref(0)

async function loadNotifications() {
  try {
    const { getUnreadCount } = await import('@/api/analytics')
    const res = await getUnreadCount() as any
    notifications.value = res?.count || 0
  } catch { /* ignore */ }
}

const breadcrumbs = computed(() => {
  const matched = route.matched.filter(item => item.meta?.title)
  return matched.map(item => ({
    path: item.path,
    title: item.meta.title as string
  }))
})

function handleCommand(command: string) {
  switch (command) {
    case 'profile':
      router.push('/settings/profile')
      break
    case 'settings':
      router.push('/settings')
      break
    case 'logout':
      userStore.clearUser()
      ElMessage.success('已退出登录')
      router.push('/login')
      break
  }
}
</script>

<style scoped lang="scss">
.header {
  background-color: #fff;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.08);

  .header-left {
    .el-breadcrumb {
      font-size: 14px;
    }
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 20px;

    .notification-badge {
      cursor: pointer;

      .header-icon {
        font-size: 20px;
        color: #606266;
      }
    }

    .user-info {
      display: flex;
      align-items: center;
      gap: 8px;
      cursor: pointer;

      .username {
        font-size: 14px;
        color: #303133;
      }
    }
  }
}
</style>
