import request from './request'

export function exportConversations(params?: { start_date?: string; end_date?: string }) {
  return request.get('/chats/export/conversations', { params, responseType: 'blob' } as any)
}

export function exportMessages(params?: { start_date?: string; end_date?: string }) {
  return request.get('/chats/export/messages', { params, responseType: 'blob' } as any)
}

export function exportAnalytics(params?: { days?: number }) {
  return request.get('/chats/export/analytics', { params, responseType: 'blob' } as any)
}

export function exportBillingRecords(params?: { start_date?: string; end_date?: string }) {
  return request.get('/billing/export/records', { params, responseType: 'blob' } as any)
}

export function exportUsageSummary(params?: { months?: number }) {
  return request.get('/billing/export/usage', { params, responseType: 'blob' } as any)
}

// Helper to download blob as file
export function downloadBlob(blob: Blob, filename: string) {
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', filename)
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}
