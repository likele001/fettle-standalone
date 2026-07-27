<template>
  <div class="page-container">
    <div class="page-header">
      <h2>套餐管理</h2>
      <el-button type="primary" @click="showCreateDialog">
        <el-icon><Plus /></el-icon>
        创建套餐
      </el-button>
    </div>

    <!-- 套餐列表 -->
    <el-row :gutter="20">
      <el-col :span="8" v-for="plan in plans" :key="plan.id">
        <el-card class="plan-card" shadow="hover">
          <div class="plan-header">
            <h3>{{ plan.name }}</h3>
            <el-tag :type="planTypeTag(plan.type)" size="small">{{ planTypeLabel(plan.type) }}</el-tag>
          </div>

          <div class="plan-price">
            <div class="price-item">
              <span class="price-value">¥{{ plan.price_monthly }}</span>
              <span class="price-label">/月</span>
            </div>
            <div class="price-item">
              <span class="price-value">¥{{ plan.price_yearly }}</span>
              <span class="price-label">/年</span>
            </div>
          </div>

          <div class="plan-quota">
            <div class="quota-item">
              <span>智能体数量</span>
              <span>{{ plan.max_agents }}</span>
            </div>
            <div class="quota-item">
              <span>知识库数量</span>
              <span>{{ plan.max_knowledge_bases }}</span>
            </div>
            <div class="quota-item">
              <span>每知识库文档数</span>
              <span>{{ plan.max_documents_per_kb }}</span>
            </div>
            <div class="quota-item">
              <span>月消息数</span>
              <span>{{ formatNumber(plan.max_messages_per_month) }}</span>
            </div>
            <div class="quota-item">
              <span>最大并发</span>
              <span>{{ plan.max_concurrent_sessions }}</span>
            </div>
          </div>

          <div class="plan-features" v-if="plan.features?.length">
            <div class="feature-tag" v-for="(feature, idx) in plan.features.slice(0, 3)" :key="idx">
              {{ feature }}
            </div>
            <div v-if="plan.features.length > 3" class="feature-more">
              +{{ plan.features.length - 3 }} 更多
            </div>
          </div>

          <div class="plan-restrictions" v-if="plan.allowed_industries?.length || plan.allowed_features?.length">
            <div class="restriction-section" v-if="plan.allowed_industries?.length">
              <div class="restriction-label">允许行业</div>
              <div class="restriction-tags">
                <div class="restriction-tag" v-for="industry in plan.allowed_industries.slice(0, 3)" :key="industry">
                  {{ industryLabels[industry] || industry }}
                </div>
                <div v-if="plan.allowed_industries.length > 3" class="feature-more">
                  +{{ plan.allowed_industries.length - 3 }}
                </div>
              </div>
            </div>
            <div class="restriction-section" v-if="plan.allowed_features?.length">
              <div class="restriction-label">允许功能</div>
              <div class="restriction-tags">
                <div class="restriction-tag" v-for="feature in plan.allowed_features.slice(0, 3)" :key="feature">
                  {{ featureLabels[feature] || feature }}
                </div>
                <div v-if="plan.allowed_features.length > 3" class="feature-more">
                  +{{ plan.allowed_features.length - 3 }}
                </div>
              </div>
            </div>
          </div>

          <div class="plan-status">
            <el-tag :type="plan.is_active ? 'success' : 'info'" size="small">
              {{ plan.is_active ? '已上架' : '已下架' }}
            </el-tag>
          </div>

          <div class="plan-actions">
            <el-button size="small" @click="editPlan(plan)">编辑</el-button>
            <el-button
              size="small"
              :type="plan.is_active ? 'warning' : 'success'"
              @click="toggleStatus(plan)"
            >
              {{ plan.is_active ? '下架' : '上架' }}
            </el-button>
            <el-popconfirm title="确定删除此套餐？" @confirm="handleDelete(plan.id)">
              <template #reference>
                <el-button size="small" type="danger">删除</el-button>
              </template>
            </el-popconfirm>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="plans.length === 0" description="暂无套餐，点击创建开始" />

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="isEdit ? '编辑套餐' : '创建套餐'"
      width="600px"
    >
      <el-form :model="formData" :rules="rules" ref="formRef" label-width="120px">
        <el-form-item label="套餐名称" prop="name">
          <el-input v-model="formData.name" placeholder="请输入套餐名称" />
        </el-form-item>

        <el-form-item label="套餐类型" prop="type">
          <el-select v-model="formData.type" placeholder="选择类型">
            <el-option label="免费版" value="free" />
            <el-option label="专业版" value="pro" />
            <el-option label="企业版" value="enterprise" />
          </el-select>
        </el-form-item>

        <el-form-item label="月付价格" prop="price_monthly">
          <el-input-number v-model="formData.price_monthly" :min="0" :step="10" />
          <span style="margin-left: 8px; color: #94a3b8;">元/月</span>
        </el-form-item>

        <el-form-item label="年付价格" prop="price_yearly">
          <el-input-number v-model="formData.price_yearly" :min="0" :step="100" />
          <span style="margin-left: 8px; color: #94a3b8;">元/年</span>
        </el-form-item>

        <el-divider>配额设置</el-divider>

        <el-form-item label="智能体数量" prop="max_agents">
          <el-input-number v-model="formData.max_agents" :min="1" />
        </el-form-item>

        <el-form-item label="知识库数量" prop="max_knowledge_bases">
          <el-input-number v-model="formData.max_knowledge_bases" :min="1" />
        </el-form-item>

        <el-form-item label="每知识库文档数" prop="max_documents_per_kb">
          <el-input-number v-model="formData.max_documents_per_kb" :min="1" />
        </el-form-item>

        <el-form-item label="月消息数" prop="max_messages_per_month">
          <el-input-number v-model="formData.max_messages_per_month" :min="100" :step="1000" />
        </el-form-item>

        <el-form-item label="最大并发" prop="max_concurrent_sessions">
          <el-input-number v-model="formData.max_concurrent_sessions" :min="1" />
        </el-form-item>

        <el-form-item label="特色功能" prop="features">
          <el-select
            v-model="formData.features"
            multiple
            filterable
            allow-create
            default-first-option
            placeholder="输入功能名称后回车添加"
            style="width: 100%"
          />
        </el-form-item>

        <el-divider>权限限制</el-divider>

        <el-form-item label="允许的行业" prop="allowed_industries">
          <el-select
            v-model="formData.allowed_industries"
            multiple
            filterable
            placeholder="选择允许使用的行业（空表示不限制）"
            style="width: 100%"
          >
            <el-option label="电商" value="ecommerce" />
            <el-option label="餐饮" value="dining" />
            <el-option label="法律" value="legal" />
            <el-option label="教育" value="education" />
            <el-option label="医疗" value="medical" />
            <el-option label="金融" value="finance" />
            <el-option label="物流" value="logistics" />
            <el-option label="房产" value="real_estate" />
            <el-option label="零售" value="retail" />
            <el-option label="娱乐" value="entertainment" />
          </el-select>
          <span style="font-size: 12px; color: #94a3b8;">选择后，该套餐租户只能使用选中行业的智能体</span>
        </el-form-item>

        <el-form-item label="允许的功能" prop="allowed_features">
          <el-select
            v-model="formData.allowed_features"
            multiple
            filterable
            placeholder="选择允许使用的功能模块（空表示不限制）"
            style="width: 100%"
          >
            <el-option label="智能体管理" value="agent_management" />
            <el-option label="知识库管理" value="knowledge_base" />
            <el-option label="技能市场" value="skill_market" />
            <el-option label="对话管理" value="conversation" />
            <el-option label="统计分析" value="analytics" />
            <el-option label="渠道配置" value="channel_config" />
            <el-option label="AI配置" value="ai_config" />
            <el-option label="自定义工作流" value="workflow" />
            <el-option label="API访问" value="api_access" />
            <el-option label="团队协作" value="team_collaboration" />
          </el-select>
          <span style="font-size: 12px; color: #94a3b8;">选择后，该套餐租户只能访问选中的功能模块</span>
        </el-form-item>

        <el-form-item label="排序" prop="sort_order">
          <el-input-number v-model="formData.sort_order" :min="0" />
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          {{ isEdit ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getPlans, createPlan, updatePlan, deletePlan, togglePlanStatus, type Plan } from '@/api/plan'
import type { FormInstance, FormRules } from 'element-plus'

const plans = ref<Plan[]>([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const editingId = ref('')

const formData = ref({
  name: '',
  type: 'free' as 'free' | 'pro' | 'enterprise',
  price_monthly: 0,
  price_yearly: 0,
  max_agents: 1,
  max_knowledge_bases: 1,
  max_documents_per_kb: 10,
  max_messages_per_month: 1000,
  max_concurrent_sessions: 5,
  features: [] as string[],
  allowed_industries: [] as string[],
  allowed_features: [] as string[],
  sort_order: 0
})

const industryLabels: Record<string, string> = {
  ecommerce: '电商',
  dining: '餐饮',
  legal: '法律',
  education: '教育',
  medical: '医疗',
  finance: '金融',
  logistics: '物流',
  real_estate: '房产',
  retail: '零售',
  entertainment: '娱乐'
}

const featureLabels: Record<string, string> = {
  agent_management: '智能体管理',
  knowledge_base: '知识库管理',
  skill_market: '技能市场',
  conversation: '对话管理',
  analytics: '统计分析',
  channel_config: '渠道配置',
  ai_config: 'AI配置',
  workflow: '自定义工作流',
  api_access: 'API访问',
  team_collaboration: '团队协作'
}

const rules: FormRules = {
  name: [{ required: true, message: '请输入套餐名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择套餐类型', trigger: 'change' }]
}

function planTypeTag(type: string) {
  const map: Record<string, string> = { free: 'info', pro: 'primary', enterprise: 'warning' }
  return map[type] || 'info'
}

function planTypeLabel(type: string) {
  const map: Record<string, string> = { free: '免费版', pro: '专业版', enterprise: '企业版' }
  return map[type] || type
}

function formatNumber(num: number) {
  if (num >= 10000) return (num / 10000).toFixed(1) + '万'
  return num.toString()
}

async function loadPlans() {
  try {
    const res = await getPlans()
    plans.value = res.items
  } catch {
    plans.value = []
  }
}

function showCreateDialog() {
  isEdit.value = false
  editingId.value = ''
  formData.value = {
    name: '',
    type: 'free',
    price_monthly: 0,
    price_yearly: 0,
    max_agents: 1,
    max_knowledge_bases: 1,
    max_documents_per_kb: 10,
    max_messages_per_month: 1000,
    max_concurrent_sessions: 5,
    features: [],
    allowed_industries: [],
    allowed_features: [],
    sort_order: 0
  }
  dialogVisible.value = true
}

function editPlan(plan: Plan) {
  isEdit.value = true
  editingId.value = plan.id
  formData.value = {
    name: plan.name,
    type: plan.type,
    price_monthly: plan.price_monthly,
    price_yearly: plan.price_yearly,
    max_agents: plan.max_agents,
    max_knowledge_bases: plan.max_knowledge_bases,
    max_documents_per_kb: plan.max_documents_per_kb,
    max_messages_per_month: plan.max_messages_per_month,
    max_concurrent_sessions: plan.max_concurrent_sessions,
    features: plan.features || [],
    allowed_industries: plan.allowed_industries || [],
    allowed_features: plan.allowed_features || [],
    sort_order: plan.sort_order
  }
  dialogVisible.value = true
}

async function handleSubmit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value) {
        await updatePlan(editingId.value, formData.value)
        ElMessage.success('套餐已更新')
      } else {
        await createPlan(formData.value)
        ElMessage.success('套餐已创建')
      }
      dialogVisible.value = false
      await loadPlans()
    } catch (e: any) {
      ElMessage.error(e.message || '操作失败')
    } finally {
      submitting.value = false
    }
  })
}

