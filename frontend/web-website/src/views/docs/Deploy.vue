<template>
  <div class="deploy-page">
    <!-- Page Header -->
    <section class="page-header">
      <div class="container">
        <SectionTitle
          tag="部署教程"
          title="快速部署指南"
          description="提供多种部署方式，选择最适合您的方案，快速搭建 AI 智能体平台"
          :center="true"
        />
      </div>
    </section>

    <!-- Environment Requirements -->
    <section class="section">
      <div class="container">
        <ScrollReveal>
          <div class="requirements-card">
            <h2>📋 环境要求</h2>
            <div class="requirements-grid">
              <div class="requirement-item">
                <div class="req-icon">🔵</div>
                <div class="req-info">
                  <strong>Go 1.22+</strong>
                  <span>后端服务运行时</span>
                </div>
              </div>
              <div class="requirement-item">
                <div class="req-icon">🐍</div>
                <div class="req-info">
                  <strong>Python 3.11+</strong>
                  <span>AI 引擎运行时</span>
                </div>
              </div>
              <div class="requirement-item">
                <div class="req-icon">💚</div>
                <div class="req-info">
                  <strong>Node.js 18+</strong>
                  <span>前端构建工具</span>
                </div>
              </div>
              <div class="requirement-item">
                <div class="req-icon">🐘</div>
                <div class="req-info">
                  <strong>PostgreSQL 16+</strong>
                  <span>主数据库</span>
                </div>
              </div>
              <div class="requirement-item">
                <div class="req-icon">🔴</div>
                <div class="req-info">
                  <strong>Redis 7+</strong>
                  <span>缓存与会话存储</span>
                </div>
              </div>
              <div class="requirement-item">
                <div class="req-icon">📨</div>
                <div class="req-info">
                  <strong>NATS</strong>
                  <span>消息队列（可选）</span>
                </div>
              </div>
            </div>
          </div>
        </ScrollReveal>
      </div>
    </section>

    <!-- Deployment Methods -->
    <section class="section section-alt">
      <div class="container">
        <SectionTitle
          tag="部署方式"
          title="选择您的部署方案"
          description="根据您的需求选择最合适的部署方式"
          :center="true"
        />

        <div class="deploy-tabs">
          <div class="tab-nav">
            <button
              v-for="tab in tabs"
              :key="tab.id"
              :class="['tab-btn', { active: activeTab === tab.id }]"
              @click="activeTab = tab.id"
            >
              <span class="tab-icon">{{ tab.icon }}</span>
              {{ tab.label }}
            </button>
          </div>

          <div class="tab-content">
            <!-- 独立私有部署 -->
            <div v-if="activeTab === 'standalone'" class="tab-panel">
              <ScrollReveal>
                <div class="deploy-section">
                  <h3>🏠 私有部署版（推荐）</h3>
                  <p class="deploy-desc">基于 Docker Compose 一键部署，数据 100% 私有化，适合对数据安全要求高的企业</p>

                  <div class="steps">
                    <div class="step">
                      <div class="step-number">1</div>
                      <div class="step-content">
                        <h4>安装 Docker 和 Docker Compose</h4>
                        <CodeBlock :code="dockerInstallCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">2</div>
                      <div class="step-content">
                        <h4>获取私有部署包</h4>
                        <p>从 releases 页面下载 <code>fettle-standalone</code> 部署包，或直接克隆仓库：</p>
                        <CodeBlock :code="standaloneCloneCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">3</div>
                      <div class="step-content">
                        <h4>配置环境变量</h4>
                        <p>复制环境变量模板，修改 JWT_SECRET：</p>
                        <CodeBlock :code="standaloneEnvCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">4</div>
                      <div class="step-content">
                        <h4>一键启动</h4>
                        <p>所有服务将在后台自动启动：</p>
                        <CodeBlock :code="standaloneUpCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">5</div>
                      <div class="step-content">
                        <h4>访问管理后台</h4>
                        <p>打开浏览器访问 <code>http://服务器IP:20009</code>，使用默认账号登录：</p>
                        <div class="login-info">
                          <p><strong>账号：</strong><code>superadmin</code></p>
                          <p><strong>密码：</strong><code>CHANGE_ME_ADMIN_PASSWORD</code></p>
                        </div>
                        <p class="note">首次登录后请立即修改密码，并配置 AI 模型的 API Key。</p>
                      </div>
                    </div>
                  </div>

                  <div class="notice-box">
                    <h4>📋 系统要求</h4>
                    <ul>
                      <li>Docker Engine 24+（含 Compose v2 插件）</li>
                      <li>4 GB 以上可用内存（Milvus 向量库约需 2 GB）</li>
                      <li>20 GB 可用磁盘空间</li>
                    </ul>
                  </div>
                </div>
              </ScrollReveal>
            </div>

            <!-- 宝塔面板部署 -->
            <div v-if="activeTab === 'bt'" class="tab-panel">
              <ScrollReveal>
                <div class="deploy-section">
                  <h3>🎯 宝塔面板部署（推荐）</h3>
                  <p class="deploy-desc">适合新手用户，可视化操作，一键部署</p>

                  <div class="steps">
                    <div class="step">
                      <div class="step-number">1</div>
                      <div class="step-content">
                        <h4>安装宝塔面板</h4>
                        <p>在服务器上执行以下命令安装宝塔面板：</p>
                        <CodeBlock :code="btInstallCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">2</div>
                      <div class="step-content">
                        <h4>安装必要软件</h4>
                        <p>在宝塔面板中安装以下软件：</p>
                        <ul>
                          <li>Nginx（用于反向代理）</li>
                          <li>PostgreSQL 16+（数据库）</li>
                          <li>Redis 7+（缓存）</li>
                          <li>PM2 管理器（进程管理）</li>
                        </ul>
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">3</div>
                      <div class="step-content">
                        <h4>上传项目代码</h4>
                        <p>将项目代码上传到服务器目录：</p>
                        <CodeBlock :code="uploadCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">4</div>
                      <div class="step-content">
                        <h4>配置环境变量</h4>
                        <p>创建环境配置文件：</p>
                        <CodeBlock :code="envCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">5</div>
                      <div class="step-content">
                        <h4>初始化数据库</h4>
                        <p>执行数据库迁移脚本：</p>
                        <CodeBlock :code="dbInitCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">6</div>
                      <div class="step-content">
                        <h4>编译并启动服务</h4>
                        <p>使用启动脚本一键启动所有服务：</p>
                        <CodeBlock :code="startCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">7</div>
                      <div class="step-content">
                        <h4>配置 Nginx 反向代理</h4>
                        <p>在宝塔面板中添加网站，配置反向代理到 Gateway 服务（端口 9100）</p>
                      </div>
                    </div>
                  </div>
                </div>
              </ScrollReveal>
            </div>

            <!-- Docker 部署 -->
            <div v-if="activeTab === 'docker'" class="tab-panel">
              <ScrollReveal>
                <div class="deploy-section">
                  <h3>🐳 Docker 部署</h3>
                  <p class="deploy-desc">适合有 Docker 经验的用户，容器化部署，便于迁移</p>

                  <div class="steps">
                    <div class="step">
                      <div class="step-number">1</div>
                      <div class="step-content">
                        <h4>安装 Docker 和 Docker Compose</h4>
                        <CodeBlock :code="dockerInstallCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">2</div>
                      <div class="step-content">
                        <h4>克隆项目代码</h4>
                        <CodeBlock :code="cloneCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">3</div>
                      <div class="step-content">
                        <h4>配置环境变量</h4>
                        <p>复制环境变量模板并修改配置：</p>
                        <CodeBlock :code="dockerEnvCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">4</div>
                      <div class="step-content">
                        <h4>启动所有服务</h4>
                        <CodeBlock :code="dockerUpCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">5</div>
                      <div class="step-content">
                        <h4>查看服务状态</h4>
                        <CodeBlock :code="dockerPsCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">6</div>
                      <div class="step-content">
                        <h4>停止服务</h4>
                        <CodeBlock :code="dockerDownCode" language="bash" />
                      </div>
                    </div>
                  </div>
                </div>
              </ScrollReveal>
            </div>

            <!-- 原生部署 -->
            <div v-if="activeTab === 'native'" class="tab-panel">
              <ScrollReveal>
                <div class="deploy-section">
                  <h3>🔧 原生部署</h3>
                  <p class="deploy-desc">适合高级用户，完全控制每个组件</p>

                  <div class="steps">
                    <div class="step">
                      <div class="step-number">1</div>
                      <div class="step-content">
                        <h4>安装依赖环境</h4>
                        <p>安装 Go、Python、Node.js、PostgreSQL、Redis</p>
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">2</div>
                      <div class="step-content">
                        <h4>克隆项目代码</h4>
                        <CodeBlock :code="cloneCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">3</div>
                      <div class="step-content">
                        <h4>编译后端服务</h4>
                        <CodeBlock :code="buildBackendCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">4</div>
                      <div class="step-content">
                        <h4>构建前端项目</h4>
                        <CodeBlock :code="buildFrontendCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">5</div>
                      <div class="step-content">
                        <h4>初始化数据库</h4>
                        <CodeBlock :code="dbInitCode" language="bash" />
                      </div>
                    </div>

                    <div class="step">
                      <div class="step-number">6</div>
                      <div class="step-content">
                        <h4>启动各个服务</h4>
                        <p>分别启动各个微服务：</p>
                        <CodeBlock :code="startServicesCode" language="bash" />
                      </div>
                    </div>
                  </div>
                </div>
              </ScrollReveal>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- Configuration -->
    <section class="section">
      <div class="container">
        <SectionTitle
          tag="配置说明"
          title="关键配置项"
          description="了解重要的配置项，确保系统正常运行"
          :center="true"
        />

        <div class="config-grid">
          <ScrollReveal>
            <div class="config-card">
              <h3>🔑 AI 模型配置</h3>
              <p>在管理后台配置 AI 模型的 API Key：</p>
              <ul>
                <li>通义千问 API Key</li>
                <li>DeepSeek API Key</li>
                <li>OpenAI API Key（可选）</li>
              </ul>
              <p class="config-note">API Key 使用 AES-256 加密存储，安全可靠</p>
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="config-card">
              <h3>🗄️ 数据库配置</h3>
              <p>PostgreSQL 数据库连接配置：</p>
              <CodeBlock :code="dbConfigCode" language="env" />
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="config-card">
              <h3>🔴 Redis 配置</h3>
              <p>Redis 缓存连接配置：</p>
              <CodeBlock :code="redisConfigCode" language="env" />
            </div>
          </ScrollReveal>

          <ScrollReveal>
            <div class="config-card">
              <h3>🌐 Nginx 配置</h3>
              <p>反向代理配置示例：</p>
              <CodeBlock :code="nginxConfigCode" language="nginx" />
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
          title="部署 FAQ"
          description="解答部署过程中的常见问题"
          :center="true"
        />

        <div class="faq-list">
          <ScrollReveal v-for="(faq, index) in faqs" :key="index">
            <div class="faq-item">
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
          <h2>需要帮助？</h2>
          <p>选择适合您的部署方式，如果遇到问题请联系我们的技术支持团队</p>
          <div class="cta-actions">
            <router-link to="/pricing" class="btn btn-primary btn-large">查看方案</router-link>
            <router-link to="/docs/guide" class="btn btn-secondary btn-large">使用指南</router-link>
          </div>
        </ScrollReveal>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useHead } from '@unhead/vue'
