import { createApp, computed, onMounted, onUpdated, nextTick } from 'vue'
import { router } from './router.js'
import { clearToken } from './api.js'
import Modal from './components/Modal.js'
import MultiSelect from './components/MultiSelect.js'
import { useToast, ToastView } from './components/Toast.js'

const { createIcons } = window.lucide

// 全局唯一 toast 实例
const globalToast = useToast()

const App = {
  components: { ToastView },
  setup() {
    const refreshIcons = () => nextTick(() => createIcons())
    onMounted(refreshIcons)
    onUpdated(refreshIcons)

    const showSidebar = computed(() => router.currentRoute.value.path !== '/login')
    const logout = () => { clearToken(); router.push('/login') }

    return { showSidebar, logout, toast: globalToast }
  },
  template: `
    <ToastView :message="toast.message.value" :type="toast.type.value"/>

    <router-view v-if="!showSidebar"/>

    <div v-else class="min-h-screen flex">
      <aside class="w-56 bg-base-100 shadow-lg flex flex-col">
        <ul class="menu p-2 flex-1">
          <li><router-link to="/" class="flex items-center gap-2">
            <i data-lucide="home" class="w-4 h-4"></i> 仪表盘
          </router-link></li>
          <li><router-link to="/platforms" class="flex items-center gap-2">
            <i data-lucide="settings" class="w-4 h-4"></i> 平台管理
          </router-link></li>
          <li><router-link to="/accounts" class="flex items-center gap-2">
            <i data-lucide="users" class="w-4 h-4"></i> 账户管理
          </router-link></li>
          <li><router-link to="/stats" class="flex items-center gap-2">
            <i data-lucide="line-chart" class="w-4 h-4"></i> 统计报表
          </router-link></li>
          <li><router-link to="/export" class="flex items-center gap-2">
            <i data-lucide="file-down" class="w-4 h-4"></i> CSV 导出
          </router-link></li>
        </ul>
        <div class="p-2 border-t">
          <button class="btn btn-ghost btn-sm w-full justify-start gap-2" @click="logout">
            <i data-lucide="log-out" class="w-4 h-4"></i> 退出登录
          </button>
        </div>
      </aside>
      <main class="flex-1 p-6 overflow-auto">
        <router-view/>
      </main>
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
