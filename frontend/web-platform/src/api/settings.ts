import request from './request'

export interface SmsConfig {
  access_key: string
  secret_key: string
  sign_name: string
  template_id: string
}

export interface EmailConfig {
  host: string
  port: number
  from_email: string
  from_name: string
  username: string
  password: string
}

export interface PlatformSettings {
  site_name: string
  site_logo: string
  contact_email: string
  sms_provider: string
  sms_config: SmsConfig
  email_provider: string
  email_config: EmailConfig
  max_upload_size: number
  allowed_file_types: string[]
  maintenance_mode: boolean
  wechat_miniapp_id: string
  wechat_miniapp_secret: string
}

export function getSettings(): Promise<PlatformSettings> {
  return request.get('/admin/settings') as any
}

export function updateSettings(data: Partial<PlatformSettings>): Promise<PlatformSettings> {
  return request.put('/admin/settings', data) as any
}

export function testSmsConfig(phone: string): Promise<{ success: boolean; message: string }> {
  return request.post('/admin/settings/test-sms', { phone }) as any
}

export function testEmailConfig(email: string): Promise<{ success: boolean; message: string }> {
  return request.post('/admin/settings/test-email', { email }) as any
}