import SectionTitle from '@/components/SectionTitle.vue'
import ScrollReveal from '@/components/ScrollReveal.vue'
import CodeBlock from '@/components/CodeBlock.vue'

useHead({
  title: '部署教程 - 辰科 fettle | 私有部署与 SaaS',
  meta: [
    { name: 'description', content: '辰科 fettle 部署指南：私有部署版 Docker Compose 一键部署（端口 20001-20025），宝塔面板部署，Docker 集群部署，原生部署。' }
  ]
})

const activeTab = ref('standalone')

const tabs = [
  { id: 'standalone', label: '私有部署', icon: '🏠' },
  { id: 'bt', label: '宝塔面板', icon: '🎯' },
  { id: 'docker', label: 'Docker 集群', icon: '🐳' },
  { id: 'native', label: '原生部署', icon: '🔧' }
]

// 代码片段
const btInstallCode = `# CentOS 安装命令
yum install -y wget && wget -O install.sh https://download.bt.cn/install/install_6.0.sh && sh install.sh ed8484bec

# Ubuntu/Debian 安装命令
wget -O install.sh https://download.bt.cn/install/install-ubuntu_6.0.sh && sudo bash install.sh ed8484bec`

const uploadCode = `# 创建项目目录
mkdir -p /www/wwwroot/ai-platform
cd /www/wwwroot/ai-platform

# 上传代码（使用 scp 或 ftp）
scp -r ./backend/* root@your-server:/www/wwwroot/ai-platform/backend/
scp -r ./frontend/* root@your-server:/www/wwwroot/ai-platform/frontend/`

