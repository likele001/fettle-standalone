<template>
  <div class="pricing-page">
    <!-- Page Header -->
    <section class="page-header">
      <div class="container">
        <SectionTitle
          tag="定价方案"
          title="选择适合您的套餐"
          description="灵活的定价方案，从个人开发者到大型企业，总有一款适合您"
          :center="true"
        />
      </div>
    </section>

    <!-- Pricing Cards -->
    <section class="section">
      <div class="container">
        <div class="pricing-grid">
          <ScrollReveal
            v-for="(plan, index) in plans"
            :key="plan.name"
            :delay="index * 100"
          >
            <div
              class="pricing-card"
              :class="{ featured: plan.featured }"
            >
              <div v-if="plan.featured" class="featured-badge">最受欢迎</div>
              <div class="card-header">
                <div class="plan-icon" :style="{ background: plan.color }">
                  {{ plan.icon }}
                </div>
                <h3>{{ plan.name }}</h3>
                <p class="plan-desc">{{ plan.description }}</p>
              </div>
              <div class="card-price">
                <div class="price-amount">
                  <span v-if="plan.price === 0" class="free-text">免费</span>
                  <template v-else>
                    <span class="currency">¥</span>
                    <span class="number">{{ plan.price }}</span>
                    <span class="period">/{{ plan.period }}</span>
                  </template>
                </div>
                <p v-if="plan.originalPrice" class="original-price">
                  原价 ¥{{ plan.originalPrice }}/{{ plan.period }}
                </p>
              </div>
              <ul class="feature-list">
                <li
                  v-for="feature in plan.features"
                  :key="feature.text"
                  :class="{ disabled: !feature.included }"
                >
                  <svg
                    v-if="feature.included"
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <polyline points="20 6 9 17 4 12"></polyline>
                  </svg>
                  <svg
                    v-else
                    viewBox="0 0 24 24"
                    fill="none"
                    stroke="currentColor"
                    stroke-width="2"
                  >
                    <line x1="18" y1="6" x2="6" y2="18"></line>
                    <line x1="6" y1="6" x2="18" y2="18"></line>
                  </svg>
                  {{ feature.text }}
                </li>
              </ul>
              <a
                v-if="plan.ctaHref"
                :href="plan.ctaHref"
                target="_blank"
                rel="noopener"
                class="plan-btn"
                :class="{ 'btn-primary': plan.featured, 'btn-outline': !plan.featured }"
              >
                {{ plan.buttonText }}
              </a>
              <a
                v-else
                :href="`mailto:${contact.bdEmail}`"
                class="plan-btn"
                :class="{ 'btn-primary': plan.featured, 'btn-outline': !plan.featured }"
              >
                {{ plan.buttonText }}
              </a>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Standalone Section -->
    <section class="section standalone-section">
      <div class="container">
        <div class="standalone-card">
          <div class="standalone-icon">🏠</div>
          <h2>私有部署版 · 数据自主可控</h2>
          <p class="standalone-desc">
            适合对数据安全有严格要求的企业，Docker Compose 一键部署，数据 100% 存储在您的服务器上。
            全功能无阉割，无限智能体、无限对话、无限知识库。
          </p>
          <div class="standalone-features">
            <div class="s-feature">
              <span class="s-feature-icon">🐳</span>
              <div>
                <strong>Docker 一键部署</strong>
                <span>一行命令启动所有服务</span>
              </div>
            </div>
            <div class="s-feature">
              <span class="s-feature-icon">🔒</span>
              <div>
                <strong>数据私有化</strong>
                <span>完全内网部署，数据不外传</span>
              </div>
            </div>
            <div class="s-feature">
              <span class="s-feature-icon">∞</span>
              <div>
                <strong>无限使用</strong>
                <span>无对话/智能体/知识库限制</span>
              </div>
            </div>
            <div class="s-feature">
              <span class="s-feature-icon">📋</span>
              <div>
                <strong>全部功能保留</strong>
                <span>工作流 / 多渠道 / RAG 全支持</span>
              </div>
            </div>
          </div>
          <div class="standalone-actions">
            <router-link to="/docs/deploy" class="btn btn-primary btn-large">
              查看部署文档
            </router-link>
            <router-link to="/advantages" class="btn btn-outline btn-large">
              了解更多
            </router-link>
          </div>
        </div>
      </div>
    </section>

    <!-- Feature Comparison -->
    <section class="section section-alt">
      <div class="container">
        <SectionTitle
          tag="功能对比"
          title="详细功能对比"
          description="全面了解各套餐的功能差异，选择最适合您的方案"
          :center="true"
        />

        <ScrollReveal>
          <div class="comparison-table">
            <div class="table-header">
              <div class="th feature-col">功能</div>
              <div class="th" v-for="plan in plans" :key="plan.name">
                <span :class="{ 'featured-name': plan.featured }">{{ plan.name }}</span>
              </div>
            </div>
            <div
              v-for="(group, gIndex) in comparisonGroups"
              :key="gIndex"
              class="comparison-group"
            >
              <div class="group-header">
                <span>{{ group.icon }} {{ group.name }}</span>
              </div>
              <div
                v-for="item in group.items"
                :key="item.name"
                class="table-row"
              >
                <div class="td feature-col">{{ item.name }}</div>
                <div
                  v-for="(val, vIndex) in item.values"
                  :key="vIndex"
                  class="td"
                >
                  <template v-if="val === true">
                    <svg class="check-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                      <polyline points="20 6 9 17 4 12"></polyline>
                    </svg>
                  </template>
                  <template v-else-if="val === false">
                    <span class="dash">—</span>
                  </template>
                  <template v-else>
                    <span>{{ val }}</span>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>

    <!-- FAQ Section -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="常见问题"
          title="关于定价的常见问题"
          description="如果您还有其他问题，欢迎联系我们"
          :center="true"
        />

        <div class="faq-list">
          <ScrollReveal
            v-for="(faq, index) in faqs"
            :key="index"
            :delay="index * 80"
          >
            <div
              class="faq-item"
              :class="{ open: openFaq === index }"
              @click="toggleFaq(index)"
            >
              <div class="faq-question">
                <span>{{ faq.question }}</span>
                <svg
                  class="faq-arrow"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                >
                  <polyline points="6 9 12 15 18 9"></polyline>
                </svg>
              </div>
              <div class="faq-answer">
                <p>{{ faq.answer }}</p>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Bottom CTA -->
    <section class="section cta-section">
      <div class="container">
        <ScrollReveal>
          <div class="cta-box">
            <h2>还有其他问题？</h2>
            <p>选择 SaaS 云服务即开即用，或私有部署数据自主可控</p>
            <div class="cta-actions">
              <router-link to="/docs/deploy" class="btn btn-primary btn-large">私有部署</router-link>
              <router-link to="/docs/guide" class="btn btn-secondary btn-large">使用指南</router-link>
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useHead } from '@unhead/vue'
import SectionTitle from '@/components/SectionTitle.vue'
import ScrollReveal from '@/components/ScrollReveal.vue'
import { useSiteConfig } from '@/composables/useSiteConfig'

