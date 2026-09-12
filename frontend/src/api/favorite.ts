import request from '@/utils/request'

export function addFavorite(activityId: number) {
  return request.post(`/activities/${activityId}/favorite`)
}

export function removeFavorite(activityId: number) {
  return request.delete(`/activities/${activityId}/favorite`)
}

export function listMyFavorites(params: { page?: number; page_size?: number }) {
  return request.get('/favorites/mine', { params })
}
