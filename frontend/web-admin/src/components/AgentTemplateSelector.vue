<template>
  <div class="template-selector">
    <div class="category-tabs">
      <el-tag
        v-for="cat in categories"
        :key="cat.id"
        :type="selectedCategory === cat.id ? 'primary' : ''"
        :effect="selectedCategory === cat.id ? 'dark' : 'plain'"
        @click="selectedCategory = cat.id"
        class="category-tag"
      >
        {{ cat.name }}
      </el-tag>
    </div>

    <div class="template-grid">
      <div
        v-for="template in filteredTemplates"
        :key="template.id"
        :class="['template-card', { active: selectedTemplate?.id === template.id }]"
        @click="selectTemplate(template)"
      >
        <div class="template-icon" :style="{ background: template.color + '15', color: template.color }">
          <el-icon :size="28">
            <component :is="getIcon(template.icon)" />
          </el-icon>
        </div>
        <h3>{{ template.name }}</h3>
        <p>{{ template.description }}</p>
        <div class="template-tags">
          <el-tag v-for="tag in template.tags.slice(0, 3)" :key="tag" size="small" effect="plain">
            {{ tag }}
          </el-tag>
        </div>
        <div v-if="selectedTemplate?.id === template.id" class="selected-badge">
          <el-icon><Check /></el-icon>
        </div>
      </div>
    </div>

    <div class="bottom-actions">
      <el-button @click="$emit('clear')">创建空白智能体</el-button>
      <el-button type="primary" @click="confirmSelect" :disabled="!selectedTemplate">
        <el-icon><ArrowRight /></el-icon>
        使用此模板创建
      </el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Headset, EditPen, DataAnalysis, Service,
  Document, TrendCharts, User, Check, ArrowRight
} from '@element-plus/icons-vue'
import { agentTemplates, templateCategories, type AgentTemplate } from '@/data/agentTemplates'

const emit = defineEmits<{
  select: [template: AgentTemplate]
  clear: []
}>()

const selectedCategory = ref('all')
const selectedTemplate = ref<AgentTemplate | null>(null)

const categories = templateCategories
const filteredTemplates = computed(() => {
  if (selectedCategory.value === 'all') {
    return agentTemplates
  }
  return agentTemplates.filter(t => t.category === selectedCategory.value)
})

const iconMap: Record<string, any> = {
  'headset': Headset,
  'code': EditPen,
  'pen-tool': EditPen,
  'bar-chart': DataAnalysis,
  'globe': Service,
  'document': Document,
  'trending-up': TrendCharts,
  'users': User
}

const getIcon = (icon: string) => {
  return iconMap[icon] || Headset
}

const selectTemplate = (template: AgentTemplate) => {
  selectedTemplate.value = template
}

const confirmSelect = () => {
  if (selectedTemplate.value) {
    emit('select', selectedTemplate.value)
  }
}
</script>

<style scoped lang="scss">
.template-selector {
  padding: 16px 0;
}

.category-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 20px;

  .category-tag {
    cursor: pointer;
    transition: all 0.2s;
    padding: 4px 14px;
  }
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
  max-height: 400px;
  overflow-y: auto;
  padding-right: 8px;
}

.template-card {
  position: relative;
  padding: 20px;
  border: 2px solid #e8e8e8;
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s;
  background: #fff;

  &:hover {
    border-color: #d9d9d9;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  }

  &.active {
    border-color: #409EFF;
    background: #f0f5ff;
  }

  .template-icon {
    width: 56px;
    height: 56px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 12px;
  }

  h3 {
    margin: 0 0 8px 0;
    font-size: 16px;
    font-weight: 600;
    color: #333;
  }

  p {
    margin: 0 0 12px 0;
    font-size: 13px;
    color: #666;
    line-height: 1.5;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .template-tags {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
  }

  .selected-badge {
    position: absolute;
    top: 12px;
    right: 12px;
    width: 24px;
    height: 24px;
    background: #409EFF;
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
}

.bottom-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 20px;
  padding-top: 16px;
  border-top: 1px solid #f0f0f0;
}
</style>