const { contact, fetchContact } = useSiteConfig()

useHead({
  title: '定价方案 - 辰科 fettle | SaaS + 私有部署',
  meta: [
    { name: 'description', content: '辰科 fettle 定价方案：免费版、标准版、专业版、企业版 SaaS 云服务，及私有部署版（Docker 一键部署，数据自主可控）。' }
  ]
})

onMounted(() => {
  fetchContact()
  const ld = document.createElement('script')
  ld.type = 'application/ld+json'
  ld.textContent = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "FAQPage",
    "mainEntity": [
      { "@type": "Question", "name": "免费版有什么限制？", "acceptedAnswer": { "@type": "Answer", "text": "免费版包含 1 个智能体、100 条对话/月、1 个知识库，适合个人试用和小型项目。" } },
      { "@type": "Question", "name": "私有部署版怎么收费？", "acceptedAnswer": { "@type": "Answer", "text": "私有部署版按年授权收费，包含一年更新和技术支持。数据 100% 私有化，Docker Compose 一键部署。" } },
      { "@type": "Question", "name": "私有部署版和 SaaS 版功能一样吗？", "acceptedAnswer": { "@type": "Answer", "text": "私有部署版保留了全部核心业务功能：智能体管理、对话、多渠道接入、知识库 RAG、AI 工作流等。仅移除了多租户管理、支付计费等纯 SaaS 功能。" } }
    ]
  })
  document.head.appendChild(ld)
})

const openFaq = ref<number | null>(0)

function toggleFaq(index: number) {
  openFaq.value = openFaq.value === index ? null : index
}

