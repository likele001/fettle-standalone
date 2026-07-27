import request from './request'

export interface TeamMember {
  id: string
  tenant_id: string
  user_id: string
  role: string
  status: string
  user_name: string
  user_phone: string
  user_email: string
  avatar_url: string
  joined_at: string
  resigned_at: string | null
}

export interface Invitation {
  id: string
  code: string
  invitee_phone: string
  invitee_email: string
  role: string
  used: boolean
  created_at: string
}

export function getTeamMembers(params?: { status?: string; page?: number; page_size?: number }) {
  return request.get('/team/members', { params })
}

export function createInvitation(data: { invitee_phone?: string; invitee_email?: string; role: string }) {
  return request.post('/team/invite', data)
}

export function getInvitations() {
  return request.get('/team/invitations')
}

export function cancelInvitation(id: string) {
  return request.delete(`/team/invitations/${id}`)
}

export function updateMemberRole(id: string, role: string) {
  return request.put(`/team/members/${id}/role`, { role })
}

export function resignMember(id: string) {
  return request.put(`/team/members/${id}/resign`)
}

export function acceptInvitation(code: string) {
  return request.post('/team/accept-invite', { code })
}

export function getInvitationInfo(code: string) {
  return request.get('/public/invitation', { params: { code } })
}

// 兼容 settings 页面的旧命名
export const listTeamMembers = getTeamMembers
export const inviteMember = createInvitation
export const removeMember = resignMember
export const updateProfile = (data: any) => request.put('/users/me', data)
