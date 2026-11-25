import { ref, computed, inject, onMounted, onUnmounted } from 'vue'
import { api } from '../api.js?v=15'

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
      <div class="flex justify-between items-center">
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <i data-lucide="refresh-cw" class="w-6 h-6"></i> 系统更新
        </h1>
        <div v-if="st" class="flex gap-2">
          <a v-if="st.releaseUrl" :href="st.releaseUrl" target="_blank" rel="noopener" class="btn btn-sm btn-ghost">
            <i data-lucide="external-link" class="w-4 h-4"></i> 发布页
          </a>
          <button class="btn btn-sm" :disabled="checking || active" @click="check">
            <span v-if="checking" class="loading loading-spinner loading-xs"></span>
            <i v-else data-lucide="search" class="w-4 h-4"></i> 检查更新
          </button>
          <button v-if="st.updateAvailable && st.supported" class="btn btn-sm btn-primary" :disabled="active" @click="confirmShow=true">
            <i data-lucide="download" class="w-4 h-4"></i> 立即更新
          </button>
        </div>
      </div>

      <div v-if="!st" class="flex justify-center py-20"><span class="loading loading-spinner loading-lg"></span></div>

      <template v-else>
        <div v-if="active" class="alert alert-warning text-sm">
          <span class="loading loading-spinner loading-sm"></span>
          <span>{{ ACTIVE[st.phase] }}… {{ st.message }}<template v-if="offline">（服务重启中，等待恢复）</template></span>
        </div>
        <div v-if="st.phase === 'failed'" class="alert alert-error text-sm">更新失败：{{ st.error || st.message }}</div>

        <div class="card bg-base-100 shadow">
          <div class="card-body p-5 gap-4">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">当前版本</div>
                <div class="text-2xl font-bold tabular-nums mt-1">{{ st.currentVersion || '—' }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">最新版本</div>
                <div class="text-2xl font-bold tabular-nums mt-1" :class="{'text-primary': st.updateAvailable}">{{ st.latestVersion || '—' }}</div>
                <div class="text-xs text-base-content/50 mt-1" v-if="st.publishedAt">发布于 {{ st.publishedAt.slice(0, 10) }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">上次检查</div>
                <div class="text-base font-semibold mt-1.5">{{ st.checkedAt || '从未检查' }}</div>
                <div class="text-xs mt-1" :class="st.updateAvailable ? 'text-primary' : 'text-base-content/50'">
                  {{ st.updateAvailable ? '有可用更新' : st.checkedAt ? '已是最新' : '' }}
                </div>
              </div>
            </div>

            <p v-if="!st.supported" class="text-xs text-base-content/60 flex items-center gap-1.5">
              <i data-lucide="info" class="w-3.5 h-3.5"></i> 当前环境不支持在线更新：{{ st.unsupportedReason || '未知原因' }}
            </p>

            <div v-if="st.releaseNotes" class="pt-4 border-t border-base-200">
              <h2 class="font-semibold mb-2">更新说明</h2>
              <div class="text-sm whitespace-pre-wrap break-words max-h-96 overflow-auto text-base-content/80">{{ st.releaseNotes }}</div>
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
