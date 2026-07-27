import request from './request'

export interface BrandingConfig {
  brand_name: string
  brand_logo_url: string
  brand_favicon_url: string
  brand_primary_color: string
  brand_secondary_color: string
  brand_footer_text: string
  white_label_enabled: boolean
}

export function getBranding() {
  return request.get('/branding')
}

export function updateBranding(data: Partial<BrandingConfig>) {
  return request.put('/branding', data)
}

export function getPublicBranding(tenantId: string) {
  return request.get(`/public/branding/${tenantId}`)
}
