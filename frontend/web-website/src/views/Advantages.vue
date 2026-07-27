<template>
  <div class="advantages-page">
    <!-- Page Header -->
    <section class="page-header">
      <div class="container">
        <SectionTitle
          tag="技术特点"
          title="为什么选择我们"
          description="企业级架构设计，安全可靠，灵活扩展，让 AI 对话更稳定、更高效"
          :center="true"
        />
      </div>
    </section>

    <!-- Advantages Grid -->
    <section class="section">
      <div class="container">
        <div class="advantages-grid">
          <ScrollReveal v-for="(advantage, index) in advantages" :key="advantage.title" :delay="index * 100">
            <div class="advantage-card">
              <div class="advantage-icon" :style="{ background: advantage.color }">
                {{ advantage.icon }}
              </div>
              <h3>{{ advantage.title }}</h3>
              <p>{{ advantage.description }}</p>
              <ul class="advantage-list">
                <li v-for="item in advantage.features" :key="item">
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                  {{ item }}
                </li>
              </ul>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Technical Details -->
    <section class="section section-alt">
      <div class="container">
        <SectionTitle
          tag="技术细节"
          title="企业级技术架构"
          description="深入了解我们的技术实现，每一个设计都经过精心考量"
          :center="true"
        />
        <div class="tech-details">
          <ScrollReveal>
            <div class="tech-section">
              <h3>🔒 安全与认证</h3>
              <div class="tech-items">
                <div class="tech-item">
                  <strong>JWT 认证</strong>
                  <p>基于 JSON Web Token 的无状态认证，支持 Token 刷新和黑名单机制</p>
                </div>
                <div class="tech-item">
                  <strong>API Key 加密</strong>
                  <p>AES-256 加密存储第三方 API Key，密钥与数据分离，安全可靠</p>
                </div>
                <div class="tech-item">
                  <strong>数据隔离</strong>
                  <p>多租户数据完全隔离，基于租户 ID 的访问控制，防止数据泄露</p>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="tech-section">
              <h3>⚡ 性能优化</h3>
              <div class="tech-items">
                <div class="tech-item">
                  <strong>SSE 流式响应</strong>
                  <p>Server-Sent Events 实现流式输出，首字响应时间 < 100ms，用户体验流畅</p>
                </div>
                <div class="tech-item">
                  <strong>Redis 缓存</strong>
                  <p>会话上下文、用户信息缓存，减少数据库查询，提升响应速度</p>
                </div>
                <div class="tech-item">
                  <strong>连接池管理</strong>
                  <p>数据库连接池、HTTP 连接池优化，支持高并发场景</p>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="tech-section">
              <h3>🔄 高可用设计</h3>
              <div class="tech-items">
                <div class="tech-item">
                  <strong>服务独立部署</strong>
                  <p>7 个微服务独立部署，故障隔离，单服务故障不影响整体系统</p>
                </div>
                <div class="tech-item">
                  <strong>水平扩展</strong>
                  <p>无状态服务设计，支持负载均衡和水平扩展，应对流量高峰</p>
                </div>
                <div class="tech-item">
                  <strong>健康检查</strong>
                  <p>每个服务提供健康检查接口，配合监控系统实现自动故障恢复</p>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="tech-section">
              <h3>📦 灵活扩展</h3>
              <div class="tech-items">
                <div class="tech-item">
                  <strong>多模型支持</strong>
                  <p>支持通义千问、DeepSeek、OpenAI 等 10+ 主流大模型，随时切换</p>
                </div>
                <div class="tech-item">
                  <strong>技能插件化</strong>
                  <p>技能系统插件化设计，支持自定义技能开发，扩展智能体能力</p>
                </div>
                <div class="tech-item">
                  <strong>多渠道接入</strong>
                  <p>提供 REST API、WebSocket、Web Widget 等多种接入方式</p>
                </div>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Comparison -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="对比优势"
          title="与传统方案对比"
          description="看看我们相比传统方案有哪些优势"
          :center="true"
        />
        <ScrollReveal>
          <div class="comparison-table">
            <div class="table-header">
              <div class="table-cell feature-col">功能特性</div>
              <div class="table-cell">AI 智能体平台</div>
              <div class="table-cell">传统方案</div>
            </div>
            <div class="table-row" v-for="item in comparison" :key="item.feature">
              <div class="table-cell feature-col">
                <strong>{{ item.feature }}</strong>
              </div>
              <div class="table-cell">
                <span class="check">{{ item.ours }}</span>
              </div>
              <div class="table-cell">
                <span class="cross">{{ item.traditional }}</span>
              </div>
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta-section">
      <div class="container">
        <ScrollReveal>
          <h2>体验企业级 AI 对话平台</h2>
          <p>安全、稳定、可扩展，让您的业务更智能</p>
          <div class="cta-actions">
            <a href="#" class="btn btn-primary btn-large">免费试用</a>
            <router-link to="/docs/deploy" class="btn btn-secondary btn-large">查看部署文档</router-link>
          </div>
        </ScrollReveal>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import SectionTitle from '@/components/SectionTitle.vue'
import ScrollReveal from '@/components/ScrollReveal.vue'