const plans = [
  {
    name: '免费版',
    icon: '🆓',
    description: '适合个人开发者和小型项目试用',
    price: 0,
    period: '月',
    color: 'linear-gradient(135deg, #94a3b8, #64748b)',
    featured: false,
    buttonText: '免费开始',
    ctaHref: 'https://fettle.cenkor.cn/register',
    features: [
      { text: '1 个智能体', included: true },
      { text: '100 条对话/月', included: true },
      { text: '基础 AI 模型', included: true },
      { text: '1 个知识库', included: true },
      { text: '网页渠道接入', included: true },
      { text: '社区支持', included: true },
      { text: '高级模型', included: false },
      { text: 'API 接入', included: false },
      { text: '自定义品牌', included: false },
      { text: '优先支持', included: false }
    ]
  },
  {
    name: '标准版',
    icon: '🚀',
    description: '适合中小企业和团队日常使用',
    price: 299,
    originalPrice: 399,
    period: '月',
    color: 'linear-gradient(135deg, #3b82f6, #2563eb)',
    featured: true,
    buttonText: '立即订阅',
    ctaHref: 'https://fettle.cenkor.cn/register',
    features: [
      { text: '5 个智能体', included: true },
      { text: '5,000 条对话/月', included: true },
      { text: '全部 AI 模型', included: true },
      { text: '5 个知识库', included: true },
      { text: '多渠道接入', included: true },
      { text: 'API 接入', included: true },
      { text: '数据分析', included: true },
      { text: '邮件支持', included: true },
      { text: '自定义品牌', included: false },
      { text: '专属客服', included: false }
    ]
  },
  {
    name: '专业版',
    icon: '💎',
    description: '适合大型企业和高频使用场景',
    price: 799,
    originalPrice: 999,
    period: '月',
    color: 'linear-gradient(135deg, #8b5cf6, #6d28d9)',
    featured: false,
    buttonText: '立即订阅',
    ctaHref: 'https://fettle.cenkor.cn/register',
    features: [
      { text: '20 个智能体', included: true },
      { text: '20,000 条对话/月', included: true },
      { text: '全部 AI 模型', included: true },
      { text: '20 个知识库', included: true },
      { text: '全渠道接入', included: true },
      { text: 'API 接入', included: true },
      { text: '高级数据分析', included: true },
      { text: '自定义品牌', included: true },
      { text: '优先技术支持', included: true },
      { text: '专属客服', included: false }
    ]
  },
  {
    name: '企业版',
    icon: '🏢',
    description: '定制化部署，满足企业级需求',
    price: 0,
    period: '',
    color: 'linear-gradient(135deg, #f59e0b, #d97706)',
    featured: false,
    buttonText: '联系销售',
    features: [
      { text: '无限智能体', included: true },
      { text: '无限对话', included: true },
      { text: '全部 AI 模型', included: true },
      { text: '无限知识库', included: true },
      { text: '全渠道接入', included: true },
      { text: 'API 接入', included: true },
      { text: '高级数据分析', included: true },
      { text: '自定义品牌', included: true },
      { text: '私有化部署', included: true },
      { text: '专属客服 + SLA', included: true }
    ]
  }
]

const comparisonGroups = [
  {
    name: '基础功能',
    icon: '📦',
    items: [
      { name: '智能体数量', values: ['1 个', '5 个', '20 个', '无限'] },
      { name: '对话次数/月', values: ['100', '5,000', '20,000', '无限'] },
      { name: '知识库数量', values: ['1 个', '5 个', '20 个', '无限'] },
      { name: '文档上传', values: ['10 个', '100 个', '500 个', '无限'] },
      { name: '团队成员', values: ['1 人', '5 人', '20 人', '无限'] }
    ]
  },
  {
    name: 'AI 能力',
    icon: '🤖',
    items: [
      { name: '基础模型', values: [true, true, true, true] },
      { name: '高级模型（GPT-4 等）', values: [false, true, true, true] },
      { name: '自定义模型接入', values: [false, false, true, true] },
      { name: 'RAG 知识检索', values: [true, true, true, true] },
      { name: '多轮对话', values: [true, true, true, true] }
    ]
  },
  {
    name: '渠道与集成',
    icon: '🔗',
    items: [
      { name: '网页嵌入', values: [true, true, true, true] },
      { name: '微信公众号', values: [false, true, true, true] },
      { name: '企业微信', values: [false, true, true, true] },
      { name: '钉钉', values: [false, false, true, true] },
      { name: 'API 接入', values: [false, true, true, true] },
      { name: 'Webhook', values: [false, false, true, true] }
    ]
  },
  {
    name: '服务与支持',
    icon: '🛡️',
    items: [
      { name: '社区支持', values: [true, true, true, true] },
      { name: '邮件支持', values: [false, true, true, true] },
      { name: '优先技术支持', values: [false, false, true, true] },
      { name: '专属客服', values: [false, false, false, true] },
      { name: 'SLA 保障', values: [false, false, false, true] },
      { name: '私有化部署', values: [false, false, false, true] }
    ]
  }
]

