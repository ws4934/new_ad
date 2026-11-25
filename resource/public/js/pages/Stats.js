import { ref, computed, inject, onMounted } from 'vue'
import { api } from '../api.js?v=16'
import TrendChart from '../components/TrendChart.js?v=16'

// 日期转 YYYY-MM-DD
function fmtDate(d) {
  const p = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}`
}

export default {
  components: { TrendChart },
  setup() {
    const toast = inject('toast')
    const accounts = ref([])
    const rows = ref([])
    const summary = ref(null)
    const loading = ref(false)

    // 默认最近 7 天 (含今天)
    const today = new Date()
    const weekAgo = new Date(today.getTime() - 6 * 86400e3)
    const filter = ref({
      startDate: fmtDate(weekAgo),
      endDate: fmtDate(today),
      accountId: 0,
    })

    const query = async () => {
      loading.value = true
      try {
        const res = await api.statsReport({
          startDate: filter.value.startDate,
          endDate: filter.value.endDate,
          accountId: filter.value.accountId || undefined,
        })
        rows.value = res.rows || []
        summary.value = res.summary
      } catch (e) {
        toast.show('查询失败: ' + e.message, 'error')
      } finally {
        loading.value = false
      }
    }

    // 按日期聚合 (多账户求和), 升序, 供趋势图使用
    const trend = computed(() => {
      const byDate = new Map()
      for (const r of rows.value) {
        let d = byDate.get(r.statDate)
        if (!d) { d = { click: 0, taolian: 0, callback: 0, fake: 0 }; byDate.set(r.statDate, d) }
        d.click += r.clickCount
        d.taolian += r.taolianOk
        d.callback += r.callbackOk
        d.fake += r.fakeOk
      }
      const dates = [...byDate.keys()].sort()
      return {
        labels: dates.map(d => d.slice(5)), // MM-DD
        series: [
          { name: '点击', color: '#6366f1', points: dates.map(d => byDate.get(d).click) },
          { name: '淘联成功', color: '#22c55e', points: dates.map(d => byDate.get(d).taolian) },
          { name: '真实回传', color: '#0ea5e9', points: dates.map(d => byDate.get(d).callback) },
          { name: '虚假回传', color: '#f59e0b', points: dates.map(d => byDate.get(d).fake) },
        ],
      }
    })

    // 客户端导出当前表格为 CSV
    const exportCSV = () => {
      if (!rows.value.length) return
      const header = '日期,账户,平台,点击,淘联成功,淘联失败,淘联回调,回调丢弃,真实回传,回传失败,联调回传,联调失败,虚假回传,虚假失败,点击丢弃'
      const lines = rows.value.map(r => [
        r.statDate, r.accountName || r.accountId, r.platformName,
        r.clickCount, r.taolianOk, r.taolianErr, r.cbRecv, r.cbDrop,
        r.callbackOk, r.callbackErr, r.debugOk, r.debugErr, r.fakeOk, r.fakeErr, r.dropCount,
      ].join(','))
      const blob = new Blob(['﻿' + [header, ...lines].join('\n')], { type: 'text/csv;charset=utf-8' })
      const a = document.createElement('a')
      a.href = URL.createObjectURL(blob)
      a.download = `report_${filter.value.startDate}_${filter.value.endDate}.csv`
      a.click()
      URL.revokeObjectURL(a.href)
    }

    onMounted(async () => {
      try {
        const res = await api.accountList()
        accounts.value = res.list || []
      } catch { /* 账户列表加载失败不阻塞报表 */ }
      await query()
    })

    return { accounts, rows, summary, loading, filter, query, exportCSV, trend }
  },
  template: `
    <div class="space-y-4">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <i data-lucide="line-chart" class="w-6 h-6"></i> 统计报表
      </h1>

      <!-- 筛选条件 -->
      <div class="card bg-base-100 shadow">
        <div class="card-body py-4 flex-row flex-wrap items-end gap-3">
          <div class="form-control">
            <label class="label py-1"><span class="label-text">开始日期</span></label>
            <input type="date" v-model="filter.startDate" class="input input-bordered input-sm">
          </div>
          <div class="form-control">
            <label class="label py-1"><span class="label-text">结束日期</span></label>
            <input type="date" v-model="filter.endDate" class="input input-bordered input-sm">
          </div>
          <div class="form-control">
            <label class="label py-1"><span class="label-text">账户</span></label>
            <select v-model.number="filter.accountId" class="select select-bordered select-sm">
              <option :value="0">全部账户</option>
              <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.name }}</option>
            </select>
          </div>
          <button class="btn btn-primary btn-sm" :disabled="loading" @click="query">
            <span v-if="loading" class="loading loading-spinner loading-xs"></span>
            查询
          </button>
          <button class="btn btn-ghost btn-sm" :disabled="!rows.length" @click="exportCSV">
            <i data-lucide="download" class="w-4 h-4"></i> 导出 CSV
          </button>
        </div>
      </div>

      <!-- 区间汇总 -->
      <div class="grid grid-cols-4 md:grid-cols-8 gap-3" v-if="summary">
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">点击</div>
          <div class="stat-value text-xl">{{ summary.clickCount }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">淘联上报</div>
          <div class="stat-value text-xl text-success">{{ summary.taolianOk }}</div>
          <div class="stat-desc" v-if="summary.taolianErr">失败 {{ summary.taolianErr }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">淘联回调</div>
          <div class="stat-value text-xl">{{ summary.cbRecv }}</div>
          <div class="stat-desc" v-if="summary.cbDrop">丢弃 {{ summary.cbDrop }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">真实回传</div>
          <div class="stat-value text-xl text-info">{{ summary.callbackOk }}</div>
          <div class="stat-desc" v-if="summary.callbackErr">失败 {{ summary.callbackErr }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">联调回传</div>
          <div class="stat-value text-xl">{{ summary.debugOk }}</div>
          <div class="stat-desc" v-if="summary.debugErr">失败 {{ summary.debugErr }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">虚假回传</div>
          <div class="stat-value text-xl text-warning">{{ summary.fakeOk }}</div>
          <div class="stat-desc" v-if="summary.fakeErr">失败 {{ summary.fakeErr }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">点击丢弃</div>
          <div class="stat-value text-xl text-error">{{ summary.dropCount }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box py-3">
          <div class="stat-title text-xs">转化率</div>
          <div class="stat-value text-xl">{{ summary.clickCount ? ((summary.callbackOk + summary.fakeOk) / summary.clickCount * 100).toFixed(2) + '%' : '-' }}</div>
          <div class="stat-desc">回传 / 点击</div>
        </div>
      </div>

      <!-- 趋势图 (按日聚合) -->
      <div class="card bg-base-100 shadow" v-if="trend.labels.length">
        <div class="card-body py-4">
          <h2 class="card-title text-base"><i data-lucide="trending-up" class="w-4 h-4"></i> 每日趋势</h2>
          <TrendChart :labels="trend.labels" :series="trend.series"/>
        </div>
      </div>

      <!-- 明细表 -->
      <div class="card bg-base-100 shadow">
        <div class="card-body p-0">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>日期</th><th>账户</th><th>平台</th>
                <th class="text-right">点击</th>
                <th class="text-right">淘联成功</th><th class="text-right">淘联失败</th>
                <th class="text-right">淘联回调</th><th class="text-right">回调丢弃</th>
                <th class="text-right">真实回传</th><th class="text-right">回传失败</th>
                <th class="text-right">联调回传</th>
                <th class="text-right">虚假回传</th><th class="text-right">虚假失败</th>
                <th class="text-right">点击丢弃</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in rows" :key="r.statDate + '_' + r.accountId">
                <td class="whitespace-nowrap">
                  {{ r.statDate }}
                  <span v-if="r.isToday" class="badge badge-primary badge-xs ml-1">实时</span>
                </td>
                <td>{{ r.accountName || r.accountId }}</td>
                <td>{{ r.platformName }}</td>
                <td class="text-right">{{ r.clickCount }}</td>
                <td class="text-right text-success">{{ r.taolianOk }}</td>
                <td class="text-right" :class="r.taolianErr ? 'text-error' : 'opacity-40'">{{ r.taolianErr }}</td>
                <td class="text-right" :class="r.cbRecv ? '' : 'opacity-40'">{{ r.cbRecv }}</td>
                <td class="text-right" :class="r.cbDrop ? 'text-error' : 'opacity-40'">{{ r.cbDrop }}</td>
                <td class="text-right text-info">{{ r.callbackOk }}</td>
                <td class="text-right" :class="r.callbackErr ? 'text-error' : 'opacity-40'">{{ r.callbackErr }}</td>
                <td class="text-right" :class="r.debugOk ? '' : 'opacity-40'">{{ r.debugOk }}</td>
                <td class="text-right text-warning">{{ r.fakeOk }}</td>
                <td class="text-right" :class="r.fakeErr ? 'text-error' : 'opacity-40'">{{ r.fakeErr }}</td>
                <td class="text-right" :class="r.dropCount ? 'text-error' : 'opacity-40'">{{ r.dropCount }}</td>
              </tr>
              <tr v-if="!rows.length && !loading">
                <td colspan="14" class="text-center py-8 opacity-50">所选区间内无数据</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `,
}
