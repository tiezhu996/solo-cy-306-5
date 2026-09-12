import request from '@/utils/request'
import type { User } from '@/types'

export function register(data: { username: string; password: string; nickname?: string; email?: string; phone?: string; role?: string }) {
  return request.post('/auth/register', data)
}

export function login(data: { username: string; password: string }) {
  return request.post('/auth/login', data)
}

export function getMe() {
  return request.get('/users/me')
}

export function updateProfile(data: Partial<User>) {
  return request.put('/users/me', data)
}

export function listUsers(params: { page?: number; page_size?: number }) {
  return request.get('/users', { params })
}
