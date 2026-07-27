<template>
  <div class="guide-page">
    <!-- Page Header -->
    <section class="page-header">
      <div class="container">
        <SectionTitle
          tag="使用指南"
          title="快速上手教程"
          description="从注册到使用，一步步教您如何使用 AI 智能体平台"
          :center="true"
        />
      </div>
    </section>

    <!-- Quick Start -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="快速开始"
          title="5 分钟快速上手"
          description="跟随以下步骤，快速体验 AI 智能体平台的核心功能"
          :center="true"
        />

        <div class="quick-start">
          <ScrollReveal>
            <div class="start-steps">
              <div class="start-step" v-for="(step, index) in quickStartSteps" :key="index">
                <div class="step-icon">{{ step.icon }}</div>
                <div class="step-content">
                  <h3>{{ step.title }}</h3>
                  <p>{{ step.description }}</p>
                </div>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Detailed Guides -->
    <section class="section section-alt">
      <div class="container">
        <SectionTitle
          tag="详细指南"
          title="功能使用详解"
          description="深入了解每个功能的使用方法"
          :center="true"
        />

        <div class="guide-categories">
          <ScrollReveal>
            <div class="category-card">
              <div class="category-header">
                <span class="category-icon">👤</span>
                <h3>账户管理</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in accountGuides" :key="item.title">
                  <h4>{{ item.title }}</h4>
                  <p>{{ item.description }}</p>
                  <ul>
                    <li v-for="step in item.steps" :key="step">{{ step }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="category-card">
              <div class="category-header">
                <span class="category-icon">🤖</span>
                <h3>智能体配置</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in agentGuides" :key="item.title">
                  <h4>{{ item.title }}</h4>
                  <p>{{ item.description }}</p>
                  <ul>
                    <li v-for="step in item.steps" :key="step">{{ step }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="category-card">
              <div class="category-header">
                <span class="category-icon">📚</span>
                <h3>知识库管理</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in knowledgeGuides" :key="item.title">
                  <h4>{{ item.title }}</h4>
                  <p>{{ item.description }}</p>
                  <ul>
                    <li v-for="step in item.steps" :key="step">{{ step }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="category-card">
              <div class="category-header">
                <span class="category-icon">💬</span>
                <h3>对话使用</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in chatGuides" :key="item.title">
                  <h4>{{ item.title }}</h4>
                  <p>{{ item.description }}</p>
                  <ul>
                    <li v-for="step in item.steps" :key="step">{{ step }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="category-card">
              <div class="category-header">
                <span class="category-icon">⚡</span>
                <h3>技能管理</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in skillGuides" :key="item.title">
                  <h4>{{ item.title }}</h4>
                  <p>{{ item.description }}</p>
                  <ul>
                    <li v-for="step in item.steps" :key="step">{{ step }}</li>
                  </ul>
                </div>
              </div>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="category-card">
              <div class="category-header">
                <span class="category-icon">🔌</span>
                <h3>API 接入</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in apiGuides" :key="item.title">
                  <h4>{{ item.title }}</h4>
                  <p>{{ item.description }}</p>
                  <CodeBlock :code="item.code" :language="item.language" />
                </div>
              </div>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Best Practices -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="最佳实践"
          title="使用技巧与建议"
          description="帮助您更好地使用 AI 智能体平台"
          :center="true"
        />

        <div class="practices-grid">
          <ScrollReveal v-for="(practice, index) in bestPractices" :key="index">
            <div class="practice-card">
              <div class="practice-icon">{{ practice.icon }}</div>
              <h3>{{ practice.title }}</h3>
              <p>{{ practice.description }}</p>
              <ul>
                <li v-for="tip in practice.tips" :key="tip">{{ tip }}</li>
              </ul>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- FAQ -->
    <section class="section section-alt">
      <div class="container">
        <SectionTitle
          tag="常见问题"
          title="使用 FAQ"
          description="解答使用过程中的常见问题"
          :center="true"
        />

        <div class="faq-grid">
          <ScrollReveal v-for="(faq, index) in faqs" :key="index">
            <div class="faq-card">
              <h4>{{ faq.question }}</h4>
              <p>{{ faq.answer }}</p>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="cta-section">
      <div class="container">
        <ScrollReveal>
          <h2>开始您的 AI 之旅</h2>
          <p>立即体验 AI 智能体平台，让智能对话为您的业务赋能</p>
          <div class="cta-actions">
            <a href="#" class="btn btn-primary btn-large">立即开始</a>
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
import CodeBlock from '@/components/CodeBlock.vue'

const quickStartSteps = [
  {
    icon: '📝',
    title: '注册账户',
    description: '访问平台官网，点击"免费试用"按钮，填写基本信息完成注册'
  },
  {
    icon: '🤖',
    title: '创建智能体',
    description: '进入管理后台，创建您的第一个 AI 智能体，设置名称和人设'
  },
  {
    icon: '📚',
    title: '上传知识库',
    description: '上传企业文档到知识库，让 AI 基于您的私有数据回答问题'
  },
  {
    icon: '💬',
    title: '开始对话',
    description: '选择智能体，开始对话体验，感受流式响应的流畅效果'
  },
  {
    icon: '🔗',
    title: '接入渠道',
    description: '将智能体接入您的网站、微信或其他渠道，扩大服务范围'
  }
]

const accountGuides = [
  {
    title: '注册账户',
    description: '创建您的平台账户',
    steps: [
      '访问平台官网',
      '点击"免费试用"按钮',
      '填写企业名称、联系人信息',
      '设置登录密码',
      '完成邮箱验证',
      '登录管理后台'
    ]
  },
  {
    title: '配置 AI 模型',
    description: '设置您使用的 AI 模型',
    steps: [
      '进入"AI 配置"页面',
      '选择 AI 模型厂商（通义千问、DeepSeek 等）',
      '输入 API Key',
      '测试连接是否成功',
      '保存配置'
    ]
  }
]

const agentGuides = [
  {
    title: '创建智能体',
    description: '创建您的第一个 AI 智能体',
    steps: [
      '进入"智能体管理"页面',
      '点击"新建智能体"按钮',
      '填写智能体名称',
      '设置人设描述（如：专业客服助手）',
      '选择关联的知识库',
      '选择安装的技能',
      '保存并启用智能体'
    ]
  },
  {
    title: '配置智能体',
    description: '自定义智能体的行为和外观',
    steps: [
      '设置对话开场白',
      '配置回复风格（正式/友好/专业）',
      '设置最大对话轮数',
      '配置超时时间',
      '自定义外观主题色'
    ]
  }
]

const knowledgeGuides = [
  {
    title: '创建知识库',
    description: '创建企业专属知识库',
    steps: [
      '进入"知识库管理"页面',
      '点击"新建知识库"',
      '填写知识库名称和描述',
      '设置分块大小（建议 500-1000 字符）',
      '选择向量化模型',
      '保存知识库'
    ]
  },
  {
    title: '上传文档',
    description: '将企业文档添加到知识库',
    steps: [
      '选择目标知识库',
      '点击"上传文档"按钮',
      '选择文件（支持 PDF、Word、TXT）',
      '等待文档解析和向量化',
      '查看文档处理状态',
      '验证文档是否可检索'
    ]
  }
]

const chatGuides = [
  {
    title: '开始对话',
    description: '与智能体进行对话',
    steps: [
      '进入"对话"页面',
      '选择要对话的智能体',
      '在输入框中输入问题',
      '等待 AI 回复（流式显示）',
      '继续提问或结束对话'
    ]
  },
  {
    title: '管理对话历史',
    description: '查看和管理历史对话',
    steps: [
      '进入"对话历史"页面',
      '查看所有历史对话列表',
      '点击对话查看详情',
      '导出对话记录',
      '删除不需要的对话'
    ]
  }
]

const skillGuides = [
  {
    title: '安装技能',
    description: '为智能体安装技能',
    steps: [
      '进入"技能市场"页面',
      '浏览可用技能列表',
      '点击"安装"按钮',
      '选择要安装到的智能体',
      '确认安装',
      '技能立即可用'
    ]
  },
  {
    title: '使用技能',
    description: '在对话中使用技能',
    steps: [
      '技能会自动在对话中触发',
      '例如：安装"网页搜索"技能后，AI 可以搜索实时信息',
      '查看技能使用日志',
      '根据需求调整技能配置'
    ]
  }
]

const apiGuides = [
  {
    title: '获取 API Key',
    description: '获取 API 访问凭证',
    steps: [
      '进入"API 管理"页面',
      '点击"创建 API Key"',
      '设置 Key 名称和权限',
      '复制生成的 API Key',
      '妥善保存（只显示一次）'
    ],
    code: '',
    language: ''
  },
  {
    title: '调用对话 API',
    description: '通过 API 调用智能体对话',
    steps: [],
    code: `curl -X POST https://api.your-domain.com/api/v1/chat \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -H "Content-Type: application/json" \\
  -d '{
    "agent_id": "your-agent-id",
    "message": "你好，请帮我查询信息",
    "stream": true
  }'`,
    language: 'bash'
  },
  {
    title: 'SSE 流式响应',
    description: '处理流式响应数据',
    steps: [],
    code: `// JavaScript 示例
const response = await fetch('/api/v1/chat', {
  method: 'POST',
  headers: {
    'Authorization': 'Bearer YOUR_API_KEY',
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    agent_id: 'your-agent-id',
    message: '你好',
    stream: true
  })
});

const reader = response.body.getReader();
const decoder = new TextDecoder();

while (true) {
  const { done, value } = await reader.read();
  if (done) break;
  
  const chunk = decoder.decode(value);
  console.log(chunk);
}`,
    language: 'javascript'
  }
]

const bestPractices = [
  {
    icon: '📝',
    title: '优化人设描述',
    description: '清晰的人设描述能让 AI 更好地理解和回应',
    tips: [
      '明确 AI 的角色定位',
      '描述目标用户群体',
      '说明回复风格要求',
      '列出禁止行为'
    ]
  },
  {
    icon: '📚',
    title: '知识库优化',
    description: '高质量的知识库带来更准确的回答',
    tips: [
      '上传结构清晰的文档',
      '定期更新知识库内容',
      '删除过时信息',
      '使用合适的分块大小'
    ]
  },
  {
    icon: '⚡',
    title: '技能组合',
    description: '合理组合技能提升智能体能力',
    tips: [
      '根据场景选择技能',
      '避免安装过多技能',
      '测试技能效果',
      '监控技能使用频率'
    ]
  },
  {
    icon: '📊',
    title: '数据分析',
    description: '利用数据持续优化智能体',
    tips: [
      '定期查看对话统计',
      '分析用户常见问题',
      '优化知识库覆盖度',
      '调整回复策略'
    ]
  }
]

const faqs = [
  {
    question: 'Q: 如何选择合适的 AI 模型？',
    answer: 'A: 根据使用场景选择：通义千问适合中文场景，DeepSeek 性价比高，GPT-4 综合能力最强。建议先试用再决定。'
  },
  {
    question: 'Q: 知识库文档有大小限制吗？',
    answer: 'A: 单个文档最大 50MB，支持 PDF、Word、TXT 格式。建议将大文档拆分成多个小文件上传。'
  },
  {
    question: 'Q: 如何提高 AI 回答的准确性？',
    answer: 'A: 1. 优化知识库内容质量；2. 调整分块大小；3. 完善智能体人设描述；4. 安装相关技能增强能力。'
  },
  {
    question: 'Q: 对话历史会保存多久？',
    answer: 'A: 默认保存 30 天，可在设置中调整。重要对话建议导出备份。'
  },
  {
    question: 'Q: 如何接入微信公众号？',
    answer: 'A: 在"渠道管理"中选择"微信接入"，按照指引配置公众号参数，完成授权即可。'
  },
  {
    question: 'Q: API 调用有频率限制吗？',
    answer: 'A: 免费版每分钟 10 次，标准版每分钟 60 次，专业版每分钟 300 次。企业版可定制。'
  }
]
</script>

<style lang="scss" scoped>
.page-header {
  padding: 140px 0 60px;
  background: linear-gradient(180deg, $gray-50 0%, $bg-white 100%);
}

.quick-start {
  max-width: 800px;
  margin: 0 auto;
}

.start-steps {
  display: flex;
  flex-direction: column;
  gap: $spacing-xl;
}

.start-step {
  display: flex;
  gap: $spacing-lg;
  padding: $spacing-xl;
  background: $bg-white;
  border-radius: $radius-lg;
  border: 1px solid $border-light;
  transition: all $transition-normal;

  &:hover {
    transform: translateX(8px);
    box-shadow: $shadow-lg;
    border-color: $primary;
  }

  .step-icon {
    font-size: 48px;
    flex-shrink: 0;
  }

  .step-content {
    h3 {
      font-size: 1.5rem;
      margin-bottom: $spacing-sm;
      color: $text-primary;
    }

    p {
      color: $text-secondary;
      line-height: 1.6;
      margin: 0;
    }
  }
}

.guide-categories {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-xl;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.category-card {
  background: $bg-white;
  border-radius: $radius-xl;
  overflow: hidden;
  box-shadow: $shadow-md;
  border: 1px solid $border-light;
}

.category-header {
  padding: $spacing-xl;
  background: linear-gradient(135deg, $primary 0%, $primary-dark 100%);
  color: white;
  display: flex;
  align-items: center;
  gap: $spacing-md;

  .category-icon {
    font-size: 40px;
  }

  h3 {
    font-size: 1.5rem;
    margin: 0;
  }
}

.category-content {
  padding: $spacing-xl;
}

.guide-item {
  margin-bottom: $spacing-xl;

  &:last-child {
    margin-bottom: 0;
  }

  h4 {
    font-size: 1.125rem;
    margin-bottom: $spacing-sm;
    color: $text-primary;
  }

  p {
    color: $text-secondary;
    margin-bottom: $spacing-md;
  }

  ul {
    margin-left: $spacing-lg;
    color: $text-secondary;

    li {
      margin-bottom: $spacing-xs;
      line-height: 1.6;
    }
  }
}

.practices-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-xl;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.practice-card {
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

  .practice-icon {
    font-size: 48px;
    margin-bottom: $spacing-md;
  }

  h3 {
    font-size: 1.25rem;
    margin-bottom: $spacing-sm;
    color: $text-primary;
  }

  p {
    color: $text-secondary;
    margin-bottom: $spacing-md;
  }

  ul {
    list-style: none;
    padding: 0;
    margin: 0;

    li {
      padding: $spacing-sm 0;
      color: $text-secondary;
      display: flex;
      align-items: center;
      gap: $spacing-sm;

      &::before {
        content: '✓';
        color: $success;
        font-weight: 700;
      }
    }
  }
}

.faq-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-lg;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.faq-card {
  background: $bg-white;
  border-radius: $radius-lg;
  padding: $spacing-xl;
  border: 1px solid $border-light;

  h4 {
    font-size: 1.125rem;
    margin-bottom: $spacing-sm;
    color: $text-primary;
  }

  p {
    color: $text-secondary;
    line-height: 1.6;
    margin: 0;
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
