import request from './request'

export interface IndustryBundle {
  id: string
  industry: string
  name: string
  description: string
  icon: string
  config: {
    agents: Array<{
      name: string
      description: string
      agent_type: string
      welcome_message: string
      personality_config: Record<string, any>
      required_plan: string
      kb_template_names: string[]
    }>
    kb_templates: Array<{
      name: string
      description: string
    }>
  }
  is_active: boolean
  created_at: string
}

export interface ApplyBundleResult {
  created_agents: Array<{ id: string; name: string }>
  created_kbs: Array<{ id: string; name: string }>
  skipped: string[]
}

export function getBundles() {
  return request.get('/agents/bundles')
}

export function getBundle(industry: string) {
  return request.get(`/agents/bundles/${industry}`)
}

export function applyBundle(industry: string) {
  return request.post('/agents/bundles/apply', { industry })
}
