import { ref, onMounted, onUnmounted } from 'vue'
import { api } from '../api.js?v=21'

export default {
  setup() {
    const overview = ref(null)
    const qps = ref(null)
    const error = ref('')
    let timer = null

    const refresh = async () => {
      try {
        const [o, q] = await Promise.all([api.overview(), api.qps()])
        overview.value = o
        qps.value = q
        error.value = ''
      } catch (e) {
        error.value = e.message
      }
    }

    onMounted(() => { refresh(); timer = setInterval(refresh, 3000) })
    onUnmounted(() => clearInterval(timer))

    return { overview, qps, error }
  },
  template: `
    <div class="space-y-6">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <i data-lucide="home" class="w-6 h-6"></i> 仪表盘
      </h1>

      <div v-if="error" class="alert alert-error">{{ error }}</div>

      <!-- QPS -->
      <div class="card bg-base-100 shadow" v-if="qps">
        <div class="card-body">
          <h2 class="card-title"><i data-lucide="line-chart" class="w-5 h-5"></i> 全局 QPS</h2>
          <div class="grid grid-cols-4 gap-4">
            <div class="stat place-items-center">
              <div class="stat-title">当前</div>
              <div class="stat-value text-primary">{{ qps.current }}</div>
              <div class="stat-desc">最近 1 秒</div>
            </div>
            <div class="stat place-items-center">
              <div class="stat-title">10s 均值</div>
              <div class="stat-value">{{ qps.avg10s?.toFixed(1) }}</div>
            </div>
            <div class="stat place-items-center">
              <div class="stat-title">60s 均值</div>
              <div class="stat-value">{{ qps.avg60s?.toFixed(1) }}</div>
            </div>
            <div class="stat place-items-center">
              <div class="stat-title">5min 峰值</div>
              <div class="stat-value text-warning">{{ qps.peak5m }}</div>
            </div>
          </div>
        </div>
      </div>

      <!-- 今日汇总 -->
      <div class="grid grid-cols-4 gap-4" v-if="overview">
        <div class="stat bg-base-100 shadow rounded-box">
          <div class="stat-title">今日点击</div>
          <div class="stat-value">{{ overview.totalClick || 0 }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box">
          <div class="stat-title">淘联上报成功</div>
          <div class="stat-value text-success">{{ overview.totalTaolian || 0 }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box">
          <div class="stat-title">真实回传</div>
          <div class="stat-value text-info">{{ overview.totalCallback || 0 }}</div>
        </div>
        <div class="stat bg-base-100 shadow rounded-box">
          <div class="stat-title">虚假回传</div>
          <div class="stat-value text-warning">{{ overview.totalFake || 0 }}</div>
        </div>
      </div>
    </div>
  `,
}