const faqs = [
  {
    question: '免费版有什么限制？',
    answer: '免费版包含 1 个智能体、100 条对话/月、1 个知识库，适合个人试用和小型项目。如需更多功能，可升级到标准版或专业版。'
  },
  {
    question: '可以随时升级或降级套餐吗？',
    answer: '可以。您可以在账户设置中随时升级或降级套餐。升级立即生效，按比例计费；降级将在当前计费周期结束后生效。'
  },
  {
    question: '对话次数用完了怎么办？',
    answer: '对话次数用完后，您可以购买额外对话包（100 条/10 元），或直接升级到更高级别的套餐以获得更多对话额度。'
  },
  {
    question: '支持哪些支付方式？',
    answer: '我们支持微信支付、支付宝、银行转账等方式。企业版客户还可以选择对公转账和开具增值税专用发票。'
  },
  {
    question: '企业版如何定价？',
    answer: '企业版根据您的具体需求定制，包括私有化部署、定制开发、专属支持等。请联系我们的销售团队，我们会根据您的规模和需求提供报价。'
  },
  {
    question: '数据安全性如何保障？',
    answer: '所有套餐均采用端到端加密传输，API Key 使用 AES-256 加密存储。企业版支持私有化部署，数据完全存储在您的服务器上。我们通过了 ISO 27001 信息安全认证。'
  },
  {
    question: '私有部署版怎么收费？',
    answer: '私有部署版按年授权收费，包含一年更新和技术支持。详情请联系我们的销售团队获取报价。'
  },
  {
    question: '私有部署版和 SaaS 版功能一样吗？',
    answer: '私有部署版保留了全部核心业务功能：智能体管理、对话、多渠道接入、知识库 RAG、AI 工作流等。仅移除了多租户管理、支付计费等纯 SaaS 功能。'
  },
  {
    question: '是否提供发票？',
    answer: '是的，所有付费套餐均可开具增值税普通发票或专用发票。您可以在账户设置中填写开票信息，发票将在 3 个工作日内寄出。'
  }
]
</script>

<style lang="scss" scoped>
.pricing-page {
  padding-top: 80px;
}

.page-header {
  padding: $spacing-4xl 0 $spacing-3xl;
  background: linear-gradient(180deg, $gray-50 0%, $bg-white 100%);
}

.section {
  padding: $spacing-3xl 0;
}

.section-alt {
  background: $gray-50;
}

.container {
  @include container;
}

// Pricing Cards
.pricing-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: $spacing-lg;

  @include desktop {
    grid-template-columns: repeat(2, 1fr);
  }

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.pricing-card {
  background: $bg-white;
  border-radius: $radius-xl;
  padding: $spacing-xl;
  border: 1px solid $border-light;
  position: relative;
  transition: all $transition-normal;
  display: flex;
  flex-direction: column;

  &:hover {
    transform: translateY(-4px);
    box-shadow: $shadow-lg;
  }

  &.featured {
    border-color: $primary;
    box-shadow: 0 0 0 1px $primary, $shadow-lg;

    .plan-btn.btn-primary {
      @include button-primary;
      width: 100%;
      padding: 14px 24px;
      font-size: 16px;
      font-weight: 600;
      border-radius: $radius-md;
    }
  }
}

.featured-badge {
  position: absolute;
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
  background: $primary;
  color: $text-white;
  padding: 4px 16px;
  border-radius: $radius-full;
  font-size: 12px;
  font-weight: 600;
  white-space: nowrap;
}

.card-header {
  text-align: center;
  margin-bottom: $spacing-lg;

  .plan-icon {
    width: 56px;
    height: 56px;
    border-radius: $radius-lg;
    @include flex-center;
    font-size: 24px;
    margin: 0 auto $spacing-md;
  }

  h3 {
    font-size: 1.25rem;
    font-weight: 700;
    color: $text-primary;
    margin-bottom: $spacing-xs;
  }

  .plan-desc {
    font-size: 0.875rem;
    color: $text-secondary;
  }
}

