import { ref, computed, inject, onMounted, onUnmounted } from 'vue'
import { api } from '../api.js?v=19'
import TrendChart from '../components/TrendChart.js?v=19'

const CATS = { 1: '超级红包', 2: '福利购', 4: '福利直降' }

export default {
  components: { TrendChart },
  setup() {
    const toast = inject('toast')
    const list = ref([])
    const loaded = ref(false)
    const error = ref('')
    const filter = ref(0)
    const checking = ref({})
    const now = ref(Date.now())
    const detailId = ref(0)
    const tab = ref('trend')
    const hours = ref(24)
    const stats = ref(null)
    const statsLoading = ref(false)
    const statsError = ref('')
    const logs = ref([])
    const logsLoading = ref(false)
    let logsFor = 0
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

    // 详情弹窗: 跟随列表刷新取最新 monitor (insight 每 5 秒更新)
    const detail = computed(() => list.value.find(m => m.id === detailId.value) || null)
    const loadStats = async () => {
      const id = detailId.value, h = hours.value
      statsLoading.value = true
      statsError.value = ''
      try {
        const r = await api.issueStats(id, h)
        if (id === detailId.value && h === hours.value) stats.value = r
      } catch (e) {
        statsError.value = e.message
        toast.show('加载趋势失败: ' + e.message, 'error')
      }
      statsLoading.value = false
    }
    const loadLogs = async () => {
      const id = logsFor = detailId.value
      logs.value = []
      logsLoading.value = true
      try {
        const r = (await api.issueLogs(id, 50)).list || []
        if (id === detailId.value) logs.value = r
      } catch (e) { toast.show('加载失败: ' + e.message, 'error') }
      logsLoading.value = false
    }
    const openDetail = (m) => {
      detailId.value = m.id
      tab.value = 'trend'
      stats.value = null
      logsFor = 0
      loadStats()
    }
    const closeDetail = () => { detailId.value = 0 }
    const setTab = (t) => {
      tab.value = t
      if (t === 'logs' && logsFor !== detailId.value) loadLogs()
    }
    const setHours = (h) => { hours.value = h; stats.value = null; loadStats() }
    const chart = computed(() => {
      const s = stats.value?.series || []
      return {
        labels: s.map(p => p.t),
        series: [
          { name: '还需手抽', color: '#d97706', points: s.map(p => p.manual) },
          { name: '剩余直发', color: '#059669', points: s.map(p => p.auto) },
        ],
        band: s.map(p => p.canIssue === true ? '#10b981' : p.canIssue === false ? '#f59e0b' : 'oklch(var(--b3))'),
        titles: s.map(p => p.full.slice(0, 16)),
        notes: s.map(p => '状态：' + (p.canIssue === true ? '直发中' : p.canIssue === false ? '手动领取' : '未知')),
        empty: s.every(p => p.manual == null && p.auto == null),
      }
    })

    // 后端时间不带时区: 用同源的 checkedAt - changedAt 计算持续时长, 不受浏览器时区影响
    const parse = (t) => new Date(t.replace(' ', 'T')).getTime()
    const dur = (sec) => {
      const d = Math.floor(sec / 86400), h = Math.floor(sec % 86400 / 3600), mi = Math.floor(sec % 3600 / 60)
      if (d) return `${d} 天 ${h} 小时`
      if (h) return `${h} 小时 ${mi} 分`
      return mi ? `${mi} 分钟` : '不足 1 分钟'
    }
    const since = (m) => {
      if (!m.changedAt || !m.checkedAt) return ''
      const sec = Math.max(0, Math.floor((parse(m.checkedAt) - parse(m.changedAt)) / 1000))
      return isNaN(sec) ? '' : dur(sec)
    }
    const STATE = {
      true: { text: '直发中', pill: 'bg-emerald-100 text-emerald-700', dot: 'bg-emerald-500', num: 'text-emerald-700' },
      false: { text: '手动领取', pill: 'bg-amber-100 text-amber-800', dot: 'bg-amber-500', num: 'text-amber-700' },
      null: { text: '未查询', pill: 'bg-base-200 text-base-content/60', dot: 'bg-base-content/30', num: 'text-base-content/40' },
    }
    const state = (v) => STATE[v == null ? 'null' : v]
    // 卡片主指标: 三种状态统一为 标签 / 数值 / 补充说明 三行, 保证卡片结构一致
    const metric = (m) => {
      // 按接口说明: 没返回数量显示 "—"; extra_msg 不展示
      if (m.canIssue === false) return {
        label: '还需手抽红包', value: m.manualClaim2Go,
        note: m.manualClaim2Go == null ? '' : '手抽完成后切换回直发',
      }
      if (m.canIssue === true) return {
        label: '剩余直发红包', value: m.autoPayoutCnt,
        note: m.manualClaim2Go == null ? '存在指定直发，距下轮目标已超展示门槛' : `距下轮直发目标还需发放 ${m.manualClaim2Go} 个`,
      }
      return { label: '暂无数据', value: null, note: m.lastError ? '查询失败，见下方错误' : '等待首次查询' }
    }
    const count = (c) => c ? list.value.filter(m => m.activityCategory === c).length : list.value.length
    const num = (v) => v == null ? '—' : v

    // insight: 速度 / 预计时间 / 今日统计, 字段缺失时统一显示 "—"
    const ins = (m) => m.insight || {}
    const rate = (m) => {
      const r = ins(m).ratePerHour
      return r == null ? '—' : (r < 10 ? r.toFixed(1) : Math.round(r)) + ' 个/时'
    }
    const etaLabel = (m) => m.canIssue === false ? '预计切回直发' : m.canIssue === true ? '预计用完' : '预计'
    const bjToday = () => new Date(Date.now() + 8 * 3600e3).toISOString().slice(0, 10)
    const eta = (m) => {
      const t = ins(m).etaAt
      if (!t) return '—'
      return t.slice(0, 10) === bjToday() ? t.slice(11, 16) : t.slice(5, 16)
    }
    const directShare = (m) => {
      const i = ins(m)
      return i.todaySec ? Math.round(i.todayDirectSec / i.todaySec * 100) + '%' : '—'
    }

    onMounted(() => { refresh(); timer = setInterval(refresh, 5000) })
    onUnmounted(() => clearInterval(timer))

    return { CATS, list, loaded, error, filter, cur, tabs, filtered, checking, check,
      detail, tab, hours, stats, statsLoading, statsError, chart, openDetail, closeDetail, setTab, setHours,
      logs, logsLoading, dur, since, state, metric, count, num, ins, rate, etaLabel, eta, directShare }
  },
  template: `
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <i data-lucide="gift" class="w-6 h-6"></i> 直发查询
        </h1>
        <span class="text-xs text-base-content/50 flex items-center gap-1.5">
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span> 每 5 秒自动刷新
        </span>
      </div>

      <div v-if="error" class="alert alert-error text-sm">{{ error }}</div>

      <div v-if="loaded && !list.length" class="card bg-base-100 shadow">
        <div class="card-body items-center text-center text-base-content/60">
          <p>还没有直发查询配置</p>
          <router-link to="/issue-config" class="btn btn-primary btn-sm">去添加配置</router-link>
        </div>
      </div>

      <template v-else-if="list.length">
        <div role="tablist" class="tabs tabs-boxed bg-base-100 shadow w-fit max-w-full overflow-x-auto flex-nowrap">
          <a v-for="t in tabs" :key="t" role="tab" class="tab gap-1.5 whitespace-nowrap"
             :class="{'tab-active': cur === t}" @click="filter=t">
            {{ t ? CATS[t] : '全部' }}<span class="text-xs opacity-60">{{ count(t) }}</span>
          </a>
        </div>

        <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          <div v-for="m in filtered" :key="m.id" class="card bg-base-100 shadow" :class="{'opacity-70': !m.enabled}">
            <div class="card-body p-5 gap-4">
              <div class="flex items-start justify-between gap-3">
                <div class="min-w-0">
                  <h2 class="font-semibold text-base truncate" :title="m.name">{{ m.name }}</h2>
                  <p class="text-xs text-base-content/50 mt-0.5 truncate">
                    {{ CATS[m.activityCategory] || m.activityCategory }} · {{ m.appKey }} · 每 {{ m.intervalSec }} 秒<template v-if="!m.enabled"> · 已停用</template>
                  </p>
                </div>
                <div class="shrink-0 flex flex-col items-end gap-1">
                  <span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium" :class="state(m.canIssue).pill">
                    <span class="w-1.5 h-1.5 rounded-full" :class="state(m.canIssue).dot"></span>{{ state(m.canIssue).text }}
                  </span>
                  <span v-if="since(m)" class="text-[11px] text-base-content/50 tabular-nums">已持续 {{ since(m) }}</span>
                </div>
              </div>

              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">{{ metric(m).label }}</div>
                <div class="flex items-baseline gap-1 mt-1.5">
                  <template v-if="metric(m).value != null">
                    <span class="text-4xl font-bold tabular-nums leading-none" :class="state(m.canIssue).num">{{ metric(m).value }}</span>
                    <span class="text-sm text-base-content/60">个</span>
                  </template>
                  <span v-else class="text-4xl font-bold leading-none text-base-content/30">—</span>
                </div>
                <div class="text-xs text-base-content/60 mt-2 truncate" :title="metric(m).note">{{ metric(m).note || '\u00a0' }}</div>
              </div>

              <div class="grid grid-cols-3 gap-2">
                <div class="min-w-0">
                  <div class="text-[11px] text-base-content/50">速度</div>
                  <div class="text-sm font-semibold tabular-nums truncate">{{ rate(m) }}</div>
                </div>
                <div class="min-w-0">
                  <div class="text-[11px] text-base-content/50">{{ etaLabel(m) }}</div>
                  <div class="text-sm font-semibold tabular-nums truncate" :title="ins(m).etaAt || ''">{{ eta(m) }}</div>
                </div>
                <div class="min-w-0" :title="ins(m).reportError || ''">
                  <div class="text-[11px] text-base-content/50">今日发放 / 使用</div>
                  <div class="text-sm font-semibold tabular-nums truncate" :class="{'text-base-content/40': ins(m).reportError}">{{ num(ins(m).todayFund) }} / {{ num(ins(m).todayUse) }}</div>
                </div>
              </div>

              <p v-if="m.lastError" class="text-xs text-error flex gap-1.5 -mt-1" :title="m.lastError">
                <i data-lucide="alert-circle" class="w-3.5 h-3.5 shrink-0 mt-px"></i>
                <span class="line-clamp-2 break-all">{{ m.lastError }}</span>
              </p>

              <div class="mt-auto pt-3 border-t border-base-200 flex items-center justify-between gap-2">
                <span class="text-xs text-base-content/50 truncate">
                  {{ m.checkedAt ? '更新于 ' + m.checkedAt.slice(11) : '尚未查询' }}
                </span>
                <div class="flex gap-1 shrink-0">
                  <button class="btn btn-ghost btn-xs" @click="openDetail(m)">
                    <i data-lucide="activity" class="w-3.5 h-3.5"></i> 详情
                  </button>
                  <button class="btn btn-primary btn-outline btn-xs" :disabled="checking[m.id]" @click="check(m)">
                    <span v-if="checking[m.id]" class="loading loading-spinner loading-xs"></span>
                    <i v-else data-lucide="refresh-cw" class="w-3.5 h-3.5"></i> 查询
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <Modal :show="!!detail" :title="'详情 · ' + (detail?.name || '')" width="max-w-4xl" :show-footer="false" @close="closeDetail">
        <template v-if="detail">
          <div role="tablist" class="tabs tabs-boxed w-fit mb-4">
            <a role="tab" class="tab" :class="{'tab-active': tab === 'trend'}" @click="setTab('trend')">趋势</a>
            <a role="tab" class="tab" :class="{'tab-active': tab === 'today'}" @click="setTab('today')">今日统计 / 直发周期</a>
            <a role="tab" class="tab" :class="{'tab-active': tab === 'logs'}" @click="setTab('logs')">变化记录</a>
          </div>

          <div v-if="tab === 'trend'" class="space-y-3">
            <div class="flex justify-between items-center gap-2">
              <div>
                <h2 class="font-semibold">数量走势</h2>
                <p class="text-xs text-base-content/50">每格取该时段最后一次查询结果</p>
              </div>
              <div class="join">
                <button v-for="h in [6, 24, 72]" :key="h" class="btn btn-xs join-item" :class="{'btn-active': hours === h}" @click="setHours(h)">{{ h }} 小时</button>
              </div>
            </div>
            <div v-if="statsLoading && !stats" class="flex justify-center py-12"><span class="loading loading-spinner"></span></div>
            <p v-else-if="statsError && !stats" class="text-sm text-error py-8 text-center">{{ statsError }}</p>
            <p v-else-if="!chart.labels.length" class="text-sm text-base-content/50 py-8 text-center">该时段暂无记录</p>
            <template v-else>
              <TrendChart :labels="chart.labels" :series="chart.series" :band="chart.band" :titles="chart.titles" :notes="chart.notes"/>
              <p v-if="chart.empty" class="text-xs text-base-content/50 text-center">该时段接口未返回数值</p>
              <div class="flex gap-4 justify-center text-[11px] text-base-content/50">
                <span class="flex items-center gap-1.5"><span class="w-2 h-2 rounded-full bg-emerald-500"></span>直发中</span>
                <span class="flex items-center gap-1.5"><span class="w-2 h-2 rounded-full bg-amber-500"></span>手动领取</span>
                <span class="flex items-center gap-1.5"><span class="w-2 h-2 rounded-full bg-base-300"></span>未知</span>
              </div>
            </template>
          </div>

          <div v-else-if="tab === 'today'" class="space-y-4">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">今日切换次数</div>
                <div class="text-2xl font-bold tabular-nums mt-1">{{ num(ins(detail).todaySwitches) }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">今日直发时长</div>
                <div class="text-2xl font-bold tabular-nums mt-1 text-emerald-700">{{ ins(detail).todayDirectSec == null ? '—' : ins(detail).todayDirectSec ? dur(ins(detail).todayDirectSec) : '0 分钟' }}</div>
                <div class="text-xs text-base-content/50 mt-1">占比 {{ directShare(detail) }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">当前速度</div>
                <div class="text-2xl font-bold tabular-nums mt-1">{{ rate(detail) }}</div>
              </div>
              <div class="rounded-box bg-base-200/70 px-4 py-3">
                <div class="text-xs text-base-content/60">{{ etaLabel(detail) }}</div>
                <div class="text-2xl font-bold tabular-nums mt-1">{{ eta(detail) }}</div>
                <div class="text-xs text-base-content/50 mt-1 tabular-nums" v-if="ins(detail).etaAt">{{ ins(detail).etaAt }}</div>
              </div>
            </div>

            <div>
              <h2 class="font-semibold">直发周期</h2>
              <p class="text-xs text-base-content/50">近 {{ hours }} 小时内的状态区间，最新在前</p>
            </div>
            <div v-if="statsLoading && !stats" class="flex justify-center py-8"><span class="loading loading-spinner"></span></div>
            <div v-else class="overflow-x-auto max-h-[45vh]">
              <table class="table table-sm table-pin-rows">
                <thead><tr><th>状态</th><th>开始</th><th>结束</th><th class="text-right">持续</th></tr></thead>
                <tbody>
                  <tr v-for="r in stats?.rounds || []" :key="r.start">
                    <td>
                      <span class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs" :class="state(r.state).pill">
                        <span class="w-1.5 h-1.5 rounded-full" :class="state(r.state).dot"></span>{{ state(r.state).text }}
                      </span>
                    </td>
                    <td class="whitespace-nowrap tabular-nums">{{ r.start }}</td>
                    <td class="whitespace-nowrap tabular-nums">{{ r.end || '进行中' }}</td>
                    <td class="text-right whitespace-nowrap tabular-nums">{{ dur(r.sec) }}</td>
                  </tr>
                  <tr v-if="!stats?.rounds?.length"><td colspan="4" class="text-center text-base-content/40">暂无记录</td></tr>
                </tbody>
              </table>
            </div>
          </div>

          <template v-else>
            <div v-if="logsLoading" class="flex justify-center py-8"><span class="loading loading-spinner"></span></div>
            <div v-else class="overflow-x-auto max-h-[60vh]">
              <table class="table table-sm table-pin-rows">
                <thead><tr><th>时间</th><th>状态</th><th class="text-right">剩余直发</th><th class="text-right">还需手抽</th><th>邮件</th></tr></thead>
                <tbody>
                  <tr v-for="l in logs" :key="l.id" :class="{'bg-base-200/60': l.statusChanged}">
                    <td class="whitespace-nowrap tabular-nums">{{ l.createdAt }}</td>
                    <td>
                      <span class="inline-flex items-center gap-1.5 rounded-full px-2 py-0.5 text-xs" :class="state(l.canIssue).pill">
                        <span class="w-1.5 h-1.5 rounded-full" :class="state(l.canIssue).dot"></span>{{ state(l.canIssue).text }}
                      </span>
                      <span v-if="l.statusChanged" class="text-xs text-base-content/50 ml-1">状态切换</span>
                    </td>
                    <td class="text-right tabular-nums">{{ num(l.autoPayoutCnt) }}</td>
                    <td class="text-right tabular-nums">{{ num(l.manualClaim2Go) }}</td>
                    <td class="text-xs text-base-content/60">{{ l.mailResult || '' }}</td>
                  </tr>
                  <tr v-if="!logs.length"><td colspan="5" class="text-center text-base-content/40">暂无记录</td></tr>
                </tbody>
              </table>
              <p class="text-xs text-base-content/40 mt-2">"—" 表示接口未返回数值</p>
            </div>
          </template>
        </template>
        <div class="modal-action"><button class="btn btn-sm" @click="closeDetail">关闭</button></div>
      </Modal>
    </div>
  `,
}
