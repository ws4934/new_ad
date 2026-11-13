import { computed } from 'vue'

// 零依赖 SVG 折线趋势图
// props:
//   labels: ['06-01', '06-02', ...]            x 轴标签
//   series: [{ name, color, points: [..] }]    多条折线, null 点断开
//   band:   ['#hex', ...] (可选)               图下方状态条, 每个点一种颜色, 与 x 轴对齐
export default {
  props: {
    labels: { type: Array, default: () => [] },
    series: { type: Array, default: () => [] },
    band: { type: Array, default: null },
  },
  setup(props) {
    const W = 760, H = 210
    const padL = 44, padR = 12, padT = 12, padB = 26

    const maxVal = computed(() => {
      let m = 0
      for (const s of props.series) for (const v of s.points) if (v != null && v > m) m = v
      return Math.max(m, 3) // 至少 3, 避免 4 条刻度取整后重复
    })

    const x = (i) => {
      const n = props.labels.length
      if (n <= 1) return padL + (W - padL - padR) / 2
      return padL + (i * (W - padL - padR)) / (n - 1)
    }
    const y = (v) => padT + (1 - v / maxVal.value) * (H - padT - padB)

    // 每条线按 null 切成多段 polyline; 点多时不画数据点, 只保留孤立点
    const lines = computed(() => props.series.map(s => {
      const segs = []
      let cur = null
      s.points.forEach((v, i) => {
        if (v == null) { cur = null; return }
        if (!cur) segs.push(cur = [])
        cur.push({ cx: x(i), cy: y(v), v })
      })
      const dense = s.points.length > 60
      return {
        name: s.name,
        color: s.color,
        paths: segs.filter(g => g.length > 1).map(g => g.map(d => `${d.cx.toFixed(1)},${d.cy.toFixed(1)}`).join(' ')),
        dots: segs.flatMap(g => (dense && g.length > 1) ? [] : g),
      }
    }))

    // 状态条: 相邻同色合并, 每段从前一点中点到后一点中点
    const bandRects = computed(() => {
      const b = props.band
      if (!b || !b.length) return []
      const half = b.length > 1 ? (x(1) - x(0)) / 2 : (W - padL - padR) / 2
      const out = []
      b.forEach((c, i) => {
        const x1 = Math.max(padL, x(i) - half), x2 = Math.min(W - padR, x(i) + half)
        const last = out[out.length - 1]
        if (last && last.c === c) last.w = x2 - last.x
        else out.push({ c, x: x1, w: x2 - x1 })
      })
      return out
    })

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

    return { W, H, padL, padR, lines, bandRects, yTicks, xTicks }
  },
  template: `
    <div>
      <svg :viewBox="'0 0 ' + W + ' ' + H" class="w-full select-none">
        <!-- 网格 + y 轴刻度 -->
        <g v-for="(t, i) in yTicks" :key="i">
          <line :x1="padL" :x2="W - padR" :y1="t.y" :y2="t.y"
            stroke="currentColor" stroke-opacity="0.08"/>
          <text :x="padL - 6" :y="t.y + 3" text-anchor="end"
            font-size="10" fill="currentColor" fill-opacity="0.5">{{ t.v }}</text>
        </g>
        <!-- x 轴标签 -->
        <text v-for="(t, i) in xTicks" :key="i" :x="t.x" :y="H - 8"
          text-anchor="middle" font-size="10" fill="currentColor" fill-opacity="0.5">{{ t.label }}</text>
        <!-- 折线 + 数据点 -->
        <g v-for="l in lines" :key="l.name">
          <polyline v-for="(p, i) in l.paths" :key="i" :points="p" fill="none" :stroke="l.color" stroke-width="2"
            stroke-linejoin="round" stroke-linecap="round"/>
          <circle v-for="(d, i) in l.dots" :key="'d' + i" :cx="d.cx" :cy="d.cy" r="2.5" :fill="l.color">
            <title>{{ l.name }}: {{ d.v }}</title>
          </circle>
        </g>
      </svg>
      <!-- 状态条 (可选), 与折线共用 x 坐标 -->
      <svg v-if="bandRects.length" :viewBox="'0 0 ' + W + ' 6'" preserveAspectRatio="none" class="w-full h-1.5 block mb-2">
        <rect v-for="(r, i) in bandRects" :key="i" :x="r.x" y="0" :width="r.w" height="6" rx="3" :style="{ fill: r.c }"/>
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
