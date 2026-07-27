<template>
  <div class="code-block">
    <div class="code-header">
      <span class="language">{{ language }}</span>
      <button class="copy-btn" @click="copyCode" :class="{ copied }">
        <svg v-if="!copied" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect>
          <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path>
        </svg>
        <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="20 6 9 17 4 12"></polyline>
        </svg>
        {{ copied ? '已复制' : '复制' }}
      </button>
    </div>
    <pre><code>{{ code }}</code></pre>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(defineProps<{
  code: string
  language?: string
}>(), {
  language: 'bash'
})

const copied = ref(false)

async function copyCode() {
  try {
    await navigator.clipboard.writeText(props.code)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('复制失败:', err)
  }
}
</script>

<style lang="scss" scoped>
.code-block {
  background: $gray-900;
  border-radius: $radius-lg;
  overflow: hidden;
  margin: $spacing-lg 0;

  .code-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: $spacing-sm $spacing-md;
    background: $gray-800;
    border-bottom: 1px solid $gray-700;

    .language {
      font-size: 12px;
      color: $gray-400;
      text-transform: uppercase;
      letter-spacing: 0.5px;
    }

    .copy-btn {
      display: flex;
      align-items: center;
      gap: 6px;
      padding: 4px 12px;
      background: transparent;
      border: 1px solid $gray-600;
      border-radius: $radius-sm;
      color: $gray-300;
      font-size: 12px;
      cursor: pointer;
      transition: all $transition-fast;

      svg {
        width: 14px;
        height: 14px;
      }

      &:hover {
        background: $gray-700;
        border-color: $gray-500;
      }

      &.copied {
        color: $success;
        border-color: $success;
      }
    }
  }

  pre {
    margin: 0;
    padding: $spacing-lg;
    overflow-x: auto;

    code {
      font-family: $font-mono;
      font-size: 14px;
      line-height: 1.6;
      color: $gray-100;
      background: none;
      padding: 0;
    }
  }
}
</style>
