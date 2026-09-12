import request from '@/utils/request'
import type { Activity } from '@/types'

export function listActivities(params: { page?: number; page_size?: number; activity_type?: string; status?: string; keyword?: string }) {
  return request.get('/activities', { params })
}

export function getActivity(id: number) {
  return request.get(`/activities/${id}`)
}

export function getCalendar(month?: string) {
  return request.get('/activities/calendar', { params: { month } })
}

export function getMyActivities(params: { page?: number; page_size?: number; status?: string }) {
  return request.get('/activities/mine', { params })
}

export function createActivity(data: Partial<Activity>) {
  return request.post('/activities', data)
}

export function updateActivity(id: number, data: Partial<Activity>) {
  return request.put(`/activities/${id}`, data)
}

export function publishActivity(id: number) {
  return request.post(`/activities/${id}/publish`)
}

export function endActivity(id: number) {
  return request.post(`/activities/${id}/end`)
}

export function deleteActivity(id: number) {
  return request.delete(`/activities/${id}`)
}

export function getActivityStats(id: number) {
  return request.get(`/activities/${id}/stats`)
}
