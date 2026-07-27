import { get, post, del } from '@/utils/request'

export interface Skill {
  id: string
  name: string
  description: string
  icon: string
  category: string
  version: string
  author: string
  is_installed: boolean
  install_count: number
  created_at: string
  updated_at: string
}

export interface SkillListParams {
  page?: number
  page_size?: number
  keyword?: string
  category?: string
}

export interface SkillListResponse {
  items: Skill[]
  total: number
  page: number
  page_size: number
}

export function getSkills(params?: SkillListParams) {
  return get<SkillListResponse>('/skills', params)
}

export function getSkillDetail(id: string) {
  return get<Skill>(`/skills/${id}`)
}

export function installSkill(skillId: string, agentId?: string) {
  return post<Skill>('/skills/install', { skill_id: skillId, agent_id: agentId })
}

export function uninstallSkill(skillId: string) {
  return del(`/skills/uninstall/${skillId}`)
}

export function getInstalledSkills() {
  return get<SkillListResponse>('/skills/installed')
}