.card-price {
  text-align: center;
  padding: $spacing-lg 0;
  border-top: 1px solid $border-light;
  border-bottom: 1px solid $border-light;
  margin-bottom: $spacing-lg;

  .price-amount {
    display: flex;
    align-items: baseline;
    justify-content: center;
    gap: 2px;

    .free-text {
      font-size: 2rem;
      font-weight: 700;
      color: $text-primary;
    }

    .currency {
      font-size: 1.25rem;
      font-weight: 600;
      color: $text-primary;
    }

    .number {
      font-size: 3rem;
      font-weight: 800;
      color: $text-primary;
      line-height: 1;
    }

    .period {
      font-size: 0.875rem;
      color: $text-secondary;
    }
  }

  .original-price {
    font-size: 0.8rem;
    color: $text-muted;
    text-decoration: line-through;
    margin-top: $spacing-xs;
  }
}

.feature-list {
  list-style: none;
  padding: 0;
  margin: 0 0 $spacing-xl;
  flex: 1;

  li {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
    padding: $spacing-sm 0;
    font-size: 0.9rem;
    color: $text-secondary;

    &.disabled {
      color: $text-muted;
      opacity: 0.6;
    }

    svg {
      width: 18px;
      height: 18px;
      flex-shrink: 0;
    }

    &:not(.disabled) svg {
      color: $success;
    }

    &.disabled svg {
      color: $text-muted;
    }
  }
}

.plan-btn {
  display: block;
  text-align: center;
  padding: 12px 24px;
  border-radius: $radius-md;
  font-weight: 600;
  font-size: 0.95rem;
  transition: all $transition-normal;
  text-decoration: none;

  &.btn-primary {
    background: $primary;
    color: $text-white;

    &:hover {
      background: $primary-dark;
      transform: translateY(-2px);
      box-shadow: $shadow-md;
    }
  }

  &.btn-outline {
    background: transparent;
    color: $primary;
    border: 1px solid $border-dark;

    &:hover {
      border-color: $primary;
      background: rgba($primary, 0.05);
    }
  }
}

// Standalone Section
.standalone-section {
  padding: $spacing-3xl 0;

  .container {
    @include container;
  }
}

.standalone-card {
  background: linear-gradient(135deg, $dark-800, $dark-900);
  border-radius: $radius-2xl;
  padding: $spacing-3xl $spacing-2xl;
  text-align: center;
  color: $text-white;
  border: 1px solid $gray-700;

  .standalone-icon {
    font-size: 64px;
    margin-bottom: $spacing-lg;
  }

  h2 {
    font-size: 2rem;
    font-weight: 700;
    margin-bottom: $spacing-md;
    color: $text-white;
  }

  .standalone-desc {
    font-size: 1.1rem;
    color: $gray-300;
    max-width: 600px;
    margin: 0 auto $spacing-2xl;
    line-height: 1.7;
  }
}

