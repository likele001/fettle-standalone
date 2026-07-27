<template>
  <div class="home">
    <!-- Hero Section -->
    <section class="hero">
      <div class="hero-bg">
        <div class="hero-grid"></div>
        <div class="hero-glow"></div>
      </div>
      <div class="container">
        <div class="hero-content">
          <div class="hero-badge">
            <span class="badge-dot"></span>
            企业级 AI 智能对话平台
          </div>
          <h1>
            用 <span class="gradient-text">AI 智能体</span><br>
            重新定义企业对话
          </h1>
          <p class="hero-desc">
            多租户、多模型、多渠道的一站式智能对话解决方案。<br>
            支持通义千问、DeepSeek 等主流大模型，开箱即用。
          </p>
          <div class="hero-actions">
            <a href="#" class="btn btn-primary btn-large">
              免费试用
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
                <line x1="5" y1="12" x2="19" y2="12"></line>
                <polyline points="12 5 19 12 12 19"></polyline>
              </svg>
            </a>
            <router-link to="/docs/deploy" class="btn btn-secondary btn-large">
              查看文档
            </router-link>
          </div>
        </div>
        <div class="hero-visual">
          <div class="hero-card">
            <div class="card-header">
              <div class="card-dots">
                <span></span><span></span><span></span>
              </div>
              <span class="card-title">AI 智能体对话</span>
            </div>
            <div class="card-body">
              <div class="chat-msg user">你好，请帮我查询今天的会议安排</div>
              <div class="chat-msg ai">
                好的，我来为您查询。今天共有 3 场会议：<br>
                1. 10:00 产品评审会<br>
                2. 14:00 技术方案讨论<br>
                3. 16:30 周例会
              </div>
              <div class="chat-input">
                <span>输入消息...</span>
                <div class="send-btn">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                    <line x1="22" y1="2" x2="11" y2="13"></line>
                    <polygon points="22 2 15 22 11 13 2 9 22 2"></polygon>
                  </svg>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Stats Section -->
    <section class="stats-section">
      <div class="container">
        <div class="stats-grid">
          <div class="stat-item" v-for="stat in stats" :key="stat.label">
            <div class="stat-value">
              <CounterUp :value="stat.value" :suffix="stat.suffix" />
            </div>
            <div class="stat-label">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- Features Section -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="核心功能"
          title="一站式 AI 对话解决方案"
          description="从智能体创建到知识库管理，从多渠道接入到数据分析，全方位赋能企业智能化"
          :center="true"
        />
        <div class="features-grid">
          <ScrollReveal v-for="(feature, index) in features" :key="feature.title" :delay="index * 100">
            <FeatureCard
              :icon="feature.icon"
              :title="feature.title"
              :description="feature.description"
              :features="feature.features"
              :color="feature.color"
            />
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Architecture Section -->
    <section class="section section-alt">
      <div class="container">
        <SectionTitle
          tag="技术架构"
          title="微服务架构，灵活可扩展"
          description="6 个独立微服务，每个服务可独立部署和扩展，满足企业级高可用需求"
          :center="true"
        />
        <div class="arch-grid">
          <ScrollReveal animation="fade-left">
            <div class="arch-services">
              <div class="arch-service" v-for="service in services" :key="service.name">
                <div class="service-icon" :style="{ background: service.color }">
                  {{ service.icon }}
                </div>
                <div class="service-info">
                  <h4>{{ service.name }}</h4>
                  <span class="service-port">:{{ service.port }}</span>
                  <p>{{ service.desc }}</p>
                </div>
              </div>
            </div>
          </ScrollReveal>
          <ScrollReveal animation="fade-right">
            <div class="arch-stack">
              <h3>技术栈</h3>
              <div class="stack-list">
                <div class="stack-item" v-for="tech in techStack" :key="tech.name">
                  <div class="stack-icon">{{ tech.icon }}</div>
                  <div>
                    <strong>{{ tech.name }}</strong>
                    <span>{{ tech.desc }}</span>
                  </div>
                </div>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Scenarios Section -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="应用场景"
          title="适用于各行各业"
          description="无论是客户服务、内部协作还是知识管理，AI 智能体平台都能为您提供最佳方案"
          :center="true"
        />
        <div class="scenarios-grid">
          <ScrollReveal v-for="(scenario, index) in scenarios" :key="scenario.title" :delay="index * 150">
            <div class="scenario-card">
              <div class="scenario-icon">{{ scenario.icon }}</div>
              <h3>{{ scenario.title }}</h3>
              <p>{{ scenario.description }}</p>
              <ul>
                <li v-for="item in scenario.items" :key="item">{{ item }}</li>
              </ul>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- CTA Section -->
    <section class="cta-section">
      <div class="container">
        <div class="cta-content">
          <ScrollReveal>
            <h2>准备好开启 AI 智能对话了吗？</h2>
            <p>立即体验企业级 AI 智能体平台，让您的业务更智能、更高效</p>
            <div class="cta-actions">
              <a href="#" class="btn btn-primary btn-large">免费开始</a>
              <router-link to="/docs/guide" class="btn btn-secondary btn-large">查看使用指南</router-link>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import SectionTitle from '@/components/SectionTitle.vue'
