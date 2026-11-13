import { ref, computed, inject, onMounted, onUnmounted } from 'vue'
import { api } from '../api.js?v=4'

const CATS = { 1: '超级红包', 2: '福利购', 4: '福利直降' }

export default {
  setup() {
    const toast = inject('toast')
    const list = ref([])
    const loaded = ref(false)
    const error = ref('')
    const filter = ref(0)
    const checking = ref({})
    const now = ref(Date.now())
    const logTarget = ref(null)
    const logs = ref([])
    const logsLoading = ref(false)
    let timer = null

    const refresh = async () => {
      now.value = Date.now()
      try {
        const res = await api.issueList()
        list.value = res.list || []
        error.value = ''
      } catch (e) { error.value = e.message }
      loaded.value = true
    }

    const tabs = computed(() => [0, ...[1, 2, 4].filter(c => list.value.some(m => m.activityCategory === c))])
    const cur = computed(() => tabs.value.includes(filter.value) ? filter.value : 0)
    const filtered = computed(() => cur.value ? list.value.filter(m => m.activityCategory === cur.value) : list.value)

    const check = async (m) => {
      checking.value[m.id] = true
      try {
        const res = await api.issueCheck(m.id)
        const i = list.value.findIndex(x => x.id === m.id)
        if (i >= 0 && res.monitor) list.value[i] = res.monitor
        if (res.monitor?.lastError) toast.show('查询失败: ' + res.monitor.lastError, 'error')
        else toast.show('查询完成', 'success')
      } catch (e) { toast.show('查询失败: ' + e.message, 'error') }
      checking.value[m.id] = false
    }

    const openLogs = async (m) => {
      logTarget.value = m
      logs.value = []
      logsLoading.value = true
      try { logs.value = (await api.issueLogs(m.id, 50)).list || [] }
      catch (e) { toast.show('加载失败: ' + e.message, 'error') }
      logsLoading.value = false
    }

    // 后端时间为 "YYYY-MM-DD HH:mm:ss"，按浏览器本地时区解析
    const since = (t) => {
      const sec = Math.max(0, Math.floor((now.value - new Date(t.replace(' ', 'T')).getTime()) / 1000))
      if (isNaN(sec)) return '—'
      const d = Math.floor(sec / 86400), h = Math.floor(sec % 86400 / 3600), mi = Math.floor(sec % 3600 / 60)
      if (d) return `${d} 天 ${h} 小时`
      if (h) return `${h} 小时 ${mi} 分钟`
      return mi ? `${mi} 分钟` : '不足 1 分钟'
    }
    const stateText = (v) => v === true ? '直发中' : v === false ? '手动领取' : '未查询'
    const stateClass = (v) => v === true ? 'badge-success' : v === false ? 'badge-warning' : 'badge-ghost'
    const num = (v) => v == null ? '—' : v

    onMounted(() => { refresh(); timer = setInterval(refresh, 5000) })
    onUnmounted(() => clearInterval(timer))

    return { CATS, list, loaded, error, filter, cur, tabs, filtered, checking, check,
      logTarget, logs, logsLoading, openLogs, since, stateText, stateClass, num }
  },
  template: `
    <div class="space-y-4">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <i data-lucide="gift" class="w-6 h-6"></i> 直发查询
        <span class="text-xs font-normal text-gray-400">每 5 秒自动刷新</span>
      </h1>

      <div v-if="error" class="alert alert-error">{{ error }}</div>

      <div v-if="loaded && !list.length" class="card bg-base-100 shadow">
        <div class="card-body items-center text-center text-gray-500">
          <p>还没有直发查询配置</p>
          <router-link to="/issue-config" class="btn btn-primary btn-sm">去添加配置</router-link>
        </div>
      </div>

      <template v-else-if="list.length">
        <div role="tablist" class="tabs tabs-boxed bg-base-100 w-fit">
          <a v-for="t in tabs" :key="t" role="tab" class="tab"
             :class="{'tab-active': cur === t}" @click="filter=t">
            {{ t ? CATS[t] : '全部' }}
          </a>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          <div v-for="m in filtered" :key="m.id" class="card bg-base-100 shadow">
            <div class="card-body p-4 gap-3">
              <div class="flex items-start justify-between gap-2">
                <div class="min-w-0">
                  <h2 class="font-bold text-lg truncate">{{ m.name }}</h2>
                  <div class="text-xs text-gray-500">AppKey: <code>{{ m.appKey }}</code></div>
                </div>
                <span class="badge" :class="stateClass(m.canIssue)">{{ stateText(m.canIssue) }}</span>
              </div>
              <div class="flex flex-wrap gap-1">
                <span class="badge badge-outline badge-sm">{{ CATS[m.activityCategory] || m.activityCategory }}</span>
                <span class="badge badge-sm" :class="m.enabled ? 'badge-primary' : 'badge-ghost'">{{ m.enabled ? '已启用' : '已停用' }}</span>
              </div>

              <div v-if="m.lastError" class="alert alert-error py-2 text-xs break-all">{{ m.lastError }}</div>

              <div class="bg-base-200 rounded-box p-3">
                <template v-if="m.canIssue === false">
                  <div v-if="m.manualClaim2Go == null" class="text-lg font-semibold text-warning">还需发放：超过展示门槛</div>
                  <div v-else class="text-lg">还需发放 <span class="text-3xl font-bold text-warning">{{ m.manualClaim2Go }}</span> 个红包达标</div>
                </template>
                <template v-else-if="m.canIssue === true">
                  <div v-if="m.autoPayoutCnt == null" class="text-lg font-semibold text-success">剩余可直发：超过展示门槛（充足）</div>
                  <div v-else class="text-lg">剩余可直发 <span class="text-3xl font-bold text-success">{{ m.autoPayoutCnt }}</span> 个</div>
                  <div class="text-sm text-gray-500 mt-1">
                    距下轮直发目标{{ m.manualClaim2Go == null ? '：超过展示门槛' : '还需 ' + m.manualClaim2Go + ' 个' }}
                  </div>
                </template>
                <div v-else class="text-gray-400">暂无数据，等待首次查询</div>
              </div>

              <div v-if="m.extraMsg" class="text-sm">提示：{{ m.extraMsg }}</div>
              <div class="text-xs text-gray-500 space-y-0.5">
                <div>最后查询：{{ m.checkedAt || '—' }}</div>
                <div v-if="m.changedAt">状态持续：{{ since(m.changedAt) }}（自 {{ m.changedAt }}）</div>
                <div>查询间隔：{{ m.intervalSec }} 秒</div>
              </div>

              <div class="card-actions justify-end">
                <button class="btn btn-sm btn-ghost" @click="openLogs(m)">
                  <i data-lucide="history" class="w-4 h-4"></i> 变化记录
                </button>
                <button class="btn btn-sm btn-primary" :disabled="checking[m.id]" @click="check(m)">
                  <i data-lucide="search" class="w-4 h-4"></i> {{ checking[m.id] ? '查询中…' : '立即查询' }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </template>

      <Modal :show="!!logTarget" :title="'变化记录 - ' + (logTarget?.name || '')" width="max-w-3xl" :show-footer="false" @close="logTarget=null">
        <div v-if="logsLoading" class="flex justify-center py-8"><span class="loading loading-spinner"></span></div>
        <div v-else class="overflow-x-auto max-h-[60vh]">
          <table class="table table-sm table-pin-rows">
            <thead><tr><th>时间</th><th>状态</th><th>剩余可直发</th><th>还需发放</th><th>状态翻转</th><th>邮件结果</th></tr></thead>
            <tbody>
              <tr v-for="l in logs" :key="l.id">
                <td class="whitespace-nowrap">{{ l.createdAt }}</td>
                <td><span class="badge badge-sm" :class="stateClass(l.canIssue)">{{ stateText(l.canIssue) }}</span></td>
                <td>{{ num(l.autoPayoutCnt) }}</td>
                <td>{{ num(l.manualClaim2Go) }}</td>
                <td><span v-if="l.statusChanged" class="badge badge-sm badge-error">翻转</span></td>
                <td class="text-xs">{{ l.mailResult || '' }}</td>
              </tr>
              <tr v-if="!logs.length"><td colspan="6" class="text-center text-gray-400">暂无记录</td></tr>
            </tbody>
          </table>
          <p class="text-xs text-gray-400 mt-2">"—" 表示接口未返回数值（未达展示门槛或不适用）</p>
        </div>
        <div class="modal-action"><button class="btn btn-sm" @click="logTarget=null">关闭</button></div>
      </Modal>
    </div>
  `,
}
