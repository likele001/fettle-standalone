import request from './request'

export function getAdminSkills(params?: { page?: number; page_size?: number }): Promise<any> {
  return request.get('/admin/skills', { params })
}

export function getAdminSkillStats(): Promise<any> {
  return request.get('/admin/skills/stats')
}
