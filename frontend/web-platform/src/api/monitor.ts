import request from './request'

export interface ServiceHealth {
  name: string
  status: 'healthy' | 'degraded' | 'down'
  uptime: string
  cpu_usage: number
  memory_usage: number
  response_time: number
  last_check: string
}

export interface SystemStats {
  total_services: number
  healthy_services: number
  total_requests_24h: number
  avg_response_time: number
  error_rate: number
}

export interface ApiCallStats {
  service: string
  calls_24h: number
  success_rate: number
  avg_latency: number
}

export interface SystemHealthResponse {
  services: ServiceHealth[]
}

export function getSystemHealth(): Promise<SystemHealthResponse> {
  return request.get('/admin/monitor/health') as any
}

export function getSystemStats(): Promise<SystemStats> {
  return request.get('/admin/monitor/stats') as any
}

export function getApiCallStats(): Promise<ApiCallStats[]> {
  return request.get('/admin/monitor/api-stats') as any
}

export function restartService(serviceName: string): Promise<void> {
  return request.post(`/admin/monitor/restart/${serviceName}`) as any
}
