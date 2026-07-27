<template>
  <div class="features-page">
    <!-- Page Header -->
    <section class="page-header">
      <div class="container">
        <SectionTitle
          tag="功能介绍"
          title="强大的 AI 对话功能"
          description="从智能体管理到知识库构建，从多渠道接入到数据分析，一站式满足企业智能化需求"
          :center="true"
        />
      </div>
    </section>

    <!-- Feature Sections -->
    <section class="feature-section" v-for="(feature, index) in features" :key="feature.title">
      <div class="container">
        <div class="feature-layout" :class="{ reverse: index % 2 !== 0 }">
          <ScrollReveal :animation="index % 2 === 0 ? 'fade-left' : 'fade-right'">
            <div class="feature-content">
              <div class="feature-icon" :style="{ background: feature.color }">
                {{ feature.icon }}
              </div>
              <h2>{{ feature.title }}</h2>
              <p class="feature-desc">{{ feature.description }}</p>
              <ul class="feature-list">
                <li v-for="item in feature.highlights" :key="item">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                  {{ item }}
                </li>
              </ul>
              <div class="feature-tags">
                <span v-for="tag in feature.tags" :key="tag" class="tag">{{ tag }}</span>
              </div>
            </div>
          </ScrollReveal>
          <ScrollReveal :animation="index % 2 === 0 ? 'fade-right' : 'fade-left'">
            <div class="feature-visual">
              <div class="visual-card" :style="{ borderColor: feature.color }">
                <div class="visual-header" :style="{ background: feature.color }">
                  <span>{{ feature.icon }}</span>
                  <span>{{ feature.title }}</span>
                </div>
                <div class="visual-body">
                  <div v-for="(item, i) in feature.demoItems" :key="i" class="demo-item">
                    <div class="demo-icon">{{ item.icon }}</div>
                    <div class="demo-info">
                      <strong>{{ item.title }}</strong>
                      <span>{{ item.desc }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta-section">
      <div class="container">
        <ScrollReveal>
          <h2>开始使用全部功能</h2>
          <p>立即体验 AI 智能体平台的所有强大功能</p>
          <div class="cta-actions">
            <a href="#" class="btn btn-primary btn-large">免费试用</a>
            <router-link to="/docs/guide" class="btn btn-secondary btn-large">查看文档</router-link>
          </div>
        </ScrollReveal>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import SectionTitle from '@/components/SectionTitle.vue'
import ScrollReveal from '@/components/ScrollReveal.vue'

const features = [
  {
    icon: '🤖',
    title: '智能体管理',
    description: '创建和管理多个 AI 智能体，每个智能体可以拥有独立的人设、知识库和技能配置，满足不同业务场景需求。',
    color: '#3b82f6',
    highlights: [
      '创建多个独立智能体',
      '自定义人设和对话风格',
      '工作流编排和自动化',
      '智能体状态监控和管理'
    ],
    tags: ['多智能体', '个性化', '工作流'],
    demoItems: [
      { icon: '🎯', title: '客服助手', desc: '7x24 小时在线' },
      { icon: '📚', title: '知识专家', desc: '基于企业知识库' },
      { icon: '💼', title: '业务顾问', desc: '行业专属配置' }
    ]
  },
  {
    icon: '📚',
    title: '知识库管理',
    description: '上传企业文档自动构建知识库，通过 RAG（检索增强生成）技术，让 AI 基于您的私有数据精准回答。',
    color: '#10b981',
    highlights: [
      '支持 PDF、Word、TXT 等多种格式',
      '自动文档分块和向量化',
      '语义检索精准匹配',
      '知识库版本管理和更新'
    ],
    tags: ['RAG', '向量检索', '文档管理'],
    demoItems: [
      { icon: '📄', title: '产品手册', desc: 'PDF 自动解析' },
      { icon: '📊', title: '技术文档', desc: '结构化存储' },
      { icon: '🔍', title: '语义检索', desc: '精准匹配答案' }
    ]
  },
  {
    icon: '⚡',
    title: '技能市场',
    description: '丰富的内置技能让智能体更强大，支持自定义技能扩展，通过技能组合实现复杂业务逻辑。',
    color: '#f59e0b',
    highlights: [
      '内置常用技能开箱即用',
      '支持自定义技能开发',
      '技能热插拔无需重启',
      '技能组合编排'
    ],
    tags: ['内置技能', '自定义', '热插拔'],
    demoItems: [
      { icon: '🌐', title: '网页搜索', desc: '实时获取信息' },
      { icon: '🧮', title: '数学计算', desc: '精确计算能力' },
      { icon: '🎨', title: '图像生成', desc: 'AI 绘图能力' }
    ]
  },
  {
    icon: '💬',
    title: 'SSE 流式对话',
    description: '基于 Server-Sent Events 的流式对话，实时显示 AI 回复，打字机效果带来流畅的用户体验。',
    color: '#8b5cf6',
    highlights: [
      '实时流式响应',
      '打字机效果逐字显示',
      '支持多轮对话上下文',
      '对话历史持久化存储'
    ],
    tags: ['SSE', '流式', '实时'],
    demoItems: [
      { icon: '⚡', title: '低延迟', desc: '毫秒级响应' },
      { icon: '📝', title: '打字机效果', desc: '逐字流畅显示' },
      { icon: '💾', title: '历史记录', desc: '自动保存对话' }
    ]
  },
  {
    icon: '🔗',
    title: '多渠道接入',
    description: '支持网页嵌入、微信公众号、小程序、API 等多种接入方式，一处配置多处使用。',
    color: '#ec4899',
    highlights: [
      '网页 Widget 一键嵌入',
      '微信公众号接入',
      'RESTful API 支持',
      'Webhook 事件通知'
    ],
    tags: ['网页', '微信', 'API'],
    demoItems: [
      { icon: '🌐', title: '网页嵌入', desc: 'JS Widget 集成' },
      { icon: '💚', title: '微信接入', desc: '公众号对接' },
      { icon: '🔌', title: 'API 接口', desc: 'RESTful 调用' }
    ]
  },
  {
    icon: '📊',
    title: '数据分析',
    description: '全面的对话分析和使用统计，帮助您洞察业务趋势，优化智能体表现。',
    color: '#06b6d4',
    highlights: [
      '对话量和活跃度统计',
      '用户满意度分析',
      '知识库命中率监控',
      '使用趋势报表导出'
    ],
    tags: ['统计', '分析', '报表'],
    demoItems: [
      { icon: '📈', title: '趋势图表', desc: '可视化数据' },
      { icon: '🎯', title: 'KPI 监控', desc: '关键指标追踪' },
      { icon: '📋', title: '报表导出', desc: 'Excel/PDF 格式' }
    ]
  }
]
</script>

<style lang="scss" scoped>
.page-header {
  padding: 140px 0 60px;
  background: linear-gradient(180deg, $gray-50 0%, $bg-white 100%);
}

.feature-section {
  padding: $spacing-3xl 0;

  &:nth-child(even) {
    background: $gray-50;
  }
}

.feature-layout {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-3xl;
  align-items: center;

  &.reverse {
    direction: rtl;

    > * {
      direction: ltr;
    }
  }

  @include mobile {
    grid-template-columns: 1fr;
    gap: $spacing-2xl;

    &.reverse {
      direction: ltr;
    }
  }
}

.feature-content {
  .feature-icon {
    width: 64px;
    height: 64px;
    border-radius: $radius-lg;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 32px;
    margin-bottom: $spacing-lg;
  }

  h2 {
    font-size: 2rem;
    font-weight: 700;
    color: $text-primary;
    margin-bottom: $spacing-md;
  }

  .feature-desc {
    font-size: 1.05rem;
    color: $text-secondary;
    line-height: 1.8;
    margin-bottom: $spacing-lg;
  }

  .feature-list {
    list-style: none;
    padding: 0;
    margin: 0 0 $spacing-lg;

    li {
      display: flex;
      align-items: center;
      gap: $spacing-sm;
      padding: $spacing-sm 0;
      color: $text-secondary;
      font-size: 0.95rem;

      svg {
        width: 18px;
        height: 18px;
        color: $success;
        flex-shrink: 0;
      }
    }
  }

  .feature-tags {
    display: flex;
    gap: $spacing-sm;
    flex-wrap: wrap;

    .tag {
      padding: 4px 12px;
      background: rgba($primary, 0.08);
      color: $primary;
      font-size: 13px;
      font-weight: 500;
      border-radius: $radius-full;
    }
  }
}

.feature-visual {
  .visual-card {
    background: $bg-white;
    border-radius: $radius-xl;
    border: 2px solid $border-light;
    overflow: hidden;
    box-shadow: $shadow-lg;
  }

  .visual-header {
    padding: $spacing-md $spacing-lg;
    color: white;
    display: flex;
    align-items: center;
    gap: $spacing-sm;
    font-weight: 600;
    font-size: 15px;
  }

  .visual-body {
    padding: $spacing-lg;
  }

  .demo-item {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    padding: $spacing-md;
    border-radius: $radius-md;
    margin-bottom: $spacing-sm;
    background: $gray-50;
    transition: all $transition-fast;

    &:hover {
      background: $gray-100;
      transform: translateX(4px);
    }

    .demo-icon {
      font-size: 24px;
      width: 44px;
      height: 44px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: $bg-white;
      border-radius: $radius-md;
      flex-shrink: 0;
    }

    .demo-info {
      strong {
        display: block;
        font-size: 0.95rem;
        color: $text-primary;
      }

      span {
        font-size: 0.85rem;
        color: $text-muted;
      }
    }
  }
}

.cta-section {
  padding: $spacing-4xl 0;
  background: linear-gradient(135deg, $dark-900 0%, $dark-800 100%);
  text-align: center;

  h2 {
    font-size: 2.5rem;
    color: $text-white;
    margin-bottom: $spacing-md;

    @include mobile {
      font-size: 2rem;
    }
  }

  p {
    font-size: 1.125rem;
    color: $gray-400;
    margin-bottom: $spacing-xl;
  }

  .cta-actions {
    display: flex;
    justify-content: center;
    gap: $spacing-md;
    flex-wrap: wrap;
  }
}
</style>
