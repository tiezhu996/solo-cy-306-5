import type { ActivityStatusType, ActivityTypeType } from '@/constants/activity'


export interface User {
  id: number
  username: string
  nickname: string
  avatar: string
  role: string
  email: string
  phone: string
  created_at: string
}

export interface Activity {
  id: number
  title: string
  description: string
  cover_image: string
  activity_type: ActivityTypeType
  start_time: string
  end_time: string
  location: string
  capacity: number
  signup_deadline: string
  status: ActivityStatusType
  organizer_id: number
  created_at: string
  registered_count?: number
}

export interface Registration {
  id: number
  activity_id: number
  user_id: number
  name: string
  phone: string
  remark: string
  voucher_no: string
  status: string
  review_status: string
  created_at: string
}

export interface CheckInRecord {
  id: number
  registration_id: number
  activity_id: number
  check_in_method: string
  check_in_time: string
  operator_id: number
}

export interface CommentItem {
  id: number
  activity_id: number
  user_id: number
  rating: number
  content: string
  created_at: string
}

export interface Favorite {
  id: number
  user_id: number
  activity_id: number
  created_at: string
}

export interface NotificationItem {
  id: number
  user_id: number
  notification_type: string
  title: string
  content: string
  is_read: boolean
  created_at: string
  type_text?: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}
