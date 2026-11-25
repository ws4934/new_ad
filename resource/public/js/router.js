import { createRouter, createWebHashHistory } from 'vue-router'
import { isLoggedIn } from './api.js?v=20'

const routes = [
  { path: '/login', component: () => import('./pages/Login.js?v=20'), meta: { public: true } },
  { path: '/', component: () => import('./pages/Dashboard.js?v=20') },
  { path: '/platforms', component: () => import('./pages/Platforms.js?v=20') },
  { path: '/accounts', component: () => import('./pages/Accounts.js?v=20') },
  { path: '/accounts/:id', component: () => import('./pages/AccountDetail.js?v=20'), props: true },
  { path: '/stats', component: () => import('./pages/Stats.js?v=20') },
  { path: '/export', component: () => import('./pages/Export.js?v=20') },
  { path: '/issue', component: () => import('./pages/IssueStatus.js?v=20') },
  { path: '/issue-report', component: () => import('./pages/IssueReport.js?v=20') },
  { path: '/issue-config', component: () => import('./pages/IssueConfig.js?v=20') },
  { path: '/system', component: () => import('./pages/System.js?v=20') },
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
