import { createRouter, createWebHashHistory } from 'vue-router'
import { isLoggedIn } from './api.js'

const routes = [
  { path: '/login', component: () => import('./pages/Login.js?v=3'), meta: { public: true } },
  { path: '/', component: () => import('./pages/Dashboard.js?v=3') },
  { path: '/platforms', component: () => import('./pages/Platforms.js?v=3') },
  { path: '/accounts', component: () => import('./pages/Accounts.js?v=3') },
  { path: '/accounts/:id', component: () => import('./pages/AccountDetail.js?v=3'), props: true },
  { path: '/stats', component: () => import('./pages/Stats.js?v=3') },
  { path: '/export', component: () => import('./pages/Export.js?v=3') },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

// 路由守卫: 未登录时重定向到 /login
router.beforeEach((to) => {
  if (!to.meta.public && !isLoggedIn()) {
    return '/login'
  }
})
