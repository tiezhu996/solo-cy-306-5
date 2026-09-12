import request from '@/utils/request'

export function listMyNotifications(params: { page?: number; page_size?: number }) {
  return request.get('/notifications/mine', { params })
}

export function markRead(id: number) {
  return request.post(`/notifications/${id}/read`)
}

export function markAllRead() {
  return request.post('/notifications/read-all')
}
