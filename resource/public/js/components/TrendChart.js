import { computed } from 'vue'

// 零依赖 SVG 折线趋势图
// props:
//   labels: ['06-01', '06-02', ...]            x 轴标签
//   series: [{ name, color, points: [..] }]    多条折线
export default {
  props: {
    labels: { type: Array, default: () => [] },
    series: { type: Array, default: () => [] },
  },
  setup(props) {
    const W = 760, H = 210
    const padL = 44, padR = 12, padT = 12, padB = 26

    const maxVal = computed(() => {
      let m = 0
      for (const s of props.series) for (const v of s.points) if (v > m) m = v
      return m || 1
    })

    const x = (i) => {
      const n = props.labels.length
      if (n <= 1) return padL + (W - padL - padR) / 2
      return padL + (i * (W - padL - padR)) / (n - 1)
    }
    const y = (v) => padT + (1 - v / maxVal.value) * (H - padT - padB)

    // 每条线的 polyline points + 数据点坐标
    const lines = computed(() => props.series.map(s => ({
      name: s.name,
      color: s.color,
      path: s.points.map((v, i) => `${x(i).toFixed(1)},${y(v).toFixed(1)}`).join(' '),
      dots: s.points.map((v, i) => ({ cx: x(i), cy: y(v), v })),
    })))

    // y 轴 4 条刻度线 (0 / 1/3 / 2/3 / max)
    const yTicks = computed(() => [0, 1, 2, 3].map(i => {
      const v = Math.round((maxVal.value * i) / 3)
      return { v, y: y(v) }
    }))

    // x 轴标签抽稀: 最多显示 ~8 个
    const xTicks = computed(() => {
      const n = props.labels.length
      const step = Math.max(1, Math.ceil(n / 8))
      const out = []
      for (let i = 0; i < n; i += step) out.push({ label: props.labels[i], x: x(i) })
      return out
    })

    return { W, H, padL, padR, lines, yTicks, xTicks }
  },
  template: `
    <div>
      <svg :viewBox="'0 0 ' + W + ' ' + H" class="w-full select-none">
        <!-- 网格 + y 轴刻度 -->
        <g v-for="t in yTicks" :key="t.v">
          <line :x1="padL" :x2="W - padR" :y1="t.y" :y2="t.y"
            stroke="currentColor" stroke-opacity="0.08"/>
          <text :x="padL - 6" :y="t.y + 3" text-anchor="end"
            font-size="10" fill="currentColor" fill-opacity="0.5">{{ t.v }}</text>
        </g>
        <!-- x 轴标签 -->
        <text v-for="t in xTicks" :key="t.label" :x="t.x" :y="H - 8"
          text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.5">{{ t.label }}</text>
        <!-- 折线 + 数据点 -->
        <g v-for="l in lines" :key="l.name">
          <polyline :points="l.path" fill="none" :stroke="l.color" stroke-width="2"
            stroke-linejoin="round" stroke-linecap="round"/>
          <circle v-for="(d, i) in l.dots" :key="i" :cx="d.cx" :cy="d.cy" r="2.5" :fill="l.color">
            <title>{{ l.name }}: {{ d.v }}</title>
          </circle>
        </g>
      </svg>
      <!-- 图例 -->
      <div class="flex gap-4 justify-center text-xs opacity-70">
        <span v-for="l in lines" :key="l.name" class="flex items-center gap-1">
          <span class="inline-block w-3 h-0.5 rounded" :style="{ background: l.color }"></span>{{ l.name }}
        </span>
      </div>
    </div>
  `,
}
