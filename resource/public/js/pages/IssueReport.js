import { ref, computed, inject, watch, onMounted } from 'vue'
import { api } from '../api.js?v=21'
import TrendChart from '../components/TrendChart.js?v=21'

const CATS = { 1: '超级红包', 2: '福利购', 3: '签到红包', 4: '福利直降', 5: '幸运赢免单', 6: '超级金币' }
// 北京时间 n 天前 (YYYY-MM-DD)
const bjDay = (n = 0) => new Date(Date.now() + 8 * 3600e3 - n * 864e5).toISOString().slice(0, 10)
const PRESETS = [{ text: '今天', days: 1 }, { text: '近 7 天', days: 7 }, { text: '近 30 天', days: 30 }]
const MAX_DAYS = 31
const addDays = (d, n) => new Date(new Date(d + 'T00:00:00Z').getTime() + n * 864e5).toISOString().slice(0, 10)
const pct = (use, fund) => fund ? (use / fund * 100).toFixed(1) + '%' : '—'

export default {
  components: { TrendChart },
  setup() {
    const toast = inject('toast')
    const monitors = ref([])
    const loaded = ref(false)
    const appKey = ref('')
    const category = ref(1)
    const from = ref(bjDay(6))
    const to = ref(bjDay())
    const report = ref(null)
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
      let r = null
      try {
        r = await api.issueReport({ appKey: appKey.value, category: category.value,
          from: from.value.replaceAll('-', ''), to: to.value.replaceAll('-', '') })
      } catch (e) { error.value = e.message }
      if (my !== seq) return
      report.value = r
      if (error.value) toast.show('加载失败: ' + error.value, 'error')
      loading.value = false
    }

    const rows = computed(() => {
      const kw = search.value.trim()
      return [...(report.value?.list || [])]
        .filter(r => !kw || String(r.pid).includes(kw))
        .sort((a, b) => b.fundNum - a.fundNum)
    })

    const trend = computed(() => report.value?.trend || [])
    const fmt = (d) => `${d.slice(0, 4)}-${d.slice(4, 6)}-${d.slice(6)}`
    const failedText = computed(() => (report.value?.failed || []).map(d => d.slice(4, 6) + '-' + d.slice(6)).join('、'))
    const days = computed(() => Math.round((new Date(to.value) - new Date(from.value)) / 864e5) + 1)
    const preset = (n) => { to.value = bjDay(); from.value = bjDay(n - 1) }
    const isPreset = (n) => to.value === bjDay() && from.value === bjDay(n - 1)
    // 改开始/结束日期时保证 from ≤ to 且不超过 31 天, 只发一次请求
    const setFrom = (v) => { if (!v) return; from.value = v; if (to.value < v) to.value = v; else if (days.value > MAX_DAYS) to.value = addDays(v, MAX_DAYS - 1) }
    const setTo = (v) => { if (!v) return; to.value = v; if (from.value > v) from.value = v; else if (days.value > MAX_DAYS) from.value = addDays(v, -(MAX_DAYS - 1)) }

    const chart = computed(() => ({
      labels: trend.value.map(d => d.date.slice(4, 6) + '-' + d.date.slice(6)),
      series: [
        { name: '发放', color: '#ea580c', points: trend.value.map(d => d.fund) },
        { name: '使用', color: '#059669', points: trend.value.map(d => d.use) },
      ],
      titles: trend.value.map(d => fmt(d.date)),
      notes: trend.value.map(d => '使用率：' + pct(d.use, d.fund)),
    }))

    const copy = async (pid) => {
      try { await navigator.clipboard.writeText(pid); toast.show('已复制 ' + pid, 'success') }
      catch { toast.show('复制失败，请手动选择复制', 'error') }
    }

    watch([appKey, category, from, to], load)

    onMounted(async () => {
      try {
        monitors.value = (await api.issueList()).list || []
        const first = monitors.value[0]
        if (first) { category.value = first.activityCategory; appKey.value = first.appKey }
      } catch (e) { error.value = e.message; toast.show('加载配置失败: ' + e.message, 'error') }
      loaded.value = true
    })

    return { CATS, PRESETS, appKeys, appKey, category, from, to, today: bjDay(), report, loading, error, search, rows, chart,
      load, copy, pct, loaded, fmt, failedText, preset, isPreset, setFrom, setTo }
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
          <div class="join">
            <button v-for="p in PRESETS" :key="p.days" class="btn btn-sm join-item"
                    :class="isPreset(p.days) ? 'btn-active' : 'btn-ghost bg-base-100 border-base-300'" @click="preset(p.days)">{{ p.text }}</button>
          </div>
          <div class="flex items-center gap-1">
            <input type="date" :value="from" :max="today" @change="setFrom($event.target.value)" class="input input-bordered input-sm" aria-label="开始日期">
            <span class="text-base-content/40">至</span>
            <input type="date" :value="to" :max="today" @change="setTo($event.target.value)" class="input input-bordered input-sm" aria-label="结束日期">
          </div>
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

        <div class="card bg-base-100 shadow transition-opacity" :class="{ 'opacity-50': loading }">
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

        <div class="card bg-base-100 shadow transition-opacity" :class="{ 'opacity-50': loading }">
          <div class="card-body p-5 gap-3">
            <div>
              <h2 class="font-semibold">每日趋势</h2>
              <p class="text-xs text-base-content/50">{{ report ? fmt(report.from) + ' 至 ' + fmt(report.to) + '，共 ' + (report.trend.length + report.failed.length) + ' 天，按北京时间日统计' : '按北京时间日统计' }}</p>
            </div>
            <p v-if="failedText" class="text-xs text-warning">以下日期查询失败，未计入合计：{{ failedText }}</p>
            <div v-if="loading && !chart.labels.length" class="flex justify-center py-12"><span class="loading loading-spinner"></span></div>
            <p v-else-if="!chart.labels.length" class="text-sm text-base-content/50 py-8 text-center">暂无数据</p>
            <TrendChart v-else :labels="chart.labels" :series="chart.series" :titles="chart.titles" :notes="chart.notes"/>
          </div>
        </div>

        <div class="card bg-base-100 shadow transition-opacity" :class="{ 'opacity-50': loading }">
          <div class="card-body p-5 gap-3">
            <div class="flex flex-wrap justify-between items-center gap-2">
              <div>
                <h2 class="font-semibold">推广位明细</h2>
                <p class="text-xs text-base-content/50">所选范围内合计，按发放数从高到低，点击 PID 复制</p>
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
