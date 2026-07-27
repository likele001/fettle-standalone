<template>
  <footer class="footer">
    <div class="container">
      <div class="footer-content">
        <div class="footer-brand">
          <div class="logo">
            <div class="logo-icon">
              <svg viewBox="0 0 32 32" fill="none">
                <rect width="32" height="32" rx="6" fill="#3b82f6"/>
                <path d="M16 8L22 12V20L16 24L10 20V12L16 8Z" fill="white" stroke="white" stroke-width="1.5"/>
                <circle cx="16" cy="16" r="3" fill="#3b82f6"/>
              </svg>
            </div>
            <span class="logo-text">AI 智能体平台</span>
          </div>
          <p class="footer-desc">企业级智能对话解决方案，多租户、多模型、多渠道</p>
        </div>

        <div class="footer-links">
          <div class="link-group">
            <h4>产品</h4>
            <ul>
              <li><router-link to="/features">功能介绍</router-link></li>
              <li><router-link to="/advantages">技术特点</router-link></li>
              <li><router-link to="/pricing">价格方案</router-link></li>
            </ul>
          </div>

          <div class="link-group">
            <h4>文档</h4>
            <ul>
              <li><router-link to="/docs/deploy">部署教程</router-link></li>
              <li><router-link to="/docs/guide">使用指南</router-link></li>
              <li><a href="#">API 文档</a></li>
            </ul>
          </div>

          <div class="link-group">
            <h4>支持</h4>
            <ul>
              <li><a href="#">帮助中心</a></li>
              <li><a href="#">联系我们</a></li>
              <li><a href="#">问题反馈</a></li>
            </ul>
          </div>
        </div>
      </div>

      <div class="footer-bottom">
        <p>&copy; 2026 AI 智能体平台. 保留所有权利.</p>
      </div>
    </div>
  </footer>
</template>

<style lang="scss" scoped>
.footer {
  background: $gray-900;
  color: $gray-300;
  padding: $spacing-3xl 0 $spacing-xl;
  margin-top: $spacing-4xl;
}

.footer-content {
  display: grid;
  grid-template-columns: 1fr 2fr;
  gap: $spacing-3xl;
  margin-bottom: $spacing-2xl;

  @include mobile {
    grid-template-columns: 1fr;
    gap: $spacing-2xl;
  }
}

.footer-brand {
  .logo {
    display: flex;
    align-items: center;
    gap: $spacing-sm;
    margin-bottom: $spacing-md;
  }

  .logo-icon {
    width: 36px;
    height: 36px;

    svg {
      width: 100%;
      height: 100%;
    }
  }

  .logo-text {
    font-family: $font-display;
    font-size: 18px;
    font-weight: 600;
    color: $text-white;
  }

  .footer-desc {
    color: $gray-400;
    font-size: 14px;
    line-height: 1.6;
  }
}

.footer-links {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: $spacing-2xl;

  @include mobile {
    grid-template-columns: repeat(2, 1fr);
    gap: $spacing-xl;
  }
}

.link-group {
  h4 {
    color: $text-white;
    font-size: 14px;
    font-weight: 600;
    margin-bottom: $spacing-md;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }

  ul {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  li {
    margin-bottom: $spacing-sm;
  }

  a {
    color: $gray-400;
    font-size: 14px;
    transition: color $transition-fast;

    &:hover {
      color: $text-white;
    }
  }
}

.footer-bottom {
  padding-top: $spacing-xl;
  border-top: 1px solid $gray-800;
  text-align: center;

  p {
    color: $gray-500;
    font-size: 13px;
    margin: 0;
  }
}
</style>
