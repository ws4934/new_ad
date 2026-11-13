import { ref, computed, inject, watch, onMounted } from 'vue'
import { api } from '../api.js?v=8'
import TrendChart from '../components/TrendChart.js?v=8'

const CATS = { 1: '超级红包', 2: '福利购', 3: '签到红包', 4: '福利直降', 5: '幸运赢免单', 6: '超级金币' }
const bjToday = () => new Date(Date.now() + 8 * 3600e3).toISOString().slice(0, 10)
const pct = (use, fund) => fund ? (use / fund * 100).toFixed(1) + '%' : '—'

export default {
  components: { TrendChart },
  setup() {
    const toast = inject('toast')
    const monitors = ref([])
    const loaded = ref(false)
    const appKey = ref('')
    const category = ref(1)
    const date = ref(bjToday())
    const report = ref(null)
    const trend = ref([])
    const loading = ref(false)
    const error = ref('')
    const search = ref('')

    // 去重 AppKey, 用第一个配置名做说明
    const appKeys = computed(() => {
      const m = new Map()
      for (const x of monitors.value) if (!m.has(x.appKey)) m.set(x.appKey, x.name)
      return [...m].map(([key, name]) => ({ key, name }))
    })

    let seq = 0 // 丢弃过期请求的结果 (筛选连续切换时)
    const load = async () => {
      if (!appKey.value) return
      const my = ++seq
      loading.value = true
      error.value = ''
      const q = { appKey: appKey.value, category: category.value }
      const [r, t] = await Promise.allSettled([
        api.issueReport({ ...q, date: date.value.replaceAll('-', '') }),
        api.issueReportTrend({ ...q, days: 7 }),
      ])
      if (my !== seq) return
      if (r.status === 'fulfilled') report.value = r.value
      else { report.value = null; error.value = r.reason.message }
      if (t.status === 'fulfilled') trend.value = t.value.list || []
      else { trend.value = []; error.value = error.value || t.reason.message }
      if (error.value) toast.show('加载失败: ' + error.value, 'error')
      loading.value = false
    }

    const rows = computed(() => {
      const kw = search.value.trim()
      return [...(report.value?.list || [])]
        .filter(r => !kw || String(r.pid).includes(kw))
        .sort((a, b) => b.fundNum - a.fundNum)
    })

    const chart = computed(() => ({
      labels: trend.value.map(d => d.date.slice(4, 6) + '-' + d.date.slice(6)),
      series: [
        { name: '发放', color: '#ea580c', points: trend.value.map(d => d.fund) },
        { name: '使用', color: '#059669', points: trend.value.map(d => d.use) },
      ],
      titles: trend.value.map(d => `${d.date.slice(0, 4)}-${d.date.slice(4, 6)}-${d.date.slice(6)}`),
      notes: trend.value.map(d => '使用率：' + pct(d.use, d.fund)),
    }))

    const copy = async (pid) => {
      try { await navigator.clipboard.writeText(pid); toast.show('已复制 ' + pid, 'success') }
      catch { toast.show('复制失败，请手动选择复制', 'error') }
    }

    watch([appKey, category, date], load)

    onMounted(async () => {
      try {
        monitors.value = (await api.issueList()).list || []
        const first = monitors.value[0]
        if (first) { category.value = first.activityCategory; appKey.value = first.appKey }
      } catch (e) { error.value = e.message; toast.show('加载配置失败: ' + e.message, 'error') }
      loaded.value = true
    })

    return { CATS, appKeys, appKey, category, date, report, loading, error, search, rows, chart, load, copy, pct, loaded }
  },
  template: `
    <div class="space-y-4">
      <div class="flex flex-wrap justify-between items-center gap-3">
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <i data-lucide="bar-chart-3" class="w-6 h-6"></i> 红包报表
        </h1>
        <div v-if="appKeys.length" class="flex flex-wrap items-center gap-2">
          <select v-model="appKey" class="select select-bordered select-sm">
            <option v-for="a in appKeys" :key="a.key" :value="a.key">{{ a.key }}（{{ a.name }}…）</option>
          </select>
          <select v-model.number="category" class="select select-bordered select-sm">
            <option v-for="(n, c) in CATS" :key="c" :value="Number(c)">{{ n }}</option>
          </select>
          <input type="date" v-model="date" class="input input-bordered input-sm">
          <button class="btn btn-primary btn-sm" :disabled="loading" @click="load">
            <span v-if="loading" class="loading loading-spinner loading-xs"></span>
            <i v-else data-lucide="refresh-cw" class="w-4 h-4"></i> 刷新
          </button>
        </div>
      </div>

      <div v-if="!loaded" class="flex justify-center py-20"><span class="loading loading-spinner loading-lg"></span></div>

      <div v-else-if="!appKeys.length" class="card bg-base-100 shadow">
        <div class="card-body items-center text-center text-base-content/60">
          <p>{{ error ? '加载配置失败：' + error : '还没有直发查询配置，无可用的 AppKey' }}</p>
          <router-link to="/issue-config" class="btn btn-primary btn-sm">去添加配置</router-link>
        </div>
      </div>

      <template v-else>
        <p v-if="error" class="text-sm text-error flex items-center gap-1.5">
          <i data-lucide="alert-circle" class="w-4 h-4 shrink-0"></i>{{ error }}
        </p>

        <div class="card bg-base-100 shadow">
          <div class="card-body p-5">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">发放数</div>
                <div class="text-2xl font-bold tabular-nums mt-1">{{ report ? report.totalFund : '—' }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">使用数</div>
                <div class="text-2xl font-bold tabular-nums mt-1 text-emerald-700">{{ report ? report.totalUse : '—' }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">使用率</div>
                <div class="text-2xl font-bold tabular-nums mt-1">{{ report ? pct(report.totalUse, report.totalFund) : '—' }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">推广位数</div>
                <div class="text-2xl font-bold tabular-nums mt-1">{{ report ? (report.list || []).length : '—' }}</div>
              </div>
            </div>
          </div>
        </div>

        <div class="card bg-base-100 shadow">
          <div class="card-body p-5 gap-3">
            <div>
              <h2 class="font-semibold">近 7 天趋势</h2>
              <p class="text-xs text-base-content/50">截至今天，按北京时间日统计</p>
            </div>
            <div v-if="loading && !chart.labels.length" class="flex justify-center py-12"><span class="loading loading-spinner"></span></div>
            <p v-else-if="!chart.labels.length" class="text-sm text-base-content/50 py-8 text-center">暂无数据</p>
            <TrendChart v-else :labels="chart.labels" :series="chart.series" :titles="chart.titles" :notes="chart.notes"/>
          </div>
        </div>

        <div class="card bg-base-100 shadow">
          <div class="card-body p-5 gap-3">
            <div class="flex flex-wrap justify-between items-center gap-2">
              <div>
                <h2 class="font-semibold">推广位明细</h2>
                <p class="text-xs text-base-content/50">按发放数从高到低，点击 PID 复制</p>
              </div>
              <input v-model="search" class="input input-bordered input-sm w-56" placeholder="搜索 PID">
            </div>
            <div v-if="loading && !report" class="flex justify-center py-8"><span class="loading loading-spinner"></span></div>
            <div v-else class="overflow-x-auto">
              <table class="table table-sm">
                <thead><tr><th>PID</th><th class="text-right">发放数</th><th class="text-right">使用数</th><th class="text-right">使用率</th></tr></thead>
                <tbody>
                  <tr v-for="r in rows" :key="r.pid">
                    <td><code class="font-mono text-xs cursor-pointer hover:text-primary" title="点击复制" @click="copy(r.pid)">{{ r.pid }}</code></td>
                    <td class="text-right tabular-nums">{{ r.fundNum }}</td>
                    <td class="text-right tabular-nums">{{ r.useNum }}</td>
                    <td class="text-right tabular-nums">{{ pct(r.useNum, r.fundNum) }}</td>
                  </tr>
                  <tr v-if="!rows.length">
                    <td colspan="4" class="text-center py-6 text-base-content/40">{{ search.trim() ? '没有匹配的 PID' : '暂无数据' }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>

        <p class="text-xs text-base-content/40">数据来自 taobao.tbk.dg.vegas.send.report，按北京时间日统计，存在延迟；使用数可能包含此前发放的红包，因此可能大于发放数。</p>
      </template>
    </div>
  `,
}
