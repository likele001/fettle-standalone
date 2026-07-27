<template>
  <div class="page-container">
    <div class="page-header">
      <h2>技能管理</h2>
    </div>

    <!-- Stats Cards -->
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="8">
        <div class="stat-card">
          <div class="stat-value">{{ skillStats.total || 0 }}</div>
          <div class="stat-label">总技能数</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card">
          <div class="stat-value">{{ skillStats.installed || 0 }}</div>
          <div class="stat-label">已安装</div>
        </div>
      </el-col>
      <el-col :span="8">
        <div class="stat-card">
          <div class="stat-value">{{ skillStats.total_installations || 0 }}</div>
          <div class="stat-label">总安装次数</div>
        </div>
      </el-col>
    </el-row>

    <!-- Skills Table -->
    <el-card>
      <el-table :data="skills" stripe v-loading="loading">
        <el-table-column prop="name" label="名称" width="200" />
        <el-table-column prop="description" label="描述" show-overflow-tooltip />
        <el-table-column prop="version" label="版本" width="100" />
        <el-table-column label="安装次数" width="120">
          <template #default="{ row }">
            {{ row.installation_count || 0 }}
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDate(row.created_at) }}
          </template>
        </el-table-column>
      </el-table>

      <div style="margin-top: 16px; display: flex; justify-content: flex-end">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="loadSkills"
          @current-change="loadSkills"
        />
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getAdminSkills, getAdminSkillStats } from '@/api/admin-skill'

const skills = ref<any[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const loading = ref(false)
const skillStats = ref<any>({})

const formatDate = (date: string) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

async function loadSkills() {
  loading.value = true
  try {
    const res = await getAdminSkills({ page: currentPage.value, page_size: pageSize.value })
    skills.value = res?.items || []
    total.value = res?.total || 0
  } catch {
    skills.value = []
  } finally {
    loading.value = false
  }
}

async function loadStats() {
  try {
    skillStats.value = await getAdminSkillStats()
  } catch {
    skillStats.value = {}
  }
}

onMounted(() => {
  loadSkills()
  loadStats()
})
</script>

<style scoped lang="scss">
.stat-card {
  background: #fff;
  border-radius: 8px;
  padding: 24px;
  text-align: center;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);

  .stat-value {
    font-size: 32px;
    font-weight: 700;
    color: #1a1a2e;
    margin-bottom: 8px;
  }

  .stat-label {
    font-size: 14px;
    color: #666;
  }
}
</style>