const envCode = `# 创建环境配置文件
cat > .env << 'EOF'
APP_ENV=production
JWT_SECRET=your-jwt-secret-key-change-this
JWT_EXPIRE_HOURS=72

# 数据库配置
DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=ai_platform
DB_PASSWORD=your-secure-password
DB_NAME=ai_platform
DB_SSL_MODE=disable

# Redis 配置
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=

# 服务地址
USER_SERVICE_ADDR=http://localhost:9200
AGENT_SERVICE_ADDR=http://localhost:9300
CHAT_SERVICE_ADDR=http://localhost:9400
SKILL_SERVICE_ADDR=http://localhost:9500
BILLING_SERVICE_ADDR=http://localhost:9600
AI_ENGINE_ADDR=http://localhost:9700
EOF`

const dbInitCode = `# 创建数据库
psql -U postgres -c "CREATE USER ai_platform WITH PASSWORD 'your-secure-password';"
psql -U postgres -c "CREATE DATABASE ai_platform OWNER ai_platform;"

# 执行迁移脚本
psql -U ai_platform -d ai_platform -f database/postgres/migrations/V001__init.sql
psql -U ai_platform -d ai_platform -f database/postgres/migrations/V002__ai_config.sql
psql -U ai_platform -d ai_platform -f database/postgres/migrations/V005__agent_knowledge_billing.sql`

