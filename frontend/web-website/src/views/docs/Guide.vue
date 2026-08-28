<template>
  <div class="guide-page">
    <!-- Page Header -->
    <section class="page-header">
      <div class="container">
        <SectionTitle
          tag="使用指南"
          title="快速上手教程"
          description="SaaS 云服务与私有部署双版本文档，从登录到使用，一步步教您使用 AI 智能体平台"
          :center="true"
        />
      </div>
    </section>

    <!-- Version Entry -->
    <section class="section section-entry">
      <div class="container">
        <div class="entry-grid">
          <ScrollReveal animation="fade-left">
            <div class="entry-card saas-card">
              <div class="entry-icon">☁️</div>
              <h3>SaaS 云服务版</h3>
              <p class="entry-desc">即开即用，免运维，适合快速体验和中小团队</p>
              <div class="entry-steps">
                <div class="entry-step">
                  <span class="step-badge">1</span>
                  <span>访问官网，点击"免费试用"注册账号</span>
                </div>
                <div class="entry-step">
                  <span class="step-badge">2</span>
                  <span>登录管理后台，配置 AI 模型 API Key</span>
                </div>
                <div class="entry-step">
                  <span class="step-badge">3</span>
                  <span>创建智能体，开始对话</span>
                </div>
              </div>
            </div>
          </ScrollReveal>
          <ScrollReveal animation="fade-right">
            <div class="entry-card standalone-card">
              <div class="entry-icon">🏠</div>
              <h3>私有部署版</h3>
              <p class="entry-desc">数据自主可控，Docker 一键部署，适合对安全有要求的企业</p>
              <div class="entry-steps">
                <div class="entry-step">
                  <span class="step-badge">1</span>
                  <span>部署服务：<code>docker compose up -d</code></span>
                </div>
                <div class="entry-step">
                  <span class="step-badge">2</span>
                  <span>浏览器访问 <code>http://服务器IP:20009</code></span>
                </div>
                <div class="entry-step">
                  <span class="step-badge">3</span>
                  <span>使用默认账号 <code>superadmin / CHANGE_ME_ADMIN_PASSWORD</code> 登录</span>
                </div>
              </div>
              <router-link to="/docs/deploy" class="entry-link">查看部署文档 →</router-link>
            </div>
          </ScrollReveal>
        </div>
      </div>
    </section>

    <!-- Quick Start -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="快速开始"
          title="5 分钟快速上手"
          description="以下步骤 SaaS 云服务和私有部署版通用"
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
              <div class="category-header" style="background: linear-gradient(135deg, #f59e0b, #d97706)">
                <span class="category-icon">⚡</span>
                <h3>AI 工作流</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in workflowGuides" :key="item.title">
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
              <div class="category-header" style="background: linear-gradient(135deg, #06b6d4, #0891b2)">
                <span class="category-icon">📡</span>
                <h3>渠道管理</h3>
              </div>
              <div class="category-content">
                <div class="guide-item" v-for="item in channelGuides" :key="item.title">
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
              <div class="category-header" style="background: linear-gradient(135deg, #8b5cf6, #6d28d9)">
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
          <h2>开始使用 AI 智能体平台</h2>
          <p>SaaS 云服务即开即用，或私有部署数据自主可控</p>
          <div class="cta-actions">
            <router-link to="/pricing" class="btn btn-primary btn-large">查看方案</router-link>
            <router-link to="/docs/deploy" class="btn btn-secondary btn-large">私有部署</router-link>
          </div>
        </ScrollReveal>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useHead } from '@unhead/vue'
import SectionTitle from '@/components/SectionTitle.vue'
import ScrollReveal from '@/components/ScrollReveal.vue'
import CodeBlock from '@/components/CodeBlock.vue'

useHead({
  title: '使用指南 - 辰科 fettle | AI 智能体平台教程',
  meta: [
    { name: 'description', content: '辰科 fettle 使用教程：SaaS 云服务与私有部署版操作指南，涵盖智能体创建、知识库 RAG、AI 工作流编排、渠道管理（微信公众号/企微/飞书/钉钉/抖音）、API 接入。' }
  ]
})

onMounted(() => {
  const ld = document.createElement('script')
  ld.type = 'application/ld+json'
  ld.textContent = JSON.stringify({
    "@context": "https://schema.org",
    "@type": "FAQPage",
    "mainEntity": [
      { "@type": "Question", "name": "SaaS 版和私有部署版的后台操作一样吗？", "acceptedAnswer": { "@type": "Answer", "text": "核心功能操作完全一致（智能体/对话/知识库/工作流/渠道管理）。差异在于 SaaS 版有租户管理和套餐计费，私有部署版使用固定超级管理员账号。" } },
      { "@type": "Question", "name": "私有部署版默认密码是多少？", "acceptedAnswer": { "@type": "Answer", "text": "默认管理员账号 superadmin，密码 CHANGE_ME_ADMIN_PASSWORD。首次登录后请立即在系统设置中修改密码。" } },
      { "@type": "Question", "name": "工作流的 Cron 表达式怎么写？", "acceptedAnswer": { "@type": "Answer", "text": "标准 5 段式 Cron 表达式：分(0-59) 时(0-23) 日(1-31) 月(1-12) 周(0-7)。例如 0 9 * * * 表示每天早上 9 点执行。" } }
    ]
  })
  document.head.appendChild(ld)
})

