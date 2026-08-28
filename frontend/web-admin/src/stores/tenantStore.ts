import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { TenantInfo } from '@/api/tenant'

export const useTenantStore = defineStore('tenant', () => {
  const currentTenant = ref<TenantInfo | null>(null)
  const tenantList = ref<TenantInfo[]>([])

  const tenantId = computed(() => currentTenant.value?.id || '')
  const tenantName = computed(() => currentTenant.value?.name || '')
  const planType = computed(() => currentTenant.value?.plan_type || 'free')
  const planExpiresAt = computed(() => currentTenant.value?.plan_expires_at || '')

  function setCurrentTenant(tenant: TenantInfo) {
    currentTenant.value = tenant
    localStorage.setItem('current_tenant', JSON.stringify(tenant))
  }

  function clearCurrentTenant() {
    currentTenant.value = null
    localStorage.removeItem('current_tenant')
  }

  function setTenantList(list: TenantInfo[]) {
    tenantList.value = list
  }

  function loadFromStorage() {
    const stored = localStorage.getItem('current_tenant')
    if (stored) {
      try {
        currentTenant.value = JSON.parse(stored)
      } catch (error) {
        console.error('Failed to parse tenant from storage', error)
      }
    }
  }

  return {
    currentTenant,
    tenantList,
    tenantId,
    tenantName,
    planType,
    planExpiresAt,
    setCurrentTenant,
    clearCurrentTenant,
    setTenantList,
    loadFromStorage
  }
})
