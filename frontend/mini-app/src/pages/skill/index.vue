<template>
  <view class="skill-page">
    <view class="search-bar">
      <input
        v-model="keyword"
        type="text"
        placeholder="搜索技能"
        class="search-input"
        @input="handleSearch"
        @confirm="handleSearch"
      />
    </view>

    <scroll-view class="category-scroll" scroll-x :show-scrollbar="false">
      <view
        v-for="cat in categories"
        :key="cat.value"
        class="category-item"
        :class="{ active: selectedCategory === cat.value }"
        @click="selectCategory(cat.value)"
      >
        {{ cat.label }}
      </view>
    </scroll-view>

    <scroll-view class="skill-list" scroll-y :show-scrollbar="false">
      <view
        v-for="skill in skills"
        :key="skill.id"
        class="skill-card"
      >
        <view class="skill-header">
          <image class="skill-icon" :src="skill.icon || '/static/default-skill.png'" mode="aspectFill" />
          <view class="skill-info">
            <view class="skill-name">{{ skill.name }}</view>
            <view class="skill-meta">
              <text class="meta-item">{{ skill.author }}</text>
              <text class="meta-divider">·</text>
              <text class="meta-item">v{{ skill.version }}</text>
              <text class="meta-divider">·</text>
              <text class="meta-item">{{ skill.install_count }}人安装</text>
            </view>
          </view>
        </view>
        <view class="skill-desc">{{ skill.description }}</view>
        <view class="skill-footer">
          <view class="skill-category">{{ getCategoryLabel(skill.category) }}</view>
          <button
            class="action-btn"
            :class="{ installed: skill.is_installed }"
            @click="handleInstall(skill)"
          >
            {{ skill.is_installed ? '已安装' : '安装' }}
          </button>
        </view>
      </view>

      <view v-if="skills.length === 0 && !loading" class="empty-state">
        <image class="empty-icon" src="/static/empty.png" mode="aspectFit" />
        <text class="empty-text">暂无技能</text>
      </view>

      <view v-if="loading" class="loading-wrapper">
        <text class="loading-text">加载中...</text>
      </view>
    </scroll-view>
  </view>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getSkills, installSkill, uninstallSkill, type Skill } from '@/api/skill'
import { useUserStore } from '@/store/user'

const userStore = useUserStore()

const skills = ref<Skill[]>([])
const loading = ref(false)
const keyword = ref('')
const selectedCategory = ref('')
const page = ref(1)
const pageSize = ref(20)

const categories = [
  { label: '全部', value: '' },
  { label: '电商', value: 'ecommerce' },
  { label: '营销', value: 'marketing' },
  { label: '客服', value: 'customer_service' },
  { label: '开发', value: 'development' },
  { label: '数据分析', value: 'data_analysis' },
  { label: '内容创作', value: 'content' },
  { label: '办公效率', value: 'productivity' }
]

onMounted(() => {
  if (!userStore.isLoggedIn) {
    uni.redirectTo({ url: '/pages/login/index' })
    return
  }
  loadSkills()
})

async function loadSkills() {
  loading.value = true
  try {
    const res = await getSkills({
      page: page.value,
      page_size: pageSize.value,
      keyword: keyword.value,
      category: selectedCategory.value
    })
    skills.value = res.items
  } catch (error) {
    uni.showToast({ title: '加载失败', icon: 'none' })
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  page.value = 1
  loadSkills()
}

function selectCategory(category: string) {
  selectedCategory.value = category
  page.value = 1
  loadSkills()
}

function getCategoryLabel(category: string) {
  const cat = categories.find(c => c.value === category)
  return cat?.label || category
}

async function handleInstall(skill: Skill) {
  if (skill.is_installed) {
    try {
      await uninstallSkill(skill.id)
      skill.is_installed = false
      uni.showToast({ title: '已卸载', icon: 'success' })
    } catch {
      uni.showToast({ title: '卸载失败', icon: 'none' })
    }
  } else {
    try {
      await installSkill(skill.id)
      skill.is_installed = true
      skill.install_count++
      uni.showToast({ title: '安装成功', icon: 'success' })
    } catch {
      uni.showToast({ title: '安装失败', icon: 'none' })
    }
  }
}
</script>

<style lang="scss" scoped>
.skill-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.search-bar {
  padding: 20rpx;
  background: #fff;
  position: sticky;
  top: 0;
  z-index: 10;

  .search-input {
    width: 100%;
    height: 72rpx;
    background: #f5f7fa;
    border-radius: 36rpx;
    padding: 0 32rpx;
    font-size: 28rpx;
    box-sizing: border-box;
  }
}

.category-scroll {
  white-space: nowrap;
  background: #fff;
  padding: 16rpx 20rpx;
  border-bottom: 1rpx solid #f0f0f0;

  .category-item {
    display: inline-block;
    padding: 12rpx 32rpx;
    margin-right: 16rpx;
    background: #f5f7fa;
    border-radius: 24rpx;
    font-size: 26rpx;
    color: #666;

    &.active {
      background: #409eff;
      color: #fff;
    }
  }
}

.skill-list {
  height: calc(100vh - 200rpx);
  padding: 20rpx;
}

.skill-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.06);
}

.skill-header {
  display: flex;
  align-items: center;
  margin-bottom: 16rpx;

  .skill-icon {
    width: 80rpx;
    height: 80rpx;
    border-radius: 16rpx;
    margin-right: 20rpx;
    flex-shrink: 0;
  }

  .skill-info {
    flex: 1;

    .skill-name {
      font-size: 32rpx;
      font-weight: 600;
      color: #333;
      margin-bottom: 8rpx;
    }

    .skill-meta {
      display: flex;
      align-items: center;
      font-size: 24rpx;
      color: #999;

      .meta-item {
        margin-right: 4rpx;
      }

      .meta-divider {
        margin: 0 8rpx;
      }
    }
  }
}

.skill-desc {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
  margin-bottom: 16rpx;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.skill-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;

  .skill-category {
    font-size: 24rpx;
    color: #409eff;
    background: #ecf5ff;
    padding: 6rpx 16rpx;
    border-radius: 8rpx;
  }

  .action-btn {
    padding: 12rpx 32rpx;
    background: #409eff;
    color: #fff;
    border: none;
    border-radius: 24rpx;
    font-size: 26rpx;

    &.installed {
      background: #f0f0f0;
      color: #999;
    }
  }
}

.empty-state {
  text-align: center;
  padding: 120rpx 0;

  .empty-icon {
    width: 200rpx;
    height: 200rpx;
    margin-bottom: 24rpx;
  }

  .empty-text {
    font-size: 28rpx;
    color: #999;
  }
}

.loading-wrapper {
  text-align: center;
  padding: 40rpx 0;

  .loading-text {
    font-size: 28rpx;
    color: #999;
  }
}
</style>
