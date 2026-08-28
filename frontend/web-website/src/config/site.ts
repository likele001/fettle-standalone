export const SITE_URLS = {
  login: import.meta.env.VITE_LOGIN_URL || 'https://fettle.cenkor.cn/login',
  register: import.meta.env.VITE_REGISTER_URL || 'https://fettle.cenkor.cn/register',
  api: import.meta.env.VITE_API_BASE_URL || 'https://fettle.cenkor.cn',
} as const
