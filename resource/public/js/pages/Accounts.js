import { ref, computed, inject, onMounted } from 'vue'
import { api } from '../api.js?v=10'

export default {
  setup() {
    const deleteTarget = ref(null)
    const deleting = ref(false)
    const list = ref([])
    const platforms = ref([])
    const options = ref({ taolianTypes: [], platformTypes: {} })
    const showCreate = ref(false)
    const showEdit = ref(false)
    const toast = inject('toast')

    const createForm = ref({
      platformId: 0, name: '',
      company: '', advertiserId: '', platform: 'taobao',
      channel: '', taskId: '', tbkId: '',
      fakeEnabled: false, fakeRate: 20, fakeDelayMin: 30, fakeDelayMax: 1800,
      debugMode: false, debugEvents: [],
      mappings: [],
    })

    const editForm = ref({
      id: 0, platformName: '', name: '',
      company: '', advertiserId: '', platform: '',
      channel: '', taskId: '', tbkId: '',
      fakeEnabled: false, fakeRate: 0,
      fakeDelayMin: 30, fakeDelayMax: 1800, status: 1,
      debugMode: false, debugEvents: [],
      mappings: [],
    })

    // 当前选择的平台名 (用于动态切换凭证字段和映射枚举)
    const createPlatformName = computed(() => {
      const p = platforms.value.find(x => x.id === createForm.value.platformId)
      return p ? p.name : ''
    })
    const createPlatformTypes = computed(() => {
      return options.value.platformTypes[createPlatformName.value] || []
    })
    const editPlatformTypes = computed(() => {
      return options.value.platformTypes[editForm.value.platformName] || []
    })

    const refresh = async () => {
      const [accts, plats, opts] = await Promise.all([
        api.accountList(), api.platformList(), api.mappingOptions(),
      ])
      list.value = accts.list || []
      platforms.value = plats.list || []
      options.value = opts
      if (platforms.value.length && !createForm.value.platformId) {
        createForm.value.platformId = platforms.value[0].id
      }
    }

    const parseJson = (s) => { try { return JSON.parse(s) } catch { return {} } }

    // --- 映射行操作 ---
    const addMappingRow = (form) => {
      form.mappings.push({ taolianType: null, platformType: '', enabled: true })
    }
    const removeMappingRow = (form, idx) => {
      form.mappings.splice(idx, 1)
    }
    const getTaolianDesc = (type) => {
      const t = options.value.taolianTypes.find(x => x.type === type)
      return t ? t.desc : ''
    }
    const getPlatformDesc = (name, type) => {
      const pts = options.value.platformTypes[name] || []
      const t = pts.find(x => x.type === type)
      return t ? t.desc : ''
    }

    // --- 新增 ---
    const create = async () => {
      try {
        const f = createForm.value
        // 1. 创建账户
        const res = await api.accountCreate({
          platformId: f.platformId,
          name: f.name,
          credentials: JSON.stringify({ company: f.company, advertiser_id: f.advertiserId, platform: f.platform }),
          taolianConfig: JSON.stringify({ channel: f.channel, taskId: f.taskId, tbkId: f.tbkId, sourcePack: 1 }),
          fakeEnabled: f.fakeEnabled,
          fakeRate: f.fakeRate,
          fakeDelayMin: f.fakeDelayMin,
          fakeDelayMax: f.fakeDelayMax,
          debugMode: f.debugMode,
          debugEvents: JSON.stringify(f.debugEvents),
        })
        // 2. 保存映射
        const p = platforms.value.find(x => x.id === f.platformId)
        const pName = p ? p.name : ''
        if (f.mappings.length > 0) {
          const items = f.mappings.filter(m => m.taolianType && m.platformType).map(m => ({
            taolianType: m.taolianType,
            taolianDesc: getTaolianDesc(m.taolianType),
            platformType: m.platformType,
            platformDesc: getPlatformDesc(pName, m.platformType),
            enabled: m.enabled,
          }))
          if (items.length > 0) {
            await api.mappingSave({ accountId: res.id, items })
          }
        }
        showCreate.value = false
        toast.show('创建成功', 'success')
        await refresh()
      } catch (e) { toast.show('失败: ' + e.message, 'error') }
    }

    // --- 编辑 ---
    const openEdit = async (a) => {
      const creds = parseJson(a.credentials)
      const tc = parseJson(a.taolianConfig)
      // 加载该账户已有映射
      let maps = []
      try {
        const res = await api.mappingList(a.id)
        maps = (res.list || []).map(m => ({
          taolianType: m.taolianType,
          platformType: m.platformType,
          enabled: m.enabled,
        }))
      } catch {}

      editForm.value = {
        id: a.id, platformName: a.platformName, name: a.name,
        company: creds.company || '', advertiserId: creds.advertiser_id || '', platform: creds.platform || '',
        channel: tc.channel || '', taskId: tc.taskId || '', tbkId: tc.tbkId || '',
        fakeEnabled: a.fakeEnabled, fakeRate: a.fakeRate,
        fakeDelayMin: a.fakeDelayMin, fakeDelayMax: a.fakeDelayMax,
        debugMode: a.debugMode,
        debugEvents: (() => { try { return JSON.parse(a.debugEvents) } catch { return [] } })(),
        status: a.status,
        mappings: maps,
      }
      showEdit.value = true
    }

    const saveEdit = async () => {
      try {
        const f = editForm.value
        // 1. 更新账户
        await api.accountUpdate(f.id, {
          name: f.name,
          credentials: JSON.stringify({ company: f.company, advertiser_id: f.advertiserId, platform: f.platform }),
          taolianConfig: JSON.stringify({ channel: f.channel, taskId: f.taskId, tbkId: f.tbkId, sourcePack: 1 }),
          fakeEnabled: f.fakeEnabled,
          fakeRate: f.fakeRate,
          fakeDelayMin: f.fakeDelayMin,
          fakeDelayMax: f.fakeDelayMax,
          debugMode: f.debugMode,
          debugEvents: JSON.stringify(f.debugEvents),
          status: f.status,
        })
        // 2. 保存映射 (先删后插)
        const items = f.mappings.filter(m => m.taolianType && m.platformType).map(m => ({
          taolianType: m.taolianType,
          taolianDesc: getTaolianDesc(m.taolianType),
          platformType: m.platformType,
          platformDesc: getPlatformDesc(f.platformName, m.platformType),
          enabled: m.enabled,
        }))
        await api.mappingSave({ accountId: f.id, items })

        showEdit.value = false
        toast.show('保存成功', 'success')
        await refresh()
      } catch (e) { toast.show('失败: ' + e.message, 'error') }
    }

    const toggleDebug = async (a) => {
      try {
        await api.accountUpdate(a.id, { debugMode: !a.debugMode })
        toast.show(a.debugMode ? '已关闭联调' : '已开启联调', 'success')
        await refresh()
      } catch (e) { toast.show('操作失败: ' + e.message, 'error') }
    }

    const copyURL = async (id) => {
      try {
        const res = await api.accountMonitorURL(id)
        await navigator.clipboard.writeText(res.url)
        toast.show('监测链接已复制', 'success')
      } catch (e) { toast.show('复制失败: ' + e.message, 'error') }
    }

    const remove = async () => {
      if (!deleteTarget.value || deleting.value) return
      deleting.value = true
      try {
        await api.accountDelete(deleteTarget.value.id)
      } catch (e) {
        toast.show('删除失败: ' + e.message, 'error')
        deleting.value = false
        return
      }
      deleteTarget.value = null
      deleting.value = false
      toast.show('删除成功', 'success')
      try { await refresh() } catch (e) { toast.show('删除成功，但列表刷新失败: ' + e.message, 'error') }
    }

    onMounted(refresh)
    return {
      deleteTarget, deleting, remove, list, platforms, options,
      showCreate, createForm, createPlatformName, createPlatformTypes, create,
      showEdit, editForm, editPlatformTypes, openEdit, saveEdit,
      addMappingRow, removeMappingRow, toggleDebug, copyURL,
    }
  },
  template: `
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <i data-lucide="users" class="w-6 h-6"></i> 账户管理
        </h1>
        <button class="btn btn-primary btn-sm" @click="showCreate=true">
          <i data-lucide="plus" class="w-4 h-4"></i> 新增账户
        </button>
      </div>

      <Modal :show="!!deleteTarget" title="删除账户" :confirm-text="deleting ? '删除中…' : '确认删除'" :confirm-disabled="deleting" @close="!deleting && (deleteTarget=null)" @confirm="remove">
        <p>确认删除账户「{{ deleteTarget?.name }}」？此操作不可恢复。</p>
      </Modal>

      <div class="card bg-base-100 shadow">
        <div class="card-body p-0">
          <table class="table">
            <thead>
              <tr>
                <th>ID</th><th>平台</th><th>账户名</th>
                <th>今日点击</th><th>淘联上报</th><th>回传</th><th>虚假</th>
                <th>虚假上报</th><th>联调</th><th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="a in list" :key="a.id">
                <td>{{ a.id }}</td>
                <td><span class="badge badge-sm">{{ a.platformName }}</span></td>
                <td>{{ a.name }}</td>
                <td>{{ a.todayClick }}</td>
                <td class="text-success">{{ a.todayTaolian }}</td>
                <td class="text-info">{{ a.todayCallback }}</td>
                <td class="text-warning">{{ a.todayFake }}</td>
                <td>
                  <span v-if="a.fakeEnabled" class="badge badge-primary badge-sm">{{ a.fakeRate }}%</span>
                  <span v-else class="badge badge-ghost badge-sm">关闭</span>
                </td>
                <td>
                  <input type="checkbox" class="toggle toggle-primary toggle-xs" :checked="a.debugMode" @change="toggleDebug(a)">
                </td>
                <td class="flex gap-1">
                  <button class="btn btn-xs btn-ghost" @click="copyURL(a.id)" title="复制监测链接">
                    <i data-lucide="copy" class="w-3.5 h-3.5"></i>
                  </button>
                  <button class="btn btn-xs btn-ghost" @click="openEdit(a)" title="编辑">
                    <i data-lucide="edit" class="w-3.5 h-3.5"></i>
                  </button>
                  <button class="btn btn-xs btn-ghost text-error" @click="deleteTarget=a" :disabled="deleting" title="删除账户">
                    <i data-lucide="trash-2" class="w-3.5 h-3.5"></i> 删除
                  </button>
                </td>
              </tr>
              <tr v-if="!list.length"><td colspan="10" class="text-center text-gray-400">暂无账户</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- 新增账户 -->
      <Modal :show="showCreate" title="新增账户" width="max-w-3xl" @close="showCreate=false" @confirm="create">
        <div class="space-y-3">
          <div class="grid grid-cols-2 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">平台</span></label>
              <select v-model="createForm.platformId" class="select select-bordered select-sm">
                <option v-for="p in platforms" :key="p.id" :value="p.id">{{ p.displayName }}</option>
              </select>
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">账户名</span></label>
              <input v-model="createForm.name" class="input input-bordered input-sm">
            </div>
          </div>

          <div class="divider text-xs my-4">账户凭证</div>
          <!-- 微博: 只需公司代码 -->
          <div v-if="createPlatformName==='weibo'" class="grid grid-cols-1 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">公司代码</span></label>
              <input v-model="createForm.company" class="input input-bordered input-sm" placeholder="联调后台分配的 company">
            </div>
          </div>
          <!-- 小红书: 只需应用标识 (advertiser_id 从监测链接宏 __ADVERTISER_ID__ 自动获取) -->
          <div v-else-if="createPlatformName==='xiaohongshu'" class="grid grid-cols-1 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">应用标识</span></label>
              <input v-model="createForm.platform" class="input input-bordered input-sm" placeholder="应用英文名, 如 taobao">
            </div>
          </div>
          <!-- 其他平台: 显示全部字段 -->
          <div v-else class="grid grid-cols-3 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">公司代码</span></label>
              <input v-model="createForm.company" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">广告主ID</span></label>
              <input v-model="createForm.advertiserId" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">应用标识</span></label>
              <input v-model="createForm.platform" class="input input-bordered input-sm">
            </div>
          </div>

          <div class="divider text-xs my-4">淘联配置</div>
          <div class="grid grid-cols-3 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">渠道ID</span></label>
              <input v-model="createForm.channel" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">任务ID</span></label>
              <input v-model="createForm.taskId" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">推广PID</span></label>
              <input v-model="createForm.tbkId" class="input input-bordered input-sm" placeholder="mm_xx_xx_xx">
            </div>
          </div>

          <div class="divider text-xs my-4">行为映射</div>
          <div class="space-y-2">
            <div v-for="(m, idx) in createForm.mappings" :key="idx" class="flex gap-2 items-center">
              <select v-model="m.taolianType" class="select select-bordered select-sm flex-1">
                <option :value="null" disabled>淘联类型</option>
                <option v-for="t in options.taolianTypes" :key="t.type" :value="t.type">{{ t.type }} - {{ t.desc }}</option>
              </select>
              <i data-lucide="arrow-right" class="w-4 h-4 shrink-0 opacity-40"></i>
              <select v-model="m.platformType" class="select select-bordered select-sm flex-1">
                <option value="" disabled>广告平台类型</option>
                <option v-for="p in createPlatformTypes" :key="p.type" :value="p.type">{{ p.type }} - {{ p.desc }}</option>
              </select>
              <button class="btn btn-ghost btn-xs btn-circle" @click="removeMappingRow(createForm, idx)" title="删除">
                <i data-lucide="minus" class="w-4 h-4"></i>
              </button>
            </div>
            <button class="btn btn-primary btn-outline btn-xs" @click="addMappingRow(createForm)">
              <i data-lucide="plus" class="w-4 h-4"></i> 添加映射
            </button>
          </div>

          <div class="divider text-xs my-4">联调模式</div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-2">
              <input type="checkbox" class="toggle toggle-primary toggle-sm" v-model="createForm.debugMode">
              <span class="label-text">开启联调</span>
            </label>
          </div>
          <div v-if="createForm.debugMode" class="mt-2">
            <MultiSelect
              v-model="createForm.debugEvents"
              :options="createPlatformTypes.map(p => ({value: p.type, label: p.type + ' - ' + p.desc}))"
              placeholder="选择关注的回传事件 (不选则全部放行)"
            />
          </div>

          <div class="divider text-xs my-4">虚假上报</div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-2">
              <input type="checkbox" class="toggle toggle-primary toggle-sm" v-model="createForm.fakeEnabled">
              <span class="label-text">开启虚假上报</span>
            </label>
          </div>
          <div v-if="createForm.fakeEnabled" class="grid grid-cols-3 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">概率 (1-100)</span></label>
              <input type="number" v-model.number="createForm.fakeRate" class="input input-bordered input-sm" min="1" max="100">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">延迟下限 (秒)</span></label>
              <input type="number" v-model.number="createForm.fakeDelayMin" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">延迟上限 (秒)</span></label>
              <input type="number" v-model.number="createForm.fakeDelayMax" class="input input-bordered input-sm">
            </div>
          </div>
        </div>
      </Modal>

      <!-- 编辑账户 (含行为映射) -->
      <Modal :show="showEdit" title="编辑账户" width="max-w-3xl" @close="showEdit=false" @confirm="saveEdit">
        <div class="space-y-3">
          <div class="form-control">
            <label class="label"><span class="label-text">账户名</span></label>
            <input v-model="editForm.name" class="input input-bordered input-sm">
          </div>

          <div class="divider text-xs my-4">账户凭证</div>
          <!-- 微博 -->
          <div v-if="editForm.platformName==='weibo'" class="grid grid-cols-1 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">公司代码</span></label>
              <input v-model="editForm.company" class="input input-bordered input-sm" placeholder="联调后台分配的 company">
            </div>
          </div>
          <!-- 小红书 -->
          <div v-else-if="editForm.platformName==='xiaohongshu'" class="grid grid-cols-1 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">应用标识</span></label>
              <input v-model="editForm.platform" class="input input-bordered input-sm" placeholder="应用英文名, 如 taobao">
            </div>
          </div>
          <!-- 其他 -->
          <div v-else class="grid grid-cols-3 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">公司代码</span></label>
              <input v-model="editForm.company" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">广告主ID</span></label>
              <input v-model="editForm.advertiserId" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">应用标识</span></label>
              <input v-model="editForm.platform" class="input input-bordered input-sm">
            </div>
          </div>

          <div class="divider text-xs my-4">淘联配置</div>
          <div class="grid grid-cols-3 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">渠道ID</span></label>
              <input v-model="editForm.channel" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">任务ID</span></label>
              <input v-model="editForm.taskId" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">推广PID</span></label>
              <input v-model="editForm.tbkId" class="input input-bordered input-sm">
            </div>
          </div>

          <div class="divider text-xs my-4">行为映射</div>
          <div class="space-y-2">
            <div v-for="(m, idx) in editForm.mappings" :key="idx" class="flex gap-2 items-center">
              <select v-model="m.taolianType" class="select select-bordered select-sm flex-1">
                <option :value="null" disabled>淘联类型</option>
                <option v-for="t in options.taolianTypes" :key="t.type" :value="t.type">{{ t.type }} - {{ t.desc }}</option>
              </select>
              <i data-lucide="arrow-right" class="w-4 h-4 shrink-0 opacity-40"></i>
              <select v-model="m.platformType" class="select select-bordered select-sm flex-1">
                <option value="" disabled>广告平台类型</option>
                <option v-for="p in editPlatformTypes" :key="p.type" :value="p.type">{{ p.type }} - {{ p.desc }}</option>
              </select>
              <button class="btn btn-ghost btn-xs btn-circle" @click="removeMappingRow(editForm, idx)" title="删除">
                <i data-lucide="minus" class="w-4 h-4"></i>
              </button>
            </div>
            <button class="btn btn-primary btn-outline btn-xs" @click="addMappingRow(editForm)">
              <i data-lucide="plus" class="w-4 h-4"></i> 添加映射
            </button>
          </div>

          <div class="divider text-xs my-4">联调模式</div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-2">
              <input type="checkbox" class="toggle toggle-primary toggle-sm" v-model="editForm.debugMode">
              <span class="label-text">开启联调</span>
            </label>
          </div>
          <div v-if="editForm.debugMode" class="mt-2">
            <MultiSelect
              v-model="editForm.debugEvents"
              :options="editPlatformTypes.map(p => ({value: p.type, label: p.type + ' - ' + p.desc}))"
              placeholder="选择关注的回传事件 (不选则全部放行)"
            />
          </div>

          <div class="divider text-xs my-4">虚假上报</div>
          <div class="form-control">
            <label class="label cursor-pointer justify-start gap-2">
              <input type="checkbox" class="toggle toggle-primary toggle-sm" v-model="editForm.fakeEnabled">
              <span class="label-text">开启虚假上报</span>
            </label>
          </div>
          <div v-if="editForm.fakeEnabled" class="grid grid-cols-3 gap-3">
            <div class="form-control">
              <label class="label"><span class="label-text">概率 (1-100)</span></label>
              <input type="number" v-model.number="editForm.fakeRate" class="input input-bordered input-sm" min="1" max="100">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">延迟下限 (秒)</span></label>
              <input type="number" v-model.number="editForm.fakeDelayMin" class="input input-bordered input-sm">
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">延迟上限 (秒)</span></label>
              <input type="number" v-model.number="editForm.fakeDelayMax" class="input input-bordered input-sm">
            </div>
          </div>

          <div class="divider text-xs my-4">状态</div>
          <select v-model.number="editForm.status" class="select select-bordered select-sm">
            <option :value="1">启用</option><option :value="0">禁用</option>
          </select>
        </div>
      </Modal>
    </div>
  `,
}