const advantages = [
  {
    icon: '🏢',
    title: '多租户架构',
    description: '完整的多租户支持，数据完全隔离，独立配置，满足 SaaS 场景需求',
    color: '#3b82f6',
    features: ['租户数据隔离', '独立配置管理', '租户级配额控制', '统一平台管理']
  },
  {
    icon: '⚡',
    title: 'SSE 流式对话',
    description: '基于 Server-Sent Events 的流式响应，实时输出，打字机效果',
    color: '#10b981',
    features: ['实时流式输出', '首字响应 < 100ms', '打字机效果', '低延迟体验']
  },
  {
    icon: '🔒',
    title: '安全可靠',
    description: 'API Key 加密存储，JWT 认证，数据隔离，全方位安全保障',
    color: '#f59e0b',
    features: ['AES-256 加密', 'JWT 认证', '数据隔离', '访问控制']
  },
  {
    icon: '🚀',
    title: '高性能',
    description: 'Go + Python 混合架构，Redis 缓存，连接池优化，支持高并发',
    color: '#8b5cf6',
    features: ['Go 高性能服务', 'Redis 缓存', '连接池优化', '水平扩展']
  },
  {
    icon: '🔌',
    title: '灵活扩展',
    description: '微服务架构，插件化设计，支持多模型、多技能、多渠道',
    color: '#ec4899',
    features: ['微服务架构', '插件化设计', '多模型支持', '多渠道接入']
  },
  {
    icon: '📊',
    title: '完善监控',
    description: '健康检查、使用统计、对话分析，全方位监控和数据分析',
    color: '#06b6d4',
    features: ['健康检查', '使用统计', '对话分析', '趋势报表']
  }
]

const comparison = [
  { feature: '部署方式', ours: 'Docker 一键部署', traditional: '手动配置复杂' },
  { feature: '模型支持', ours: '10+ 主流模型', traditional: '单一模型' },
  { feature: '响应速度', ours: 'SSE 流式 < 100ms', traditional: '同步等待 2-5s' },
  { feature: '多租户', ours: '完整支持', traditional: '不支持' },
  { feature: '知识库', ours: 'RAG 检索增强', traditional: '无' },
  { feature: '技能扩展', ours: '插件化支持', traditional: '固定功能' },
  { feature: '数据分析', ours: '完整统计报表', traditional: '基础日志' },
  { feature: '安全性', ours: 'API Key 加密', traditional: '明文存储' }
]
</script>

<style lang="scss" scoped>
.page-header {
  padding: 140px 0 60px;
  background: linear-gradient(180deg, $gray-50 0%, $bg-white 100%);
}

.advantages-grid {
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

.advantage-card {
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

  .advantage-icon {
    width: 56px;
    height: 56px;
    border-radius: $radius-md;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 28px;
    margin-bottom: $spacing-lg;
  }

  h3 {
    font-size: 1.25rem;
    font-weight: 600;
    color: $text-primary;
    margin-bottom: $spacing-sm;
  }

  p {
    color: $text-secondary;
    font-size: 0.95rem;
    line-height: 1.6;
    margin-bottom: $spacing-md;
  }

  .advantage-list {
    list-style: none;
    padding: 0;
    margin: 0;

    li {
      display: flex;
      align-items: center;
      gap: $spacing-sm;
      padding: 6px 0;
      color: $text-secondary;
      font-size: 0.9rem;

      svg {
        width: 16px;
        height: 16px;
        color: $success;
        flex-shrink: 0;
      }
    }
  }
}

.tech-details {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-2xl;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.tech-section {
  h3 {
    font-size: 1.5rem;
    font-weight: 600;
    color: $text-primary;
    margin-bottom: $spacing-lg;
  }
}

.tech-items {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;
}

.tech-item {
  background: $bg-white;
  border-radius: $radius-md;
  padding: $spacing-lg;
  border: 1px solid $border-light;

  strong {
    display: block;
    font-size: 1rem;
    color: $text-primary;
    margin-bottom: $spacing-xs;
  }

  p {
    color: $text-secondary;
    font-size: 0.9rem;
    line-height: 1.6;
    margin: 0;
  }
}

.comparison-table {
  background: $bg-white;
  border-radius: $radius-lg;
  overflow: hidden;
  border: 1px solid $border-light;
  box-shadow: $shadow-md;
}

.table-header {
  display: grid;
  grid-template-columns: 2fr 1.5fr 1.5fr;
  background: $gray-50;
  border-bottom: 2px solid $border-light;

  .table-cell {
    padding: $spacing-md $spacing-lg;
    font-weight: 600;
    color: $text-primary;
    font-size: 0.95rem;

    &:not(:first-child) {
      text-align: center;
    }
  }
}

.table-row {
  display: grid;
  grid-template-columns: 2fr 1.5fr 1.5fr;
  border-bottom: 1px solid $border-light;

  &:last-child {
    border-bottom: none;
  }

  .table-cell {
    padding: $spacing-md $spacing-lg;

    &:not(:first-child) {
      text-align: center;
      display: flex;
      align-items: center;
      justify-content: center;
    }

    strong {
      color: $text-primary;
      font-size: 0.95rem;
    }

    .check {
      color: $success;
      font-weight: 500;
      font-size: 0.9rem;
    }

    .cross {
      color: $text-muted;
      font-size: 0.9rem;
    }
  }

  &:hover {
    background: $gray-50;
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
