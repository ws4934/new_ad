// 多选下拉组件 — 基于 DaisyUI select 样式改造
// <MultiSelect :options="[{value,label}]" v-model="arr" placeholder="请选择"/>
import { ref, computed } from 'vue'

export default {
  props: {
    modelValue: { type: Array, default: () => [] },
    options: { type: Array, default: () => [] },
    placeholder: { type: String, default: '请选择' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const open = ref(false)

    const selected = computed({
      get: () => props.modelValue,
      set: (v) => emit('update:modelValue', v),
    })

    const available = computed(() =>
      props.options.filter(o => !selected.value.includes(o.value))
    )

    const getLabel = (val) => {
      const o = props.options.find(x => x.value === val)
      return o ? o.label : val
    }

    const add = (val) => { selected.value = [...selected.value, val] }
    const remove = (val) => { selected.value = selected.value.filter(v => v !== val) }

    return { selected, available, getLabel, add, remove, open }
  },
  template: `
    <div class="dropdown w-full" :class="{ 'dropdown-open': open }">
      <div tabindex="0" role="button"
           class="select select-bordered select-sm w-full h-auto min-h-[2rem] flex flex-wrap items-center gap-1 py-1"
           @click="open = !open">
        <span v-for="v in selected" :key="v"
              class="badge badge-primary badge-sm gap-0.5 shrink-0"
              @click.stop>
          {{ getLabel(v) }}
          <span class="cursor-pointer ml-0.5 opacity-70 hover:opacity-100" @click.stop="remove(v)">&times;</span>
        </span>
        <span v-if="!selected.length" class="text-gray-400 text-sm select-none">{{ placeholder }}</span>
      </div>
      <ul v-show="open" tabindex="0"
          class="dropdown-content menu bg-base-100 rounded-box shadow-md border border-base-300 z-50 w-full mt-1 max-h-44 overflow-y-auto p-1 text-sm">
        <li v-for="o in available" :key="o.value" @click="add(o.value); open = false">
          <a>{{ o.label }}</a>
        </li>
        <li v-if="!available.length" class="disabled"><a class="text-gray-400">已全部选择</a></li>
      </ul>
      <div v-if="open" class="fixed inset-0 z-40" @click="open = false"></div>
    </div>
  `,
}