const startCode = `# 赋予执行权限
chmod +x start_services.sh

# 启动所有服务
./start_services.sh

# 查看服务状态
ps aux | grep -E 'user-service|agent-service|chat-service|skill-service|billing-service|gateway'`

const dockerInstallCode = `# 安装 Docker
curl -fsSL https://get.docker.com | sh

# 安装 Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# 验证安装
docker --version
docker-compose --version`

const cloneCode = `# 克隆项目
git clone <your-repo-url> ai-platform
cd ai-platform`

const dockerEnvCode = `# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
vim .env

# 必须修改以下配置：
# - JWT_SECRET
# - DB_PASSWORD
# - AI 模型 API Key`

const dockerUpCode = `# 构建并启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 查看服务状态
docker-compose ps`

const dockerPsCode = `# 查看所有容器
docker-compose ps

# 应该看到以下服务运行中：
# - gateway
# - user-service
# - agent-service
# - chat-service
# - skill-service
# - billing-service
# - ai-engine
# - postgres
# - redis`

const standaloneCloneCode = `# 克隆私有部署仓库
git clone <your-repo-url> fettle-standalone
cd fettle-standalone

# 或直接下载部署包解压
wget <release-url>/fettle-standalone.tar.gz
tar -xzf fettle-standalone.tar.gz
cd fettle-standalone`

const standaloneEnvCode = `# 复制环境变量模板
cp .env.example .env

# 编辑环境变量（务必修改 JWT_SECRET）
vim .env

# 关键修改项：
# JWT_SECRET=替换为随机字符串
# 可选：配置 AI 模型 API Key（也可在后台配置）`

const standaloneUpCode = `# 一键启动所有服务
docker compose up -d

# 查看启动日志
docker compose logs -f

# 查看服务状态
docker compose ps

# 停止服务
docker compose down`

const dockerDownCode = `# 停止所有服务
docker-compose down

# 停止并删除数据卷（谨慎使用）
docker-compose down -v`

const buildBackendCode = `# 编译所有后端服务
cd backend

# 编译 Gateway
cd gateway && go build -o ../../build/gateway .
cd ..

# 编译 User Service
cd user-service && go build -o ../../build/user-service .
cd ..

# 编译 Agent Service
cd agent-service && go build -o ../../build/agent-service .
cd ..

# 编译 Chat Service
cd chat-service && go build -o ../../build/chat-service .
cd ..

# 编译 Skill Service
cd skill-service && go build -o ../../build/skill-service .
cd ..

# 编译 Billing Service
cd billing-service && go build -o ../../build/billing-service .
cd ..`

