import { createApp, ref, computed, watch, onMounted, onUpdated, nextTick } from 'vue'
import { router } from './router.js?v=5'
import { api, clearToken } from './api.js?v=5'
import Modal from './components/Modal.js'
import MultiSelect from './components/MultiSelect.js'
import { useToast, ToastView } from './components/Toast.js'

const { createIcons } = window.lucide

// 全局唯一 toast 实例
const globalToast = useToast()

// 侧边栏菜单 (分组)
const MENU = [
  { title: '概览', items: [
    { to: '/', icon: 'home', text: '仪表盘' },
  ] },
  { title: '广告监测', items: [
    { to: '/platforms', icon: 'settings', text: '平台管理' },
    { to: '/accounts', icon: 'users', text: '账户管理' },
    { to: '/stats', icon: 'line-chart', text: '统计报表' },
    { to: '/export', icon: 'file-down', text: 'CSV 导出' },
  ] },
  { title: '淘宝客直发', items: [
    { to: '/issue', icon: 'gift', text: '直发查询' },
    { to: '/issue-report', icon: 'bar-chart-3', text: '红包报表' },
    { to: '/issue-config', icon: 'sliders-horizontal', text: '直发配置' },
  ] },
  { title: '系统', items: [
    { to: '/system', icon: 'refresh-cw', text: '系统更新' },
  ] },
]

const App = {
  components: { ToastView },
  setup() {
    const refreshIcons = () => nextTick(() => createIcons())
    onMounted(refreshIcons)
    onUpdated(refreshIcons)

    const path = computed(() => router.currentRoute.value.path)
    const showSidebar = computed(() => path.value !== '/login')
    const logout = () => { clearToken(); router.push('/login') }

    // 子路由 (如 /accounts/1) 也高亮父菜单
    const isSub = (to) => to !== '/' && path.value.startsWith(to + '/')
    const title = computed(() => {
      for (const g of MENU) for (const i of g.items) if (i.to === path.value || isSub(i.to)) return i.text
      return '广告监测平台'
    })

    // 移动端: 导航后收起抽屉
    const drawerOpen = ref(false)
    watch(path, () => { drawerOpen.value = false })

    // 版本号 + 更新提示, 登录后加载一次, 失败忽略
    const version = ref('')
    const updateAvailable = ref(false)
    watch(showSidebar, async (v) => {
      if (!v || version.value) return
      try {
        const s = await api.updateStatus()
        version.value = s.currentVersion || ''
        updateAvailable.value = !!s.updateAvailable
      } catch { /* 忽略 */ }
    }, { immediate: true })

    return { MENU, showSidebar, logout, toast: globalToast, isSub, title, drawerOpen, version, updateAvailable }
  },
  template: `
    <ToastView :message="toast.message.value" :type="toast.type.value"/>

    <router-view v-if="!showSidebar"/>

    <div v-else class="drawer lg:drawer-open">
      <input id="nav-drawer" type="checkbox" class="drawer-toggle" v-model="drawerOpen">
      <div class="drawer-content flex flex-col min-h-screen min-w-0">
        <header class="lg:hidden sticky top-0 z-30 bg-base-100 shadow-sm flex items-center gap-2 px-2 h-12">
          <label for="nav-drawer" class="btn btn-ghost btn-sm btn-square" aria-label="打开菜单">
            <i data-lucide="menu" class="w-5 h-5"></i>
          </label>
          <span class="font-semibold truncate">{{ title }}</span>
        </header>
        <main class="flex-1 p-4 lg:p-6 overflow-auto">
          <router-view/>
        </main>
      </div>
      <div class="drawer-side z-40">
        <label for="nav-drawer" class="drawer-overlay" aria-label="关闭菜单"></label>
        <aside class="w-60 min-h-full bg-base-100 shadow-lg flex flex-col">
          <div class="flex items-center gap-2.5 px-4 h-14 border-b border-base-200 shrink-0">
            <span class="w-8 h-8 rounded-lg bg-primary text-primary-content grid place-items-center">
              <i data-lucide="radar" class="w-4 h-4"></i>
            </span>
            <span class="font-bold">广告监测平台</span>
          </div>
          <ul class="menu p-2 flex-1 w-full">
            <template v-for="g in MENU" :key="g.title">
              <li class="menu-title text-xs font-medium text-base-content/40 mt-2">{{ g.title }}</li>
              <li v-for="i in g.items" :key="i.to">
                <router-link :to="i.to" class="flex items-center gap-2" :class="{'router-link-active': isSub(i.to)}">
                  <i :data-lucide="i.icon" class="w-4 h-4"></i>
                  <span class="flex-1">{{ i.text }}</span>
                  <span v-if="i.to === '/system' && updateAvailable" class="w-1.5 h-1.5 rounded-full bg-primary" title="有可用更新"></span>
                </router-link>
              </li>
            </template>
          </ul>
          <div class="p-2 border-t border-base-200 shrink-0">
            <div v-if="version" class="px-4 pb-1 text-[11px] text-base-content/40 tabular-nums">版本 {{ version }}</div>
            <button class="btn btn-ghost btn-sm w-full justify-start gap-2" @click="logout">
              <i data-lucide="log-out" class="w-4 h-4"></i> 退出登录
            </button>
          </div>
        </aside>
      </div>
    </div>
  `,
}

const app = createApp(App)
app.use(router)
app.component('Modal', Modal)
app.component('MultiSelect', MultiSelect)

// provide 同一个 toast 实例给所有子组件
app.provide('toast', globalToast)

app.mixin({
  updated() { nextTick(() => createIcons()) },
  mounted() { nextTick(() => createIcons()) },
})
app.mount('#app')
