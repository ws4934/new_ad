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

    // 后端时间不带时区: 用同源的 checkedAt - changedAt 计算持续时长, 不受浏览器时区影响
    const parse = (t) => new Date(t.replace(' ', 'T')).getTime()
    const since = (m) => {
      if (!m.changedAt || !m.checkedAt) return ''
      const sec = Math.max(0, Math.floor((parse(m.checkedAt) - parse(m.changedAt)) / 1000))
      if (isNaN(sec)) return ''
      const d = Math.floor(sec / 86400), h = Math.floor(sec % 86400 / 3600), mi = Math.floor(sec % 3600 / 60)
      if (d) return `${d} 天 ${h} 小时`
      if (h) return `${h} 小时 ${mi} 分`
      return mi ? `${mi} 分钟` : '不足 1 分钟'
    }
    const STATE = {
      true: { text: '直发中', pill: 'bg-emerald-100 text-emerald-700', dot: 'bg-emerald-500', num: 'text-emerald-700' },
      false: { text: '手动领取', pill: 'bg-amber-100 text-amber-800', dot: 'bg-amber-500', num: 'text-amber-700' },
      null: { text: '未查询', pill: 'bg-base-200 text-base-content/60', dot: 'bg-base-content/30', num: 'text-base-content/40' },
    }
    const state = (v) => STATE[v == null ? 'null' : v]
    // 卡片主指标: 三种状态统一为 标签 / 数值 / 补充说明 三行, 保证卡片结构一致
    const metric = (m) => {
      if (m.canIssue === false) return {
        label: '还需发放红包达标', value: m.manualClaim2Go, empty: '超过展示门槛',
        note: m.extraMsg || '达标后自动切换为直发',
      }
      if (m.canIssue === true) return {
        label: '剩余可直发红包', value: m.autoPayoutCnt, empty: '充足',
        note: m.manualClaim2Go == null ? '距下轮直发目标：超过展示门槛' : `距下轮直发目标还需 ${m.manualClaim2Go} 个`,
      }
      return { label: '暂无数据', value: null, empty: '—', note: m.lastError ? '查询失败，见下方错误' : '等待首次查询' }
    }
    const count = (c) => c ? list.value.filter(m => m.activityCategory === c).length : list.value.length
    const num = (v) => v == null ? '—' : v

    onMounted(() => { refresh(); timer = setInterval(refresh, 5000) })
    onUnmounted(() => clearInterval(timer))

    return { CATS, list, loaded, error, filter, cur, tabs, filtered, checking, check,
      logTarget, logs, logsLoading, openLogs, since, state, metric, count, num }
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
        <div role="tablist" class="tabs tabs-boxed bg-base-100 shadow w-fit">
          <a v-for="t in tabs" :key="t" role="tab" class="tab gap-1.5"
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
                  <span v-else class="text-2xl font-semibold leading-9" :class="state(m.canIssue).num">{{ metric(m).empty }}</span>
                </div>
                <div class="text-xs text-base-content/60 mt-2 truncate" :title="metric(m).note">{{ metric(m).note }}</div>
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
                  <button class="btn btn-ghost btn-xs" @click="openLogs(m)">
                    <i data-lucide="history" class="w-3.5 h-3.5"></i> 记录
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

      <Modal :show="!!logTarget" :title="'变化记录 · ' + (logTarget?.name || '')" width="max-w-3xl" :show-footer="false" @close="logTarget=null">
        <div v-if="logsLoading" class="flex justify-center py-8"><span class="loading loading-spinner"></span></div>
        <div v-else class="overflow-x-auto max-h-[60vh]">
          <table class="table table-sm table-pin-rows">
            <thead><tr><th>时间</th><th>状态</th><th class="text-right">剩余可直发</th><th class="text-right">还需发放</th><th>邮件</th></tr></thead>
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
          <p class="text-xs text-base-content/40 mt-2">"—" 表示接口未返回数值（超过展示门槛或不适用）</p>
        </div>
        <div class="modal-action"><button class="btn btn-sm" @click="logTarget=null">关闭</button></div>
      </Modal>
    </div>
  `,
}
