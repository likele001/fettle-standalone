<template>
  <div class="skill-list">
    <div class="page-header">
      <h2>技能市场</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建技能
      </el-button>
    </div>

    <el-tabs v-model="activeTab">
      <el-tab-pane label="技能市场" name="market">
        <el-row :gutter="20">
          <el-col :xs="24" :sm="12" :md="8" v-for="skill in marketSkills" :key="skill.id">
            <el-card class="skill-card" shadow="hover">
              <div class="skill-header">
                <div class="skill-icon" :style="{ background: getIconBg(skill.category) }">
                  <img v-if="isUrlIcon(skill.icon)" :src="skill.icon" class="skill-icon-img" />
                  <span v-else class="skill-icon-emoji">{{ skill.icon || getCategoryEmoji(skill.category) }}</span>
                </div>
                <div class="skill-info">
                  <h3>{{ skill.name }}</h3>
                  <el-tag :type="getCategoryTagType(skill.category)" size="small">
                    {{ getCategoryLabel(skill.category) }}
                  </el-tag>
                </div>
              </div>
              <p class="skill-desc">{{ skill.description || '暂无描述' }}</p>
              <div class="skill-footer">
                <span class="install-count">{{ skill.install_count || 0 }} 次安装</span>
                <el-button
                  v-if="!skill.installed"
                  type="primary"
                  size="small"
                  @click="installSkill(skill)"
                  :loading="installingId === skill.id"
                >
                  安装
                </el-button>
                <el-tag v-else type="success" size="small" effect="dark">已安装</el-tag>
              </div>
            </el-card>
          </el-col>
        </el-row>
        <el-empty v-if="marketSkills.length === 0 && !loadingMarket" description="暂无可用技能" />
      </el-tab-pane>

      <el-tab-pane label="已安装技能" name="installed">
        <el-table :data="installedSkills" stripe v-loading="loadingInstalled">
          <el-table-column label="名称" width="200">
            <template #default="{ row }">
              <div style="display: flex; align-items: center; gap: 8px;">
                <span class="mini-icon">{{ row.icon || getCategoryEmoji(row.category) }}</span>
                {{ row.name }}
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" show-overflow-tooltip />
          <el-table-column label="分类" width="120">
            <template #default="{ row }">
              <el-tag :type="getCategoryTagType(row.category)" size="small">
                {{ getCategoryLabel(row.category) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
                {{ row.status === 'active' ? '启用' : '禁用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="configureSkill(row)">配置</el-button>
              <el-button size="small" @click="toggleSkill(row)">
                {{ row.status === 'active' ? '禁用' : '启用' }}
              </el-button>
              <el-popconfirm title="确定卸载？" @confirm="uninstallSkill(row.id)">
                <template #reference>
                  <el-button size="small" type="danger">卸载</el-button>
                </template>
              </el-popconfirm>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-if="installedSkills.length === 0 && !loadingInstalled" description="暂未安装任何技能" />
      </el-tab-pane>
    </el-tabs>

    <!-- 创建技能对话框 -->
    <el-dialog v-model="createDialogVisible" title="创建自定义技能" width="600px">
      <el-form :model="createForm" label-width="100px">
        <el-form-item label="技能名称" required>
          <el-input v-model="createForm.name" placeholder="请输入技能名称" />
        </el-form-item>
        <el-form-item label="技能图标">
          <el-input v-model="createForm.icon" placeholder="输入 emoji 如 🌤️ 或图片 URL" />
        </el-form-item>
        <el-form-item label="技能描述" required>
          <el-input v-model="createForm.description" type="textarea" :rows="3" placeholder="描述技能的功能" />
        </el-form-item>
        <el-form-item label="分类">
          <el-select v-model="createForm.category" placeholder="选择分类">
            <el-option label="生活服务" value="life" />
            <el-option label="金融财经" value="finance" />
            <el-option label="实用工具" value="utility" />
            <el-option label="数据处理" value="data" />
            <el-option label="内容生成" value="content" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="技能代码">
          <el-input v-model="createForm.code" type="textarea" :rows="10" placeholder="输入 Python 代码实现" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="createDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>

    <!-- 技能配置对话框 -->
    <el-dialog v-model="configDialogVisible" :title="`配置技能 - ${editingSkill?.name || ''}`" width="600px">
      <el-form :model="{ config: skillConfig }" label-width="100px">
        <el-form-item label="技能配置">
          <el-input
            v-model="skillConfig"
            type="textarea"
            :rows="10"
            placeholder="输入 JSON 格式的技能配置，如：{api_key: xxx}"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleConfigSave">保存配置</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { listSkills, createSkill, installSkill as installSkillApi, uninstallSkill as uninstallSkillApi, toggleSkill as toggleSkillApi, getInstalledSkills, getSkillConfig, updateSkillConfig } from '@/api/skill'
import type { Skill, SkillInstallation } from '@/api/skill'

const activeTab = ref('market')
const marketSkills = ref<any[]>([])
const installedSkills = ref<any[]>([])
const loadingMarket = ref(false)
const loadingInstalled = ref(false)
const installingId = ref<string | null>(null)

const createDialogVisible = ref(false)
const createForm = ref({
  name: '',
  description: '',
  category: '',
  icon: '',
  code: ''
})

const configDialogVisible = ref(false)
const editingSkill = ref<any>(null)
const skillConfig = ref('')

// 判断是否是 URL 图标
function isUrlIcon(icon: string) {
  return icon && (icon.startsWith('http://') || icon.startsWith('https://') || icon.startsWith('/'))
}

// 分类 emoji 映射
const categoryEmojiMap: Record<string, string> = {
  life: '🌤️',
  finance: '📈',
  utility: '🔧',
  data: '📊',
  content: '✍️',
  tool: '🛠️',
  other: '⚡'
}

function getCategoryEmoji(category: string) {
  return categoryEmojiMap[category] || '⚡'
}

// 分类中文标签
const categoryLabelMap: Record<string, string> = {
  life: '生活服务',
  finance: '金融财经',
  utility: '实用工具',
  data: '数据处理',
  content: '内容生成',
  tool: '工具调用',
  other: '其他'
}

function getCategoryLabel(category: string) {
  return categoryLabelMap[category] || category || '其他'
}

function getCategoryTagType(category: string) {
  const map: Record<string, string> = {
    life: 'success',
    finance: 'warning',
    utility: '',
    data: 'info',
    content: 'danger',
    tool: '',
    other: 'info'
  }
  return map[category] || 'info'
}

// 图标背景色
function getIconBg(category: string) {
  const map: Record<string, string> = {
    life: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
    finance: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
    utility: 'linear-gradient(135deg, #4facfe 0%, #00f2fe 100%)',
    data: 'linear-gradient(135deg, #43e97b 0%, #38f9d7 100%)',
    content: 'linear-gradient(135deg, #fa709a 0%, #fee140 100%)',
    tool: 'linear-gradient(135deg, #a18cd1 0%, #fbc2eb 100%)',
    other: 'linear-gradient(135deg, #fccb90 0%, #d57eeb 100%)'
  }
  return map[category] || map.other
}

const showCreateDialog = () => {
  createForm.value = { name: '', description: '', category: '', icon: '', code: '' }
  createDialogVisible.value = true
}

const handleCreate = async () => {
  if (!createForm.value.name || !createForm.value.description) {
    ElMessage.warning('请填写技能名称和描述')
    return
  }
  try {
    await createSkill({
      name: createForm.value.name,
      description: createForm.value.description,
      category: createForm.value.category,
      icon: createForm.value.icon,
      code: createForm.value.code,
    } as any)
    ElMessage.success('创建成功')
    createDialogVisible.value = false
    loadMarketSkills()
  } catch (e: any) {
    ElMessage.error(e.message || '创建失败')
  }
}

const installSkill = async (skill: any) => {
  installingId.value = skill.id
  try {
    await installSkillApi(skill.id)
    ElMessage.success(`已安装技能: ${skill.name}`)
    skill.installed = true
    loadInstalledSkills()
  } catch (e: any) {
    ElMessage.error(e.message || '安装失败')
  } finally {
    installingId.value = null
  }
}

const uninstallSkill = async (id: string) => {
  try {
    await uninstallSkillApi(id)
    ElMessage.success('卸载成功')
    loadInstalledSkills()
    loadMarketSkills()
  } catch (e: any) {
    ElMessage.error(e.message || '卸载失败')
  }
}

const configureSkill = async (skill: any) => {
  editingSkill.value = skill
  try {
    const data = await getSkillConfig(skill.id)
    skillConfig.value = data.config || '{}'
  } catch {
    skillConfig.value = '{}'
  }
  configDialogVisible.value = true
}

const handleConfigSave = async () => {
  if (!editingSkill.value) return
  try {
    await updateSkillConfig(editingSkill.value.id, skillConfig.value)
    ElMessage.success('配置保存成功')
    configDialogVisible.value = false
  } catch (e: any) {
    ElMessage.error(e.message || '保存失败')
  }
}

const toggleSkill = async (skill: any) => {
  try {
    await toggleSkillApi(skill.id)
    ElMessage.success(skill.status === 'active' ? '已禁用' : '已启用')
    loadInstalledSkills()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

const loadMarketSkills = async () => {
  loadingMarket.value = true
  try {
    const data = await listSkills({ type: 'built-in' })
    marketSkills.value = data.items || []
  } catch {
    // fallback: try without type filter
    try {
      const data = await listSkills()
      marketSkills.value = data.items || []
    } catch {
      marketSkills.value = []
    }
  } finally {
    loadingMarket.value = false
  }
}

const loadInstalledSkills = async () => {
  loadingInstalled.value = true
  try {
    const data = await getInstalledSkills()
    installedSkills.value = (data.items || []).map((inst: any) => ({
      id: inst.skill_id,
      name: inst.skill_name || inst.name || '',
      description: inst.skill_description || inst.description || '',
      category: inst.skill_category || inst.category || '',
      icon: inst.skill_icon || inst.icon || '',
      status: inst.status
    }))
  } catch {
    installedSkills.value = []
  } finally {
    loadingInstalled.value = false
  }
}

onMounted(() => {
  loadMarketSkills()
  loadInstalledSkills()
})
</script>

<style scoped lang="scss">
.skill-list {
  padding: 0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24px;

  h2 {
    margin: 0;
    font-size: 20px;
  }
}

.skill-card {
  margin-bottom: 20px;
  border-radius: 12px;
  transition: transform 0.2s;

  &:hover {
    transform: translateY(-2px);
  }

  .skill-header {
    display: flex;
    align-items: center;
    gap: 14px;
    margin-bottom: 14px;

    h3 {
      margin: 0 0 6px 0;
      font-size: 16px;
      font-weight: 600;
    }
  }

  .skill-icon {
    width: 48px;
    height: 48px;
    border-radius: 12px;
    display: flex;
    align-items: center;
    justify-content: center;
    flex-shrink: 0;
  }

  .skill-icon-emoji {
    font-size: 24px;
    line-height: 1;
  }

  .skill-icon-img {
    width: 32px;
    height: 32px;
    object-fit: contain;
  }

  .skill-desc {
    color: #666;
    font-size: 14px;
    margin-bottom: 16px;
    min-height: 40px;
    line-height: 1.5;
  }

  .skill-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .install-count {
      color: #999;
      font-size: 12px;
    }
  }
}

.mini-icon {
  font-size: 16px;
}
</style>