const buildFrontendCode = `# 构建管理后台
cd frontend/web-admin
npm install
npm run build

# 构建平台管理端
cd ../web-platform
npm install
npm run build

# 构建官网
cd ../web-website
npm install
npm run build`

const startServicesCode = `# 启动 Gateway
nohup ./build/gateway > logs/gateway.log 2>&1 &

# 启动 User Service
nohup ./build/user-service > logs/user-service.log 2>&1 &

# 启动 Agent Service
nohup ./build/agent-service > logs/agent-service.log 2>&1 &

# 启动 Chat Service
nohup ./build/chat-service > logs/chat-service.log 2>&1 &

# 启动 Skill Service
nohup ./build/skill-service > logs/skill-service.log 2>&1 &

# 启动 Billing Service
nohup ./build/billing-service > logs/billing-service.log 2>&1 &

# 启动 AI Engine
cd backend/ai-engine
nohup python main.py > ../../logs/ai-engine.log 2>&1 &`

const dbConfigCode = `DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=ai_platform
DB_PASSWORD=your-secure-password
DB_NAME=ai_platform
DB_SSL_MODE=disable`

const redisConfigCode = `REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0`

const nginxConfigCode = `server {
    listen 80;
    server_name your-domain.com;

    # 官网
    location / {
        root /www/wwwroot/ai-platform/frontend/web-website/dist;
        index index.html;
        try_files $uri $uri/ /index.html;
    }

    # API 网关
    location /api/ {
        proxy_pass http://127.0.0.1:9100/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # SSE 支持
        proxy_buffering off;
        proxy_cache off;
        proxy_read_timeout 300s;
    }

    # 管理后台
    location /admin {
        alias /www/wwwroot/ai-platform/frontend/web-admin/dist;
        index index.html;
        try_files $uri $uri/ /admin/index.html;
    }
}`

const faqs = [
  {
    question: 'Q: SaaS 云服务和私有部署版有什么区别？',
    answer: 'A: SaaS 版是多租户云服务，即开即用免运维；私有部署版（Standalone）是单租户 Docker 一键部署，数据完全存储在您的服务器上，适合对数据安全有严格要求的企业。'
  },
  {
    question: 'Q: 私有部署版最低服务器配置要求？',
    answer: 'A: 建议最低配置：2 核 CPU、4GB 内存、40GB SSD。推荐配置：4 核 CPU、8GB 内存、100GB SSD。Milvus 向量数据库需要约 2GB 内存。'
  },
  {
    question: 'Q: 私有部署版支持哪些操作系统？',
    answer: 'A: 只要支持 Docker Engine 24+ 的 Linux 发行版均可，包括 CentOS 7+、Ubuntu 18.04+、Debian 10+ 等。'
  },
  {
    question: 'Q: 私有部署版的默认账号密码是什么？',
    answer: 'A: 默认管理员账号 superadmin，密码 CHANGE_ME_ADMIN_PASSWORD。首次登录后建议立即修改密码。'
  },
  {
    question: 'Q: 私有部署版如何配置 AI 模型？',
    answer: 'A: 有两种方式：1）在 .env 文件中配置 QWEN_API_KEY / DEEPSEEK_API_KEY；2）登录后台后在「AI 模型配置」页面添加。API Key 使用 AES-256 加密存储。'
  },
  {
    question: 'Q: 如何备份和恢复数据？',
    answer: 'A: 备份 PostgreSQL 数据库（docker compose exec postgres pg_dump）和 Minio 存储桶数据。恢复时先重建数据库再导入备份文件。'
  },
  {
    question: 'Q: 私有部署版与多租户版功能上有什么差异？',
    answer: 'A: 私有部署版移除了租户管理、支付计费、企业注册等纯 SaaS 功能，保留了全部业务核心功能：智能体管理、对话、多渠道接入、知识库 RAG、AI 工作流等。'
  }
]
</script>

<style lang="scss" scoped>
.page-header {
  padding: 140px 0 60px;
  background: linear-gradient(180deg, $gray-50 0%, $bg-white 100%);
}