const quickStartSteps = [
  {
    icon: '🔑',
    title: '登录后台',
    description: 'SaaS 版通过官网登录，私有部署版访问 http://服务器IP:20009，使用 superadmin 账号登录'
  },
  {
    icon: '🤖',
    title: '配置 AI 模型',
    description: '进入"AI 模型配置"页面，添加 DeepSeek、通义千问等模型的 API Key，测试连接'
  },
  {
    icon: '🧠',
    title: '创建智能体',
    description: '进入智能体管理，创建您的第一个 AI 智能体，设置人设和关联知识库'
  },
  {
    icon: '💬',
    title: '开始对话',
    description: '选择智能体开始对话，体验 SSE 流式响应的打字机效果'
  },
  {
    icon: '🔗',
    title: '接入渠道',
    description: '在渠道管理中将智能体接入微信公众号/企业微信/飞书/钉钉/抖音'
  }
]

const accountGuides = [
  {
    title: 'SaaS 版：注册与登录',
    description: '使用云服务版本',
    steps: [
      '访问平台官网，点击"免费试用"',
      '填写企业信息、设置登录密码',
      '完成邮箱验证',
      '登录管理后台，首次进入需完善租户信息',
      '进入"AI 模型配置"页面添加 API Key'
    ]
  },
  {
    title: '私有部署版：首次登录',
    description: 'Docker 部署完成后首次使用',
    steps: [
      '确保服务已启动（docker compose ps）',
      '浏览器访问 http://服务器IP:20009',
      '使用默认账号登录：superadmin / CHANGE_ME_ADMIN_PASSWORD',
      '进入"AI 模型配置"页面添加 API Key（通义千问、DeepSeek 等）',
      '强烈建议立即修改默认密码'
    ]
  },
  {
    title: '配置 AI 模型',
    description: '设置您使用的 AI 模型（两版通用）',
    steps: [
      '进入"AI 模型配置"页面',
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

const workflowGuides = [
  {
    title: '创建工作流',
    description: '可视化编排 AI 自动化任务',
    steps: [
      '进入"工作流管理"页面，点击"新建工作流"',
      '输入工作流名称和描述',
      '从左侧调色板拖拽节点到画布（支持 LLM / RAG / Condition / HTTP 等 9 种节点）',
      '连接节点：从输出端口拖拽连线到输入端口',
      '配置每个节点的参数（如 LLM 节点选择模型和提示词）',
      '保存并发布工作流'
    ]
  },
  {
    title: '配置定时触发器',
    description: '设置 Cron 表达式定时执行工作流',
    steps: [
      '在编辑器右侧面板点击"触发器设置"',
      '选择"定时触发 (Cron)"',
      '输入 Cron 表达式（5 段式：分 时 日 月 周）',
      '选择时区（默认 Asia/Shanghai）',
      '保存后调度器将自动按计划执行'
    ]
  },
  {
    title: '配置 Webhook 触发器',
    description: '通过外部 HTTP 请求触发工作流',
    steps: [
      '在编辑器右侧面板点击"触发器设置"',
      '选择"Webhook 触发"',
      '保存后系统生成唯一的 Webhook URL',
      '外部系统通过 POST 请求调用该 URL 触发工作流',
      'Webhook 路径格式：POST /api/v1/webhook/trigger/{workflow_id}'
    ]
  },
  {
    title: '查看运行记录',
    description: '追踪工作流执行状态和结果',
    steps: [
      '进入工作流列表，点击工作流名称',
      '切换到"运行记录"标签页',
      '查看每次执行的开始时间、状态和输出',
      '点击单次记录查看详细执行日志',
      '执行失败时可查看错误信息进行排查'
    ]
  }
]

const channelGuides = [
  {
    title: '微信公众号接入',
    description: '将智能体接入微信公众号',
    steps: [
      '进入"渠道管理"页面，选择"微信渠道"',
      '填写微信公众号 AppID 和 AppSecret',
      '配置公众号服务器地址（指向平台网关）',
      '在微信公众号后台设置 IP 白名单',
      '保存配置，扫码关注测试对话效果'
    ]
  },
  {
    title: '企业微信接入',
    description: '将智能体接入企业微信应用',
    steps: [
      '进入"渠道管理"页面，选择"企业微信"',
      '在企业微信后台创建自建应用',
      '填写 CorpID、AgentID 和 Secret',
      '配置企业微信回调 URL',
      '保存后企业微信成员可直接与 AI 智能体对话'
    ]
  },
  {
    title: '飞书/钉钉/抖音接入',
    description: '其他第三方平台接入',
    steps: [
      '进入"渠道管理"页面，选择对应平台',
      '按照平台指引创建应用并获取凭证',
      '填写相关配置参数',
      '配置消息回调 URL',
      '保存并测试连接'
    ]
  },
  {
    title: '网页嵌入',
    description: '将智能体嵌入到您的网站',
    steps: [
      '进入"渠道管理"页面，选择"网页嵌入"',
      '选择要嵌入的智能体',
      '自定义对话窗口样式（主题色、位置等）',
      '复制生成的嵌入代码',
      '将代码粘贴到您网站 HTML 的 &lt;/body&gt; 标签前'
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
    title: '工作流编排技巧',
    description: '合理设计工作流提升自动化效率',
    tips: [
      '从简单场景开始，逐步增加复杂度',
      '善用 Condition 节点实现分支逻辑',
      'HTTP 节点可对接外部 API 扩展能力',
      '定时任务建议先手动执行测试'
    ]
  },
  {
    icon: '📊',
    title: '数据分析',
    description: '利用数据持续优化智能体',
    tips: [
      '定期查看对话统计和分析页面',
      '分析用户常见问题优化知识库',
      '监控渠道接入的使用情况',
      '查看工作流执行记录排查失败任务'
    ]
  }
]

const faqs = [
  {
    question: 'Q: SaaS 版和私有部署版的后台操作一样吗？',
    answer: 'A: 核心功能操作完全一致（智能体/对话/知识库/工作流/渠道管理）。差异在于：SaaS 版有租户管理和套餐计费，私有部署版使用固定超级管理员账号。'
  },
  {
    question: 'Q: 如何选择合适的 AI 模型？',
    answer: 'A: 通义千问适合中文场景，DeepSeek 性价比高，GPT-4 综合能力强。建议先在 AI 模型配置中测试不同模型的效果再决定。'
  },
  {
    question: 'Q: 私有部署版默认密码是多少？',
    answer: 'A: 默认管理员账号 superadmin，密码 CHANGE_ME_ADMIN_PASSWORD。首次登录后请立即在"系统设置"中修改密码。'
  },
  {
    question: 'Q: 知识库文档有大小限制吗？',
    answer: 'A: 单个文档最大 50MB，支持 PDF、Word、TXT 格式。建议将大文档拆分成多个小文件上传。'
  },
  {
    question: 'Q: 如何提高 AI 回答的准确性？',
    answer: 'A: 1. 优化知识库内容质量；2. 调整分块大小（建议 500-1000 字符）；3. 完善智能体人设描述；4. 使用工作流编排复杂任务。'
  },
  {
    question: 'Q: 工作流的 Cron 表达式怎么写？',
    answer: 'A: 标准 5 段式 Cron 表达式：分(0-59) 时(0-23) 日(1-31) 月(1-12) 周(0-7)。例如 "0 9 * * *" 表示每天早上 9 点执行。'
  },
  {
    question: 'Q: 如何接入微信公众号？',
    answer: 'A: 在"渠道管理"中选择"微信渠道"，填写公众号 AppID 和 AppSecret，配置服务器地址指向网关地址，并在公众号后台设置 IP 白名单。'
  },
  {
    question: 'Q: API 调用地址是什么？',
    answer: 'A: SaaS 版使用平台提供的 API 域名；私有部署版使用 http://服务器IP:20001（网关端口），前端 baseURL 为 /api/v1。'
  }
]
</script>

<style lang="scss" scoped>
.page-header {
  padding: 140px 0 60px;
  background: linear-gradient(180deg, $gray-50 0%, $bg-white 100%);
}

.section-entry {
  padding: $spacing-lg 0 $spacing-2xl;

  .container {
    @include container;
  }
}

.entry-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-2xl;
  max-width: 900px;
  margin: 0 auto;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.entry-card {
  background: $bg-white;
  border-radius: $radius-xl;
  padding: $spacing-2xl;
  border: 1px solid $border-light;
  transition: all $transition-normal;

  &:hover {
    transform: translateY(-4px);
    box-shadow: $shadow-lg;
  }

  &.saas-card {
    border-top: 4px solid $primary;
  }

  &.standalone-card {
    border-top: 4px solid $success;
  }

  .entry-icon {
    font-size: 48px;
    margin-bottom: $spacing-md;
  }

  h3 {
    font-size: 1.5rem;
    font-weight: 700;
    color: $text-primary;
    margin-bottom: $spacing-sm;
  }

  .entry-desc {
    color: $text-secondary;
    font-size: 0.95rem;
    line-height: 1.6;
    margin-bottom: $spacing-lg;
  }

  .entry-steps {
    display: flex;
    flex-direction: column;
    gap: $spacing-md;
    margin-bottom: $spacing-lg;
  }

  .entry-step {
    display: flex;
    align-items: center;
    gap: $spacing-md;
    font-size: 0.95rem;
    color: $text-primary;

    .step-badge {
      width: 28px;
      height: 28px;
      background: $gray-100;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-weight: 700;
      font-size: 0.85rem;
      color: $primary;
      flex-shrink: 0;
    }

    code {
      background: $gray-100;
      padding: 2px 8px;
      border-radius: 4px;
      font-size: 0.9rem;
      color: $error;
    }
  }

  .entry-link {
    color: $success;
    font-weight: 500;
    text-decoration: none;
    font-size: 0.95rem;

    &:hover {
      text-decoration: underline;
    }
  }
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
