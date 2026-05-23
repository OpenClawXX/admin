import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/login/index.vue'),
    meta: { requiresAuth: false },
  },
  {
    path: '/',
    component: () => import('@/components/Layout.vue'),
    redirect: '/dashboard',
    children: [
      {
        path: 'dashboard',
        name: 'Dashboard',
        component: () => import('@/views/dashboard/index.vue'),
        meta: { title: '工作台', icon: 'icon-dashboard' },
      },
      {
        path: 'tickets',
        name: 'Tickets',
        component: () => import('@/views/ticket/index.vue'),
        meta: { title: '工单管理', icon: 'icon-file' },
      },
      {
        path: 'tickets/:id',
        name: 'TicketDetail',
        component: () => import('@/views/ticket/detail.vue'),
        meta: { title: '工单详情', hidden: true },
      },
      {
        path: 'projects',
        name: 'Projects',
        component: () => import('@/views/project/index.vue'),
        meta: { title: '项目管理', icon: 'icon-folder', roles: ['admin', 'supervisor'] },
      },
      {
        path: 'engineers',
        name: 'Engineers',
        component: () => import('@/views/engineer/index.vue'),
        meta: { title: '工程师管理', icon: 'icon-user', roles: ['admin', 'supervisor'] },
      },
      {
        path: 'teams',
        name: 'Teams',
        component: () => import('@/views/team/index.vue'),
        meta: { title: '团队管理', icon: 'icon-user-group', roles: ['admin', 'supervisor'] },
      },
      {
        path: 'knowledge',
        name: 'Knowledge',
        component: () => import('@/views/knowledge/index.vue'),
        meta: { title: '知识库', icon: 'icon-book' },
      },
      {
        path: 'schedule',
        name: 'Schedule',
        component: () => import('@/views/schedule/index.vue'),
        meta: { title: '排班管理', icon: 'icon-calendar', roles: ['admin', 'supervisor'] },
      },
      {
        path: 'assets',
        name: 'Assets',
        component: () => import('@/views/asset/index.vue'),
        meta: { title: '资产管理', icon: 'icon-desktop', roles: ['admin', 'supervisor'] },
      },
      {
        path: 'system',
        name: 'System',
        component: () => import('@/views/system/index.vue'),
        meta: { title: '系统设置', icon: 'icon-settings', roles: ['admin'] },
      },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth !== false && !token) {
    next('/login')
  } else {
    next()
  }
})

export default router
