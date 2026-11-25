// fetch 封装, 对齐后端实际路由
// GoFrame 标准响应: { code: 0, message: "OK", data: {...} }

// token 管理
export function getToken() { return localStorage.getItem('ad_token') || '' }
export function setToken(t) { localStorage.setItem('ad_token', t) }
export function clearToken() { localStorage.removeItem('ad_token') }
export function isLoggedIn() { return !!getToken() }

async function request(path, { method = 'GET', body, query } = {}) {
  let url = path
  if (query) {
    const qs = new URLSearchParams(
      Object.entries(query).filter(([, v]) => v != null && v !== '')
    ).toString()
    if (qs) url += '?' + qs
  }
  const opts = { method, headers: {} }

  // 自动带 token (登录接口不需要, 但带了也不影响)
  const token = getToken()
  if (token) {
    opts.headers['Authorization'] = 'Bearer ' + token
  }

  if (body) {
    opts.headers['Content-Type'] = 'application/json'
    opts.body = JSON.stringify(body)
  }
  const resp = await fetch(url, opts)

  // 401 → 清 token, 跳登录
  if (resp.status === 401) {
    clearToken()
    window.location.hash = '#/login'
    throw new Error('未登录或 token 已过期')
  }

  if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
  const json = await resp.json().catch(() => ({}))
  if (json.code != null && json.code !== 0) {
    throw new Error(json.message || `code=${json.code}`)
  }
  return json.data ?? json
}

export const api = {
  // 登录
  login: (password) => request('/api/v2/login', { method: 'POST', body: { password } }),

  // 平台
  platformDelete: (id) => request(`/api/v2/platforms/${id}`, { method: 'DELETE' }),
  platformList: () => request('/api/v2/platforms'),
  platformCreate: (body) => request('/api/v2/platforms', { method: 'POST', body }),
  platformUpdate: (id, body) => request(`/api/v2/platforms/${id}`, { method: 'PUT', body }),

  // 账户
  accountDelete: (id) => request(`/api/v2/accounts/${id}`, { method: 'DELETE' }),
  accountList: () => request('/api/v2/accounts'),
  accountCreate: (body) => request('/api/v2/accounts', { method: 'POST', body }),
  accountUpdate: (id, body) => request(`/api/v2/accounts/${id}`, { method: 'PUT', body }),
  accountMonitorURL: (id) => request(`/api/v2/accounts/${id}/monitor-url`),
  accountRequests: (id, type) => request(`/api/v2/accounts/${id}/requests`, { query: { type } }),

  // 映射
  mappingList: (accountId) => request('/api/v2/mappings', { query: { accountId } }),
  mappingSave: (body) => request('/api/v2/mappings', { method: 'POST', body }),
  mappingDelete: (id) => request(`/api/v2/mappings/${id}`, { method: 'DELETE' }),
  mappingOptions: () => request('/api/v2/mapping-options'),

  // 统计
  overview: () => request('/api/v2/stats/overview'),
  qps: () => request('/api/v2/stats/qps'),
  statsReport: (query) => request('/api/v2/stats/report', { query }),
  statsResetToday: () => request('/api/v2/stats/reset-today', { method: 'POST' }),

  // 导出
  exportFiles: (query) => request('/api/v2/stats/export-files', { query }),
  exportDownloadURL: (filename) => `/api/v2/stats/export-download?filename=${encodeURIComponent(filename)}`,
  exportDelete: (filename) => request('/api/v2/stats/export-delete', { method: 'POST', body: { filename } }),

  // 直发查询
  issueList: () => request('/api/v2/issue-monitors'),
  issueCreate: (body) => request('/api/v2/issue-monitors', { method: 'POST', body }),
  issueUpdate: (id, body) => request(`/api/v2/issue-monitors/${id}`, { method: 'PUT', body }),
  issueDelete: (id) => request(`/api/v2/issue-monitors/${id}`, { method: 'DELETE' }),
  issueCheck: (id) => request(`/api/v2/issue-monitors/${id}/check`, { method: 'POST' }),
  issueLogs: (id, limit = 50) => request(`/api/v2/issue-monitors/${id}/logs`, { query: { limit } }),
  issueStats: (id, hours = 24) => request(`/api/v2/issue-monitors/${id}/stats`, { query: { hours } }),
  issueReport: (query) => request('/api/v2/issue-report', { query }),

  // 邮件
  smtpGet: () => request('/api/v2/settings/smtp'),
  smtpSave: (body) => request('/api/v2/settings/smtp', { method: 'PUT', body }),
  smtpTest: (to) => request('/api/v2/settings/smtp/test', { method: 'POST', body: { to } }),

  // 系统更新
  updateStatus: () => request('/api/v2/system/update'),
  updateCheck: () => request('/api/v2/system/update/check', { method: 'POST' }),
  updateStart: () => request('/api/v2/system/update/start', { method: 'POST' }),
}