.standalone-features {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: $spacing-lg;
  margin-bottom: $spacing-2xl;

  @include tablet {
    grid-template-columns: repeat(2, 1fr);
  }

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.s-feature {
  background: rgba(255, 255, 255, 0.06);
  border-radius: $radius-lg;
  padding: $spacing-lg;
  display: flex;
  align-items: center;
  gap: $spacing-md;
  text-align: left;
  transition: all $transition-fast;

  &:hover {
    background: rgba(255, 255, 255, 0.1);
    transform: translateY(-2px);
  }

  .s-feature-icon {
    font-size: 36px;
    flex-shrink: 0;
  }

  strong {
    display: block;
    font-size: 1rem;
    color: $text-white;
    margin-bottom: 2px;
  }

  span {
    font-size: 0.85rem;
    color: $gray-400;
  }
}

.standalone-actions {
  display: flex;
  justify-content: center;
  gap: $spacing-md;
  flex-wrap: wrap;

  .btn-primary {
    background: #10b981;
    color: $text-white;
    padding: 14px 32px;
    border-radius: $radius-md;
    font-weight: 600;
    text-decoration: none;
    font-size: 1rem;
    transition: all $transition-fast;

    &:hover {
      background: #059669;
      transform: translateY(-2px);
      box-shadow: $shadow-lg;
    }
  }

  .btn-outline {
    background: transparent;
    color: $text-white;
    border: 1px solid $gray-500;
    padding: 14px 32px;
    border-radius: $radius-md;
    font-weight: 600;
    text-decoration: none;
    font-size: 1rem;
    transition: all $transition-fast;

    &:hover {
      background: rgba(255, 255, 255, 0.1);
      border-color: $gray-300;
    }
  }
}

// Comparison Table
.comparison-table {
  background: $bg-white;
  border-radius: $radius-xl;
  overflow: hidden;
  border: 1px solid $border-light;
  box-shadow: $shadow-sm;
}

.table-header {
  display: grid;
  grid-template-columns: 2fr repeat(4, 1fr);
  background: $dark-800;
  color: $text-white;

  .th {
    padding: $spacing-md $spacing-lg;
    text-align: center;
    font-weight: 600;
    font-size: 0.95rem;

    &.feature-col {
      text-align: left;
    }
  }

  .featured-name {
    color: $primary-light;
  }
}

.comparison-group {
  .group-header {
    background: $gray-50;
    padding: $spacing-sm $spacing-lg;
    font-weight: 600;
    font-size: 0.9rem;
    color: $text-primary;
    border-top: 1px solid $border-light;
    border-bottom: 1px solid $border-light;
  }
}

.table-row {
  display: grid;
  grid-template-columns: 2fr repeat(4, 1fr);
  border-bottom: 1px solid $border-light;
  transition: background $transition-fast;

  &:hover {
    background: $gray-50;
  }

  &:last-child {
    border-bottom: none;
  }

  .td {
    padding: $spacing-sm $spacing-lg;
    text-align: center;
    font-size: 0.9rem;
    color: $text-secondary;

    &.feature-col {
      text-align: left;
      font-weight: 500;
      color: $text-primary;
    }
  }

  .check-icon {
    width: 20px;
    height: 20px;
    color: $success;
  }

  .dash {
    color: $text-muted;
  }
}

// Responsive table
@include mobile {
  .table-header,
  .table-row {
    grid-template-columns: 1.5fr repeat(4, 1fr);

    .td,
    .th {
      padding: $spacing-sm $spacing-xs;
      font-size: 0.75rem;
    }
  }
}

// FAQ
.faq-list {
  max-width: 800px;
  margin: 0 auto;
}

.faq-item {
  background: $bg-white;
  border: 1px solid $border-light;
  border-radius: $radius-lg;
  margin-bottom: $spacing-md;
  cursor: pointer;
  transition: all $transition-normal;
  overflow: hidden;

  &:hover {
    border-color: $primary;
  }

  &.open {
    border-color: $primary;
    box-shadow: 0 0 0 1px $primary;
  }
}

.faq-question {
  @include flex-between;
  padding: $spacing-lg $spacing-xl;
  font-weight: 600;
  color: $text-primary;
  font-size: 1rem;

  .faq-arrow {
    width: 20px;
    height: 20px;
    color: $text-muted;
    transition: transform $transition-normal;
    flex-shrink: 0;
  }

  .open & .faq-arrow {
    transform: rotate(180deg);
    color: $primary;
  }
}

.faq-answer {
  max-height: 0;
  overflow: hidden;
  transition: max-height $transition-slow;

  .open & {
    max-height: 300px;
  }

  p {
    padding: 0 $spacing-xl $spacing-lg;
    color: $text-secondary;
    line-height: 1.7;
    font-size: 0.95rem;
  }
}

// CTA
.cta-section {
  padding: $spacing-4xl 0;
}

.cta-box {
  @include flex-column;
  align-items: center;
  text-align: center;
  background: linear-gradient(135deg, $dark-800, $dark-900);
  border-radius: $radius-2xl;
  padding: $spacing-3xl $spacing-xl;
  color: $text-white;

  h2 {
    font-size: 2rem;
    font-weight: 700;
    margin-bottom: $spacing-md;
  }

  p {
    font-size: 1.1rem;
    color: $gray-300;
    margin-bottom: $spacing-xl;
    max-width: 500px;
  }

  .cta-actions {
    display: flex;
    gap: $spacing-md;

    @include mobile {
      flex-direction: column;
      width: 100%;
      max-width: 300px;
    }
  }

  .btn {
    @include button-base;
    padding: 14px 32px;
    font-size: 1rem;
    font-weight: 600;
    border-radius: $radius-md;
    text-decoration: none;

    &.btn-primary {
      background: $primary;
      color: $text-white;

      &:hover {
        background: $primary-dark;
        transform: translateY(-2px);
        box-shadow: $shadow-lg;
      }
    }

    &.btn-secondary {
      background: transparent;
      color: $text-white;
      border: 1px solid $gray-500;

      &:hover {
        background: rgba(255, 255, 255, 0.1);
        border-color: $gray-300;
      }
    }
  }
}
</style>