import FeatureCard from '@/components/FeatureCard.vue'
import ScrollReveal from '@/components/ScrollReveal.vue'
import CounterUp from '@/components/CounterUp.vue'

const stats = [
  { value: 6, label: '微服务', suffix: '' },
  { value: 10, label: '支持模型', suffix: '+' },
  { value: 99, label: '可用性 %', suffix: '.9' },
  { value: 50, label: '响应速度 ms', suffix: '' }
]

const features = [
  {
    icon: '🤖',
    title: '智能体管理',
    description: '创建、配置和管理多个 AI 智能体，支持个性化人设和工作流',
    features: ['多智能体并行', '个性化人设', '工作流编排'],
    color: '#3b82f6'
  },
  {
    icon: '📚',
    title: '知识库',
    description: '上传文档构建企业知识库，RAG 检索增强生成，精准回答',
    features: ['文档自动分块', '向量检索', 'RAG 增强'],
    color: '#10b981'
  },
  {
    icon: '⚡',
    title: '技能市场',
    description: '内置丰富技能，支持自定义技能扩展，让智能体更强大',
    features: ['内置技能', '自定义扩展', '热插拔'],
    color: '#f59e0b'
  },
  {
    icon: '💬',
    title: 'SSE 流式对话',
    description: '实时流式响应，打字机效果，用户体验更佳',
    features: ['实时流式', '打字机效果', '低延迟'],
    color: '#8b5cf6'
  },
  {
    icon: '🔗',
    title: '多渠道接入',
    description: '支持网页、微信、API 等多种接入方式，一处配置多处使用',
    features: ['网页嵌入', '微信接入', 'REST API'],
    color: '#ec4899'
  },
  {
    icon: '📊',
    title: '数据分析',
    description: '全面的对话分析和使用统计，洞察业务趋势',
    features: ['会话分析', '用量统计', '趋势报表'],
    color: '#06b6d4'
  }
]

const services = [
  { name: 'Gateway', port: 20001, icon: '🌐', color: '#3b82f6', desc: '统一入口，路由分发' },
  { name: 'User Service', port: 20002, icon: '👤', color: '#10b981', desc: '用户认证，租户管理' },
  { name: 'Agent Service', port: 20003, icon: '🤖', color: '#8b5cf6', desc: '智能体，知识库管理' },
  { name: 'Chat Service', port: 20004, icon: '💬', color: '#f59e0b', desc: '对话管理，消息处理' },
  { name: 'Skill Service', port: 20005, icon: '⚡', color: '#ec4899', desc: '技能管理，安装卸载' },
  { name: 'Billing Service', port: 20006, icon: '💰', color: '#06b6d4', desc: '计费管理，订阅套餐' },
  { name: 'AI Engine', port: 20007, icon: '🧠', color: '#ef4444', desc: 'AI 模型调用，流式生成' }
]

const techStack = [
  { name: 'Go 1.22+', icon: '🔵', desc: '高性能后端服务' },
  { name: 'Python 3.11+', icon: '🐍', desc: 'AI 引擎与模型调用' },
  { name: 'Vue 3 + Vite', icon: '💚', desc: '现代化前端框架' },
  { name: 'PostgreSQL 16+', icon: '🐘', desc: '关系型数据库' },
  { name: 'Redis 7+', icon: '🔴', desc: '缓存与消息队列' },
  { name: 'NATS', icon: '📨', desc: '轻量级消息系统' }
]

