import request from './request'

export interface Skill {
  id: string
  name: string
  description: string
  type: string
  icon: string
  version: string
  author: string
  status: string
  created_at: string
  category?: string
  install_count?: number
}

export interface SkillInstallation {
  id: string
  skill_id: string
  tenant_id: string
  status: string
  config: Record<string, any>
  created_at: string
}

export function listSkills(params?: { page?: number; page_size?: number; type?: string }): Promise<{ items: Skill[]; total: number }> {
  return request.get('/skills', { params })
}

export function getSkill(id: string): Promise<Skill> {
  return request.get(`/skills/${id}`)
}

export function createSkill(data: Partial<Skill>): Promise<Skill> {
  return request.post('/skills', data)
}

export function updateSkill(id: string, data: Partial<Skill>): Promise<Skill> {
  return request.put(`/skills/${id}`, data)
}

export function deleteSkill(id: string): Promise<void> {
  return request.delete(`/skills/${id}`)
}

export function installSkill(skillId: string): Promise<SkillInstallation> {
  return request.post(`/skills/${skillId}/install`)
}

export function uninstallSkill(skillId: string): Promise<void> {
  return request.post(`/skills/${skillId}/uninstall`)
}

export function toggleSkill(skillId: string): Promise<SkillInstallation> {
  return request.post(`/skills/${skillId}/toggle`)
}

export function getInstalledSkills(): Promise<{ items: SkillInstallation[]; total: number }> {
  return request.get('/skills/installed')
}

export function getSkillConfig(skillId: string): Promise<{ config: string }> {
  return request.get(`/skills/${skillId}/config`)
}

export function updateSkillConfig(skillId: string, config: string): Promise<void> {
  return request.put(`/skills/${skillId}/config`, { config })
}
