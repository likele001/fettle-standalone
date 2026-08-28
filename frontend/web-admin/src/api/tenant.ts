export interface TenantInfo {
  id: string
  name: string
  code?: string | null
  plan_type: string
  plan_expires_at?: string
  status?: string
  config?: Record<string, any>
  created_at?: string
  updated_at?: string
}