.requirements-card {
  background: $bg-white;
  border-radius: $radius-xl;
  padding: $spacing-2xl;
  box-shadow: $shadow-lg;
  border: 1px solid $border-light;

  h2 {
    font-size: 1.75rem;
    margin-bottom: $spacing-xl;
    color: $text-primary;
  }
}

.requirements-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: $spacing-lg;

  @include mobile {
    grid-template-columns: repeat(2, 1fr);
  }
}

.requirement-item {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-md;
  background: $gray-50;
  border-radius: $radius-md;

  .req-icon {
    font-size: 32px;
    flex-shrink: 0;
  }

  .req-info {
    strong {
      display: block;
      font-size: 1rem;
      color: $text-primary;
      margin-bottom: 2px;
    }

    span {
      font-size: 0.85rem;
      color: $text-muted;
    }
  }
}

.deploy-tabs {
  background: $bg-white;
  border-radius: $radius-xl;
  overflow: hidden;
  box-shadow: $shadow-lg;
  border: 1px solid $border-light;
}

.tab-nav {
  display: flex;
  background: $gray-50;
  border-bottom: 1px solid $border-light;

  @include mobile {
    flex-direction: column;
  }
}

.tab-btn {
  flex: 1;
  padding: $spacing-lg;
  background: none;
  border: none;
  font-size: 1rem;
  font-weight: 500;
  color: $text-secondary;
  cursor: pointer;
  transition: all $transition-fast;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;

  .tab-icon {
    font-size: 1.25rem;
  }

  &:hover {
    background: $gray-100;
    color: $text-primary;
  }

  &.active {
    background: $bg-white;
    color: $primary;
    box-shadow: 0 -2px 0 $primary inset;
  }
}

.tab-content {
  padding: $spacing-2xl;
}

.deploy-section {
  h3 {
    font-size: 1.75rem;
    margin-bottom: $spacing-sm;
    color: $text-primary;
  }

  .deploy-desc {
    font-size: 1rem;
    color: $text-secondary;
    margin-bottom: $spacing-2xl;
  }
}

.steps {
  display: flex;
  flex-direction: column;
  gap: $spacing-xl;
}

.step {
  display: flex;
  gap: $spacing-lg;

  .step-number {
    width: 40px;
    height: 40px;
    background: $primary;
    color: white;
    border-radius: 50%;
    display: flex;
    align-items: center;
    justify-content: center;
    font-weight: 700;
    font-size: 1.125rem;
    flex-shrink: 0;
  }

  .step-content {
    flex: 1;

    h4 {
      font-size: 1.25rem;
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
      }
    }

    .login-info {
      background: $gray-50;
      border-radius: $radius-md;
      padding: $spacing-md $spacing-lg;
      margin-bottom: $spacing-md;
      display: inline-block;

      p {
        margin: 4px 0;
        font-size: 0.95rem;
      }

      code {
        background: $gray-200;
        padding: 2px 8px;
        border-radius: 4px;
        font-size: 0.9rem;
      }
    }

    .note {
      font-size: 0.9rem;
      color: $text-muted;
      font-style: italic;
    }
  }
}

.notice-box {
  background: $gray-50;
  border: 1px solid $border-light;
  border-radius: $radius-lg;
  padding: $spacing-xl;
  margin-top: $spacing-2xl;

  h4 {
    font-size: 1.1rem;
    margin-bottom: $spacing-md;
    color: $text-primary;
  }

  ul {
    margin: 0;
    padding-left: $spacing-lg;
    color: $text-secondary;

    li {
      margin-bottom: $spacing-xs;
      font-size: 0.95rem;
    }
  }
}

.config-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-lg;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.config-card {
  background: $bg-white;
  border-radius: $radius-lg;
  padding: $spacing-xl;
  border: 1px solid $border-light;

  h3 {
    font-size: 1.25rem;
    margin-bottom: $spacing-md;
    color: $text-primary;
  }

  p {
    color: $text-secondary;
    margin-bottom: $spacing-md;
  }

  ul {
    margin-left: $spacing-lg;
    color: $text-secondary;
    margin-bottom: $spacing-md;

    li {
      margin-bottom: $spacing-xs;
    }
  }

  .config-note {
    font-size: 0.9rem;
    color: $text-muted;
    font-style: italic;
  }
}

.faq-list {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: $spacing-lg;

  @include mobile {
    grid-template-columns: 1fr;
  }
}

.faq-item {
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
