<template>
  <div class="header">
    <div class="header-left">
      <span class="page-title">{{ pageTitle }}</span>
    </div>
    <div class="header-right">
      <el-dropdown trigger="click">
        <div class="user-info">
          <el-icon :size="20"><UserFilled /></el-icon>
          <span class="user-name">{{ userStore.userInfo?.name || '管理员' }}</span>
          <el-icon><ArrowDown /></el-icon>
        </div>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="handleLogout">
              <el-icon><SwitchButton /></el-icon>
              退出登录
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { logout as logoutApi } from '@/api/auth'
import { UserFilled, ArrowDown, SwitchButton } from '@element-plus/icons-vue'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const pageTitle = computed(() => route.meta.title || '管理后台')

async function handleLogout() {
  try { await logoutApi() } catch { /* ignore */ }
  userStore.clearUser()
  router.push('/login')
}
</script>

<style scoped lang="scss">
.header {
  height: $header-height;
  background: $bg-header;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  border-bottom: 1px solid #e2e8f0;
  flex-shrink: 0;

  .page-title {
    font-size: 16px;
    font-weight: 600;
    color: #1e293b;
  }

  .user-info {
    display: flex;
    align-items: center;
    gap: 6px;
    cursor: pointer;
    padding: 6px 10px;
    border-radius: 6px;
    transition: background 0.2s;

    &:hover {
      background: #f1f5f9;
    }

    .user-name {
      font-size: 14px;
      color: #475569;
    }
  }
}
</style>
