import { createRouter, createWebHistory } from 'vue-router'
import { setupGuards } from './guards'
import Layout from '@/pages/Layout.vue'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: Layout,
      children: [
        { path: '', redirect: '/calendar' },
        { path: 'calendar', name: 'calendar', component: () => import('@/pages/Calendar.vue') },
        { path: 'activities', name: 'activities', component: () => import('@/pages/Activities.vue') },
        { path: 'activities/:id', name: 'activity-detail', component: () => import('@/pages/ActivityDetail.vue') },
        {
          path: 'organizer/activities',
          name: 'organizer-activities',
          component: () => import('@/pages/OrganizerActivities.vue'),
          meta: { requiresAuth: true, roles: ['organizer', 'admin'] },
        },
        {
          path: 'organizer/registrations',
          name: 'organizer-registrations',
          component: () => import('@/pages/OrganizerRegistrations.vue'),
          meta: { requiresAuth: true, roles: ['organizer', 'admin'] },
        },
        { path: 'profile', name: 'profile', component: () => import('@/pages/Profile.vue'), meta: { requiresAuth: true } },
      ],
    },
    { path: '/login', name: 'login', component: () => import('@/pages/Login.vue') },
    { path: '/register', name: 'register', component: () => import('@/pages/Register.vue') },
  ],
})

setupGuards(router)
export default router
