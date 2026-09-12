import request from '@/utils/request'

export function signup(data: { activity_id: number; name: string; phone: string; remark?: string }) {
  return request.post('/registrations', data)
}

export function listRegistrations(params: { page?: number; page_size?: number; activity_id?: number; status?: string; review_status?: string }) {
  return request.get('/registrations', { params })
}

export function listMyRegistrations(params: { page?: number; page_size?: number }) {
  return request.get('/registrations/mine', { params })
}

export function cancelRegistration(id: number) {
  return request.post(`/registrations/${id}/cancel`)
}

export function reviewRegistration(id: number, reviewStatus: string) {
  return request.post(`/registrations/${id}/review`, { review_status: reviewStatus })
}

export function offlineSignup(data: { activity_id: number; name: string; phone: string; remark?: string }) {
  return request.post('/registrations/offline', data)
}

export function exportRegistrations(activityId: number) {
  return request.get('/registrations/export', { params: { activity_id: activityId }, responseType: 'blob' })
}