async function toggleStatus(plan: Plan) {
  const newStatus = plan.is_active ? 'inactive' : 'active'
  const label = newStatus === 'active' ? '上架' : '下架'
  try {
    await togglePlanStatus(plan.id, newStatus)
    ElMessage.success(`套餐已${label}`)
    await loadPlans()
  } catch (e: any) {
    ElMessage.error(e.message || '操作失败')
  }
}

async function handleDelete(id: string) {
  try {
    await deletePlan(id)
    ElMessage.success('套餐已删除')
    await loadPlans()
  } catch (e: any) {
    ElMessage.error(e.message || '删除失败')
  }
}

onMounted(() => loadPlans())
</script>

<style scoped lang="scss">
.plan-card {
  margin-bottom: 20px;

  .plan-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;

    h3 {
      font-size: 18px;
      font-weight: 600;
      margin: 0;
    }
  }

  .plan-price {
    display: flex;
    gap: 24px;
    margin-bottom: 16px;
    padding: 12px;
    background: #f8fafc;
    border-radius: 8px;

    .price-item {
      .price-value {
        font-size: 24px;
        font-weight: 700;
        color: #3b82f6;
      }
      .price-label {
        font-size: 12px;
        color: #94a3b8;
        margin-left: 4px;
      }
    }
  }

  .plan-quota {
    margin-bottom: 16px;

    .quota-item {
      display: flex;
      justify-content: space-between;
      padding: 8px 0;
      border-bottom: 1px solid #f1f5f9;
      font-size: 14px;

      &:last-child {
        border-bottom: none;
      }

      span:first-child {
        color: #64748b;
      }

      span:last-child {
        font-weight: 500;
      }
    }
  }

  .plan-features {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 16px;

    .feature-tag {
      background: #e0f2fe;
      color: #0369a1;
      padding: 4px 8px;
      border-radius: 4px;
      font-size: 12px;
    }

    .feature-more {
      color: #94a3b8;
      font-size: 12px;
      padding: 4px 8px;
    }
  }

  .plan-restrictions {
    margin-bottom: 16px;
    padding: 12px;
    background: #faf5ff;
    border-radius: 8px;
    border: 1px solid #f3e8ff;

    .restriction-section {
      margin-bottom: 10px;

      &:last-child {
        margin-bottom: 0;
      }
    }

    .restriction-label {
      font-size: 12px;
      color: #7c3aed;
      font-weight: 500;
      margin-bottom: 6px;
    }

    .restriction-tags {
      display: flex;
      flex-wrap: wrap;
      gap: 6px;
    }

    .restriction-tag {
      background: #ede9fe;
      color: #6d28d9;
      padding: 3px 8px;
      border-radius: 4px;
      font-size: 11px;
    }
  }

  .plan-status {
    margin-bottom: 16px;
  }

  .plan-actions {
    display: flex;
    gap: 8px;
  }
}
</style>
