import { ref, computed, inject, onMounted } from 'vue'
import { api } from '../api.js?v=6'

export default {
  props: ['id'],
  setup(props) {
    const account = ref(null)
    const mappings = ref([])
    const options = ref({ taolianTypes: [], platformTypes: {} })
    const showAdd = ref(false)
    const addForm = ref({ taolianType: null, platformType: '' })
    const toast = inject('toast')

    const platformTypes = computed(() => {
      if (!account.value) return []
      return options.value.platformTypes[account.value.platformName] || []
    })

    const refresh = async () => {
      const [accts, maps, opts] = await Promise.all([
        api.accountList(),
        api.mappingList(props.id),
        api.mappingOptions(),
      ])
      account.value = (accts.list || []).find(a => a.id === Number(props.id))
      mappings.value = maps.list || []
      options.value = opts
    }

    const addMapping = async () => {
      try {
        const tt = options.value.taolianTypes.find(t => t.type === addForm.value.taolianType)
        const pt = platformTypes.value.find(t => t.type === addForm.value.platformType)
        const items = [...mappings.value.map(m => ({
          taolianType: m.taolianType, taolianDesc: m.taolianDesc,
          platformType: m.platformType, platformDesc: m.platformDesc, enabled: m.enabled,
        })), {
          taolianType: addForm.value.taolianType,
          taolianDesc: tt?.desc || '',
          platformType: addForm.value.platformType,
          platformDesc: pt?.desc || '',
          enabled: true,
        }]
        await api.mappingSave({ accountId: Number(props.id), items })
        showAdd.value = false
        addForm.value = { taolianType: null, platformType: '' }
        toast.show('映射已添加', 'success')
        await refresh()
      } catch (e) { toast.show('失败: ' + e.message, 'error') }
    }

    const delMapping = async (m) => {
      if (!confirm('确认删除?')) return
      try {
        await api.mappingDelete(m.id)
        toast.show('已删除', 'success')
        await refresh()
      } catch (e) { toast.show('删除失败: ' + e.message, 'error') }
    }

    onMounted(refresh)
    return { account, mappings, options, platformTypes, showAdd, addForm, addMapping, delMapping }
  },
  template: `
    <div class="space-y-4" v-if="account">
      <div class="flex items-center gap-3">
        <router-link to="/accounts" class="btn btn-ghost btn-sm">
          <i data-lucide="chevron-left" class="w-4 h-4"></i> 返回
        </router-link>
        <h1 class="text-2xl font-bold">{{ account.name }}</h1>
        <span class="badge">{{ account.platformName }}</span>
      </div>

      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <div class="flex justify-between items-center">
            <h2 class="card-title">行为映射</h2>
            <button class="btn btn-sm btn-primary" @click="showAdd=true">
              <i data-lucide="plus" class="w-4 h-4"></i> 新增映射
            </button>
          </div>
          <table class="table mt-2">
            <thead>
              <tr><th>淘联 transformType</th><th></th><th>广告平台 action</th><th>启用</th><th></th></tr>
            </thead>
            <tbody>
              <tr v-for="m in mappings" :key="m.id">
                <td>{{ m.taolianType }} - {{ m.taolianDesc }}</td>
                <td><i data-lucide="arrow-right" class="w-4 h-4"></i></td>
                <td>{{ m.platformType }} - {{ m.platformDesc }}</td>
                <td><span :class="m.enabled?'badge badge-primary badge-sm':'badge badge-ghost badge-sm'">{{ m.enabled?'启用':'禁用' }}</span></td>
                <td><button class="btn btn-xs btn-ghost text-error" @click="delMapping(m)"><i data-lucide="trash-2" class="w-3.5 h-3.5"></i></button></td>
              </tr>
              <tr v-if="!mappings.length"><td colspan="5" class="text-center text-gray-400">暂无映射, 请新增</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <Modal :show="showAdd" title="新增行为映射"
        :confirm-disabled="!addForm.taolianType||!addForm.platformType"
        @close="showAdd=false" @confirm="addMapping">
        <div class="flex gap-3 items-center">
          <select v-model="addForm.taolianType" class="select select-bordered select-sm flex-1">
            <option :value="null" disabled>淘联类型</option>
            <option v-for="t in options.taolianTypes" :key="t.type" :value="t.type">{{ t.type }} - {{ t.desc }}</option>
          </select>
          <i data-lucide="arrow-right" class="w-5 h-5 shrink-0"></i>
          <select v-model="addForm.platformType" class="select select-bordered select-sm flex-1">
            <option value="" disabled>广告平台类型</option>
            <option v-for="p in platformTypes" :key="p.type" :value="p.type">{{ p.type }} - {{ p.desc }}</option>
          </select>
        </div>
      </Modal>
    </div>
    <div v-else class="flex justify-center py-20"><span class="loading loading-spinner loading-lg"></span></div>
  `,
}
