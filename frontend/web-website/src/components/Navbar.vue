<template>
  <nav class="navbar" :class="{ scrolled: isScrolled }">
    <div class="container">
      <div class="navbar-content">
        <router-link to="/" class="logo">
          <div class="logo-icon">
            <svg viewBox="0 0 32 32" fill="none">
              <rect width="32" height="32" rx="6" fill="#3b82f6"/>
              <path d="M16 8L22 12V20L16 24L10 20V12L16 8Z" fill="white" stroke="white" stroke-width="1.5"/>
              <circle cx="16" cy="16" r="3" fill="#3b82f6"/>
            </svg>
          </div>
          <span class="logo-text">AI 智能体平台</span>
        </router-link>

        <div class="nav-links" :class="{ active: mobileMenuOpen }">
          <router-link to="/" @click="closeMobile">首页</router-link>
          <router-link to="/features" @click="closeMobile">功能</router-link>
          <router-link to="/advantages" @click="closeMobile">特点</router-link>
          <router-link to="/pricing" @click="closeMobile">定价</router-link>
          <div class="dropdown">
            <span class="dropdown-trigger">文档</span>
            <div class="dropdown-menu">
              <router-link to="/docs/deploy" @click="closeMobile">部署教程</router-link>
              <router-link to="/docs/guide" @click="closeMobile">使用指南</router-link>
            </div>
          </div>
        </div>

        <div class="nav-actions">
          <a href="#" class="btn btn-ghost btn-small">登录</a>
          <a href="#" class="btn btn-primary btn-small">免费试用</a>
        </div>

        <button class="mobile-toggle" @click="toggleMobile" :class="{ active: mobileMenuOpen }">
          <span></span>
          <span></span>
          <span></span>
        </button>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const isScrolled = ref(false)
const mobileMenuOpen = ref(false)

const handleScroll = () => {
  isScrolled.value = window.scrollY > 20
}

const toggleMobile = () => {
  mobileMenuOpen.value = !mobileMenuOpen.value
  document.body.style.overflow = mobileMenuOpen.value ? 'hidden' : ''
}

const closeMobile = () => {
  mobileMenuOpen.value = false
  document.body.style.overflow = ''
}

onMounted(() => {
  window.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll)
})
</script>

<style lang="scss" scoped>
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: $z-fixed;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  transition: all $transition-normal;

  &.scrolled {
    box-shadow: $shadow-sm;
    background: rgba(255, 255, 255, 0.98);
  }
}

.navbar-content {
  @include flex-between;
  height: 72px;
}

.logo {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  text-decoration: none;
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
  color: $text-primary;
}

.nav-links {
  display: flex;
  align-items: center;
  gap: $spacing-xl;

  a {
    color: $text-secondary;
    font-size: 15px;
    font-weight: 500;
    transition: color $transition-fast;

    &:hover,
    &.router-link-active {
      color: $primary;
    }
  }
}

.dropdown {
  position: relative;

  .dropdown-trigger {
    color: $text-secondary;
    font-size: 15px;
    font-weight: 500;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 4px;

    &::after {
      content: '';
      width: 0;
      height: 0;
      border-left: 4px solid transparent;
      border-right: 4px solid transparent;
      border-top: 4px solid currentColor;
    }

    &:hover {
      color: $primary;
    }
  }

  .dropdown-menu {
    position: absolute;
    top: 100%;
    left: 50%;
    transform: translateX(-50%) translateY(10px);
    background: $bg-white;
    border-radius: $radius-md;
    box-shadow: $shadow-lg;
    padding: $spacing-sm;
    min-width: 160px;
    opacity: 0;
    visibility: hidden;
    transition: all $transition-fast;

    a {
      display: block;
      padding: $spacing-sm $spacing-md;
      border-radius: $radius-sm;
      white-space: nowrap;

      &:hover {
        background: $gray-100;
      }
    }
  }

  &:hover .dropdown-menu {
    opacity: 1;
    visibility: visible;
    transform: translateX(-50%) translateY(0);
  }
}

.nav-actions {
  display: flex;
  align-items: center;
  gap: $spacing-md;
}

.mobile-toggle {
  display: none;
  flex-direction: column;
  justify-content: center;
  gap: 5px;
  width: 28px;
  height: 28px;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;

  span {
    display: block;
    width: 100%;
    height: 2px;
    background: $text-primary;
    border-radius: 2px;
    transition: all $transition-fast;
  }

  &.active {
    span:nth-child(1) {
      transform: translateY(7px) rotate(45deg);
    }

    span:nth-child(2) {
      opacity: 0;
    }

    span:nth-child(3) {
      transform: translateY(-7px) rotate(-45deg);
    }
  }
}

@include mobile {
  .nav-links {
    position: fixed;
    top: 72px;
    left: 0;
    right: 0;
    bottom: 0;
    background: $bg-white;
    flex-direction: column;
    align-items: stretch;
    padding: $spacing-xl;
    gap: $spacing-lg;
    transform: translateX(100%);
    transition: transform $transition-normal;

    &.active {
      transform: translateX(0);
    }

    a {
      font-size: 18px;
      padding: $spacing-md 0;
      border-bottom: 1px solid $border-light;
    }
  }

  .dropdown {
    .dropdown-menu {
      position: static;
      transform: none;
      box-shadow: none;
      padding-left: $spacing-md;
      opacity: 1;
      visibility: visible;
    }
  }

  .nav-actions {
    display: none;
  }

  .mobile-toggle {
    display: flex;
  }
}
</style>
