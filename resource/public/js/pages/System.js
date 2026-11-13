import { ref, computed, inject, onMounted, onUnmounted } from 'vue'
import { api } from '../api.js?v=4'

const ACTIVE = { downloading: '正在下载', verifying: '正在校验', restarting: '正在重启' }

export default {
  setup() {
    const toast = inject('toast')
    const st = ref(null)
    const checking = ref(false)
    const confirmShow = ref(false)
    const starting = ref(false)
    const offline = ref(false)
    let baseVersion = ''
    let timer = null

    const active = computed(() => !!ACTIVE[st.value?.phase])

    // 更新进行中: 每 2 秒轮询, 服务重启期间的请求失败静默忽略
    const poll = async () => {
      try {
        const s = await api.updateStatus()
        offline.value = false
        if (baseVersion && s.currentVersion && s.currentVersion !== baseVersion) {
          clearInterval(timer)
          toast.show('已更新到 ' + s.currentVersion + '，即将刷新页面', 'success')
          setTimeout(() => location.reload(), 1000)
          return
        }
        st.value = s
        if (!ACTIVE[s.phase]) clearInterval(timer)
      } catch { offline.value = true }
    }
    const startPolling = (v) => {
      baseVersion = v
      clearInterval(timer)
      timer = setInterval(poll, 2000)
    }

    const check = async () => {
      checking.value = true
      try {
        st.value = await api.updateCheck()
        toast.show(st.value.updateAvailable ? '发现新版本 ' + st.value.latestVersion : '已是最新版本', 'success')
      } catch (e) { toast.show('检查失败: ' + e.message, 'error') }
      checking.value = false
    }

    const start = async () => {
      starting.value = true
      try {
        const before = st.value.currentVersion
        st.value = await api.updateStart()
        confirmShow.value = false
        startPolling(before)
      } catch (e) { toast.show('启动更新失败: ' + e.message, 'error') }
      starting.value = false
    }

    onMounted(async () => {
      try {
        st.value = await api.updateStatus()
        if (active.value) startPolling(st.value.currentVersion)
        else if (!st.value.checkedAt) st.value = await api.updateCheck()
      } catch (e) { toast.show('获取更新状态失败: ' + e.message, 'error') }
    })
    onUnmounted(() => clearInterval(timer))

    return { st, ACTIVE, active, checking, check, confirmShow, starting, start, offline }
  },
  template: `
    <div class="space-y-4">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <i data-lucide="refresh-cw" class="w-6 h-6"></i> 系统更新
      </h1>

      <div v-if="!st" class="flex justify-center py-20"><span class="loading loading-spinner loading-lg"></span></div>

      <template v-else>
        <div v-if="!st.supported" class="alert alert-info">当前环境不支持在线更新：{{ st.unsupportedReason || '未知原因' }}</div>

        <div v-if="active" class="alert alert-warning">
          <span class="loading loading-spinner loading-sm"></span>
          <span>{{ ACTIVE[st.phase] }}… {{ st.message }}<template v-if="offline">（服务重启中，等待恢复）</template></span>
        </div>
        <div v-if="st.phase === 'failed'" class="alert alert-error">更新失败：{{ st.error || st.message }}</div>

        <div class="card bg-base-100 shadow">
          <div class="card-body">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div class="stat bg-base-200 rounded-box">
                <div class="stat-title">当前版本</div>
                <div class="stat-value text-2xl">{{ st.currentVersion || '—' }}</div>
              </div>
              <div class="stat bg-base-200 rounded-box">
                <div class="stat-title">最新版本</div>
                <div class="stat-value text-2xl" :class="{'text-primary': st.updateAvailable}">{{ st.latestVersion || '—' }}</div>
                <div class="stat-desc" v-if="st.publishedAt">发布于 {{ st.publishedAt }}</div>
              </div>
              <div class="stat bg-base-200 rounded-box">
                <div class="stat-title">上次检查</div>
                <div class="stat-value text-base">{{ st.checkedAt || '从未检查' }}</div>
                <div class="stat-desc">
                  <span v-if="st.updateAvailable" class="badge badge-primary badge-sm">有可用更新</span>
                  <span v-else-if="st.checkedAt">已是最新</span>
                </div>
              </div>
            </div>

            <div class="flex flex-wrap gap-2 mt-2">
              <button class="btn btn-sm" :disabled="checking || active" @click="check">
                <i data-lucide="search" class="w-4 h-4"></i> {{ checking ? '检查中…' : '检查更新' }}
              </button>
              <button v-if="st.updateAvailable && st.supported" class="btn btn-sm btn-primary" :disabled="active" @click="confirmShow=true">
                <i data-lucide="download" class="w-4 h-4"></i> 立即更新
              </button>
              <a v-if="st.releaseUrl" :href="st.releaseUrl" target="_blank" rel="noopener" class="btn btn-sm btn-ghost">
                <i data-lucide="external-link" class="w-4 h-4"></i> 查看发布页
              </a>
            </div>

            <div v-if="st.releaseNotes" class="mt-2">
              <h3 class="font-semibold mb-1">更新说明</h3>
              <div class="bg-base-200 rounded-box p-3 text-sm whitespace-pre-wrap break-words max-h-96 overflow-auto">{{ st.releaseNotes }}</div>
            </div>
          </div>
        </div>
      </template>

      <Modal :show="confirmShow" title="确认更新" :confirm-text="starting ? '启动中…' : '开始更新'" :confirm-disabled="starting"
        @close="!starting && (confirmShow=false)" @confirm="start">
        <p>将从 {{ st?.currentVersion }} 更新到 {{ st?.latestVersion }}。</p>
        <p class="mt-2 text-warning">更新期间服务会自动重启约数秒，页面将在完成后自动刷新。</p>
      </Modal>
    </div>
  `,
}