const scenarios = [
  {
    icon: '🏢',
    title: '企业客服',
    description: '7x24 小时智能客服，自动回答常见问题，提升客户满意度',
    items: ['自动问答', '工单流转', '多轮对话']
  },
  {
    icon: '🎓',
    title: '教育培训',
    description: '智能辅导系统，根据学生水平提供个性化学习方案',
    items: ['知识问答', '学习辅导', '作业批改']
  },
  {
    icon: '🏥',
    title: '医疗健康',
    description: '医疗知识助手，辅助医生查询资料，帮助患者了解病情',
    items: ['症状分析', '知识检索', '健康咨询']
  }
]
</script>

<style lang="scss" scoped>
// Hero
.hero {
  position: relative;
  padding: 140px 0 80px;
  overflow: hidden;
  min-height: 90vh;
  display: flex;
  align-items: center;
}

.hero-bg {
  position: absolute;
  inset: 0;
  z-index: 0;

  .hero-grid {
    position: absolute;
    inset: 0;
    background-image:
      linear-gradient(rgba($primary, 0.03) 1px, transparent 1px),
      linear-gradient(90deg, rgba($primary, 0.03) 1px, transparent 1px);
    background-size: 60px 60px;
  }

  .hero-glow {
    position: absolute;
    top: -200px;
    right: -200px;
    width: 600px;
    height: 600px;
    background: radial-gradient(circle, rgba($primary, 0.08) 0%, transparent 70%);
    border-radius: 50%;
  }
}

