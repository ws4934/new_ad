import { createRouter, createWebHashHistory } from 'vue-router'
import { isLoggedIn } from './api.js?v=12'

const routes = [
  { path: '/login', component: () => import('./pages/Login.js?v=12'), meta: { public: true } },
  { path: '/', component: () => import('./pages/Dashboard.js?v=12') },
  { path: '/platforms', component: () => import('./pages/Platforms.js?v=12') },
  { path: '/accounts', component: () => import('./pages/Accounts.js?v=12') },
  { path: '/accounts/:id', component: () => import('./pages/AccountDetail.js?v=12'), props: true },
  { path: '/stats', component: () => import('./pages/Stats.js?v=12') },
  { path: '/export', component: () => import('./pages/Export.js?v=12') },
  { path: '/issue', component: () => import('./pages/IssueStatus.js?v=12') },
  { path: '/issue-report', component: () => import('./pages/IssueReport.js?v=12') },
  { path: '/issue-config', component: () => import('./pages/IssueConfig.js?v=12') },
  { path: '/system', component: () => import('./pages/System.js?v=12') },
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
