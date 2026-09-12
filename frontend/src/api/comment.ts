import request from '@/utils/request'

export function listComments(activityId: number) {
  return request.get(`/activities/${activityId}/comments`)
}

export function createComment(activityId: number, data: { rating: number; content: string }) {
  return request.post(`/activities/${activityId}/comments`, data)
}

export function listMyComments() {
  return request.get('/comments/mine')
}

export function deleteComment(id: number) {
  return request.delete(`/comments/${id}`)
}