.hero .container {
  position: relative;
  z-index: 1;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-3xl;
  align-items: center;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.hero-content {
  .hero-badge {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 8px 16px;
    background: rgba($primary, 0.08);
    border: 1px solid rgba($primary, 0.15);
    border-radius: $radius-full;
    font-size: 14px;
    color: $primary;
    font-weight: 500;
    margin-bottom: $spacing-lg;

    .badge-dot {
      width: 8px;
      height: 8px;
      background: $primary;
      border-radius: 50%;
      animation: pulse 2s infinite;
    }
  }

  h1 {
    font-size: 3.5rem;
    font-weight: 800;
    line-height: 1.15;
    margin-bottom: $spacing-lg;
    color: $text-primary;

    @include mobile {
      font-size: 2.5rem;
    }
  }

  .gradient-text {
    @include gradient-text($primary, #8b5cf6);
  }

  .hero-desc {
    font-size: 1.125rem;
    color: $text-secondary;
    line-height: 1.8;
    margin-bottom: $spacing-xl;
  }

  .hero-actions {
    display: flex;
    gap: $spacing-md;
    flex-wrap: wrap;
  }
}

.hero-visual {
  @include mobile {
    display: none;
  }
}

.hero-card {
  background: $bg-white;
  border-radius: $radius-xl;
  box-shadow: $shadow-xl;
  overflow: hidden;
  border: 1px solid $border-light;

  .card-header {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    padding: $spacing-md $spacing-lg;
    background: $gray-50;
    border-bottom: 1px solid $border-light;

    .card-dots {
      display: flex;
      gap: 6px;

      span {
        width: 10px;
        height: 10px;
        border-radius: 50%;

        &:nth-child(1) { background: #ef4444; }
        &:nth-child(2) { background: #f59e0b; }
        &:nth-child(3) { background: #10b981; }
      }
    }

    .card-title {
      font-size: 13px;
      color: $text-muted;
    }
  }

  .card-body {
    padding: $spacing-lg;
  }

  .chat-msg {
    padding: $spacing-md;
    border-radius: $radius-md;
    margin-bottom: $spacing-md;
    font-size: 14px;
    line-height: 1.6;
    max-width: 85%;

    &.user {
      background: $primary;
      color: white;
      margin-left: auto;
      border-bottom-right-radius: 4px;
    }

    &.ai {
      background: $gray-100;
      color: $text-primary;
      border-bottom-left-radius: 4px;
    }
  }

  .chat-input {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: $spacing-md;
    background: $gray-50;
    border-radius: $radius-md;
    border: 1px solid $border-light;

    span {
      color: $text-muted;
      font-size: 14px;
    }

    .send-btn {
      width: 32px;
      height: 32px;
      background: $primary;
      border-radius: $radius-sm;
      display: flex;
      align-items: center;
      justify-content: center;
      color: white;
    }
  }
}

// Stats
.stats-section {
  padding: $spacing-3xl 0;
  background: $bg-white;
  border-bottom: 1px solid $border-light;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: $spacing-xl;
  text-align: center;

  @include mobile {
    grid-template-columns: repeat(2, 1fr);
    gap: $spacing-lg;
  }
}

.stat-item {
  .stat-value {
    font-size: 2.5rem;
    font-weight: 800;
    color: $primary;
    font-family: $font-display;

    @include mobile {
      font-size: 2rem;
    }
  }

  .stat-label {
    font-size: 0.95rem;
    color: $text-secondary;
    margin-top: 4px;
  }
}

// Features
.features-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: $spacing-lg;

  @include tablet {
    grid-template-columns: repeat(2, 1fr);
  }

  @include mobile {
    grid-template-columns: 1fr;
  }
}

// Architecture
.arch-grid {
  display: grid;
  grid-template-columns: 1.2fr 0.8fr;
  gap: $spacing-2xl;
  align-items: start;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.arch-services {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.arch-service {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-md;
  background: $bg-white;
  border-radius: $radius-md;
  border: 1px solid $border-light;
  transition: all $transition-fast;

  &:hover {
    box-shadow: $shadow-md;
    border-color: transparent;
  }

  .service-icon {
    width: 44px;
    height: 44px;
    border-radius: $radius-md;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 20px;
    flex-shrink: 0;
  }

  .service-info {
    h4 {
      font-size: 0.95rem;
      font-weight: 600;
      color: $text-primary;
      display: inline;
    }

    .service-port {
      font-size: 0.8rem;
      color: $text-muted;
      margin-left: 6px;
    }

    p {
      font-size: 0.85rem;
      color: $text-secondary;
      margin: 2px 0 0;
    }
  }
}

.arch-stack {
  background: $bg-white;
  border-radius: $radius-lg;
  padding: $spacing-xl;
  border: 1px solid $border-light;

  h3 {
    font-size: 1.25rem;
    margin-bottom: $spacing-lg;
    color: $text-primary;
  }
}

.stack-list {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.stack-item {
  display: flex;
  align-items: center;
  gap: $spacing-md;

  .stack-icon {
    font-size: 24px;
    width: 40px;
    text-align: center;
  }

  strong {
    display: block;
    font-size: 0.9rem;
    color: $text-primary;
  }

  span {
    font-size: 0.8rem;
    color: $text-muted;
  }
}

// Scenarios
.scenarios-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: $spacing-lg;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.scenario-card {
  background: $bg-white;
  border-radius: $radius-lg;
  padding: $spacing-xl;
  border: 1px solid $border-light;
  transition: all $transition-normal;

  &:hover {
    transform: translateY(-4px);
    box-shadow: $shadow-lg;
    border-color: transparent;
  }

  .scenario-icon {
    font-size: 40px;
    margin-bottom: $spacing-md;
  }

  h3 {
    font-size: 1.25rem;
    color: $text-primary;
    margin-bottom: $spacing-sm;
  }

  p {
    color: $text-secondary;
    font-size: 0.95rem;
    margin-bottom: $spacing-md;
  }

  ul {
    list-style: none;
    padding: 0;
    margin: 0;

    li {
      padding: 6px 0;
      font-size: 0.9rem;
      color: $text-secondary;
      display: flex;
      align-items: center;
      gap: 8px;

      &::before {
        content: '';
        width: 6px;
        height: 6px;
        background: $primary;
        border-radius: 50%;
        flex-shrink: 0;
      }
    }
  }
}

// CTA
.cta-section {
  padding: $spacing-4xl 0;
  background: linear-gradient(135deg, $dark-900 0%, $dark-800 100%);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: -50%;
    right: -20%;
    width: 500px;
    height: 500px;
    background: radial-gradient(circle, rgba($primary, 0.15) 0%, transparent 70%);
    border-radius: 50%;
  }
}

.cta-content {
  text-align: center;
  position: relative;
  z-index: 1;

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
