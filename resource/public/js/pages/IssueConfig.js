import { ref, inject, onMounted } from 'vue'
import { api } from '../api.js?v=8'

const CATS = { 1: '超级红包', 2: '福利购', 4: '福利直降' }
const emptyForm = () => ({ id: 0, name: '', appKey: '34070427', appSecret: '', activityCategory: 1, intervalSec: 30, notifyEmails: '', enabled: true })
const LAST_KEY = 'issue_last_form' // 上次新增的配置 (不含 AppSecret)
const loadLast = () => { try { return JSON.parse(localStorage.getItem(LAST_KEY)) } catch { return null } }

export default {
  setup() {
    const toast = inject('toast')
    const list = ref([])
    const showForm = ref(false)
    const form = ref(emptyForm())
    const saving = ref(false)
    const deleteTarget = ref(null)
    const deleting = ref(false)
    const smtp = ref({ host: '', port: 465, username: '', password: '', from: '', passwordSet: false })
    const smtpSaving = ref(false)
    const testTo = ref('')
    const testing = ref(false)

    const refresh = async () => {
      try { list.value = (await api.issueList()).list || [] }
      catch (e) { toast.show('加载失败: ' + e.message, 'error') }
    }
    const loadSmtp = async () => {
      try {
        const r = await api.smtpGet()
        smtp.value = { host: '', username: '', from: '', ...r, port: r.port || 465, password: '' }
      }
      catch (e) { toast.show('加载邮件配置失败: ' + e.message, 'error') }
    }

    // 基于模板 (上次新增 / 某行) 预填, 活动类型切到该 AppKey 尚未配置的下一个, 名称里的活动名同步替换
    const openCreate = (tpl) => {
      const base = { ...emptyForm(), ...(tpl || loadLast() || {}), id: 0, appSecret: '' }
      const used = new Set(list.value.filter(m => m.appKey === base.appKey).map(m => m.activityCategory))
      const next = [1, 2, 4].find(c => !used.has(c))
      if (next && next !== base.activityCategory) {
        const oldLabel = CATS[base.activityCategory]
        base.name = base.name.includes(oldLabel) ? base.name.replace(oldLabel, CATS[next]) : base.name
        base.activityCategory = next
      }
      form.value = {
        id: 0, name: base.name, appKey: base.appKey, appSecret: '', activityCategory: base.activityCategory,
        intervalSec: base.intervalSec, notifyEmails: base.notifyEmails || '', enabled: base.enabled,
      }
      showForm.value = true
    }
    const hasSecret = (appKey) => list.value.some(m => m.appKey === appKey.trim())
    const openEdit = (m) => {
      const { id, name, appKey, activityCategory, intervalSec, notifyEmails, enabled } = m
      form.value = { id, name, appKey, appSecret: '', activityCategory, intervalSec, notifyEmails: notifyEmails || '', enabled }
      showForm.value = true
    }

    const save = async () => {
      const f = form.value
      if (!f.name.trim() || !f.appKey.trim()) return toast.show('名称和 AppKey 必填', 'error')
      if (!f.id && !f.appSecret && !hasSecret(f.appKey)) return toast.show('新增时 AppSecret 必填', 'error')
      if (!(f.intervalSec >= 10)) return toast.show('查询间隔不能小于 10 秒', 'error')
      const body = {
        name: f.name.trim(), appKey: f.appKey.trim(), appSecret: f.appSecret,
        activityCategory: f.activityCategory, intervalSec: f.intervalSec, enabled: f.enabled,
        notifyEmails: f.notifyEmails.split(/[\s,，;；]+/).filter(Boolean).join(','),
      }
      saving.value = true
      try {
        if (f.id) await api.issueUpdate(f.id, body)
        else {
          await api.issueCreate(body)
          const { appSecret, ...last } = body
          localStorage.setItem(LAST_KEY, JSON.stringify(last))
        }
        showForm.value = false
        toast.show('保存成功', 'success')
        await refresh()
      } catch (e) { toast.show('保存失败: ' + e.message, 'error') }
      saving.value = false
    }

    const remove = async () => {
      if (!deleteTarget.value || deleting.value) return
      deleting.value = true
      try {
        await api.issueDelete(deleteTarget.value.id)
        deleteTarget.value = null
        toast.show('删除成功', 'success')
        await refresh()
      } catch (e) { toast.show('删除失败: ' + e.message, 'error') }
      deleting.value = false
    }

    const saveSmtp = async () => {
      const { host, port, username, password, from } = smtp.value
      smtpSaving.value = true
      try {
        await api.smtpSave({ host: host.trim(), port: Number(port) || 0, username: username.trim(), password, from: from.trim() })
        toast.show('邮件配置已保存', 'success')
        await loadSmtp()
      } catch (e) { toast.show('保存失败: ' + e.message, 'error') }
      smtpSaving.value = false
    }

    const sendTest = async () => {
      if (!testTo.value.trim()) return toast.show('请填写收件人', 'error')
      testing.value = true
      try {
        await api.smtpTest(testTo.value.trim())
        toast.show('测试邮件已发送', 'success')
      } catch (e) { toast.show('发送失败: ' + e.message, 'error') }
      testing.value = false
    }

    onMounted(() => { refresh(); loadSmtp() })
    return { CATS, list, showForm, form, saving, openCreate, openEdit, save, hasSecret, deleteTarget, deleting, remove,
      smtp, smtpSaving, saveSmtp, testTo, testing, sendTest }
  },
  template: `
    <div class="space-y-4">
      <div class="flex justify-between items-center">
        <h1 class="text-2xl font-bold flex items-center gap-2">
          <i data-lucide="sliders-horizontal" class="w-6 h-6"></i> 直发配置
        </h1>
        <button class="btn btn-primary btn-sm" @click="openCreate()" title="自动带入上次新增的配置">
          <i data-lucide="plus" class="w-4 h-4"></i> 新增配置
        </button>
      </div>

      <div class="card bg-base-100 shadow">
        <div class="card-body p-0">
          <div class="px-5 pt-4 pb-1">
            <h2 class="font-semibold">查询配置</h2>
            <p class="text-xs text-base-content/50 mt-0.5">同一 AppKey 可分别添加超级红包 / 福利购 / 福利直降</p>
          </div>
          <div class="overflow-x-auto">
            <table class="table">
              <thead><tr><th>名称</th><th>活动</th><th>AppKey</th><th>间隔</th><th>通知邮箱</th><th>状态</th><th class="text-right">操作</th></tr></thead>
              <tbody>
                <tr v-for="m in list" :key="m.id">
                  <td class="font-medium">{{ m.name }}</td>
                  <td>{{ CATS[m.activityCategory] || m.activityCategory }}</td>
                  <td><code>{{ m.appKey }}</code></td>
                  <td class="tabular-nums">{{ m.intervalSec }} 秒</td>
                  <td class="max-w-xs truncate text-base-content/70" :title="m.notifyEmails">{{ m.notifyEmails || '—' }}</td>
                  <td><span :class="m.enabled ? 'badge badge-primary badge-sm' : 'badge badge-ghost badge-sm'">{{ m.enabled ? '启用' : '停用' }}</span></td>
                  <td class="whitespace-nowrap text-right">
                    <button class="btn btn-xs btn-ghost" @click="openEdit(m)">
                      <i data-lucide="edit" class="w-3.5 h-3.5"></i> 编辑
                    </button>
                    <button class="btn btn-xs btn-ghost" @click="openCreate(m)" title="以此配置为模板新增其他活动类型">
                      <i data-lucide="copy" class="w-3.5 h-3.5"></i> 复制
                    </button>
                    <button class="btn btn-xs btn-ghost text-error" @click="deleteTarget=m" :disabled="deleting">
                      <i data-lucide="trash-2" class="w-3.5 h-3.5"></i> 删除
                    </button>
                  </td>
                </tr>
                <tr v-if="!list.length"><td colspan="7" class="text-center text-base-content/40">暂无配置，点击右上角「新增配置」</td></tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div class="card bg-base-100 shadow">
        <div class="card-body p-5 gap-4">
          <div>
            <h2 class="font-semibold">邮件通知 (SMTP)</h2>
            <p class="text-xs text-base-content/50 mt-0.5">直发状态切换时，向各配置填写的通知邮箱发送提醒</p>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-x-4 gap-y-3">
            <label class="form-control xl:col-span-2">
              <span class="label-text text-xs mb-1">SMTP 服务器</span>
              <input v-model="smtp.host" class="input input-bordered input-sm" placeholder="smtp.qq.com">
            </label>
            <label class="form-control xl:col-span-2">
              <span class="label-text text-xs mb-1">端口 <span class="text-base-content/40">465 = SSL，587 / 25 = STARTTLS</span></span>
              <input v-model.number="smtp.port" type="number" class="input input-bordered input-sm">
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">用户名</span>
              <input v-model="smtp.username" class="input input-bordered input-sm" autocomplete="off">
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">密码 / 授权码</span>
              <input v-model="smtp.password" type="password" class="input input-bordered input-sm" autocomplete="new-password"
                :placeholder="smtp.passwordSet ? '已设置，留空不修改' : ''">
            </label>
            <label class="form-control md:col-span-2">
              <span class="label-text text-xs mb-1">发件人 <span class="text-base-content/40">可选，留空使用用户名</span></span>
              <input v-model="smtp.from" class="input input-bordered input-sm" placeholder="直发提醒 <a@b.com>">
            </label>
          </div>
          <div class="flex flex-wrap items-center justify-between gap-3 pt-4 border-t border-base-200">
            <div class="flex flex-wrap items-center gap-2">
              <input v-model="testTo" class="input input-bordered input-sm w-60" placeholder="测试收件人邮箱">
              <button class="btn btn-sm btn-ghost" :disabled="testing" @click="sendTest">
                <span v-if="testing" class="loading loading-spinner loading-xs"></span>
                <i v-else data-lucide="send" class="w-4 h-4"></i> 发送测试邮件
              </button>
              <span class="text-xs text-base-content/40">使用已保存的配置发送</span>
            </div>
            <button class="btn btn-primary btn-sm" :disabled="smtpSaving" @click="saveSmtp">
              <i data-lucide="save" class="w-4 h-4"></i> {{ smtpSaving ? '保存中…' : '保存邮件配置' }}
            </button>
          </div>
        </div>
      </div>

      <Modal :show="showForm" :title="form.id ? '编辑查询配置' : '新增查询配置'" :confirm-text="saving ? '保存中…' : '保存'"
        :confirm-disabled="saving" @close="showForm=false" @confirm="save">
        <div class="space-y-3">
          <label class="form-control">
            <span class="label-text text-xs mb-1">名称</span>
            <input v-model="form.name" class="input input-bordered input-sm">
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">AppKey</span>
            <input v-model="form.appKey" class="input input-bordered input-sm">
          </label>
          <label class="form-control">
            <span class="label-text text-xs mb-1">AppSecret</span>
            <input v-model="form.appSecret" type="password" class="input input-bordered input-sm" autocomplete="new-password"
              :placeholder="form.id ? '留空不修改' : (hasSecret(form.appKey) ? '留空则复用该 AppKey 已保存的 Secret' : '必填')">
          </label>
          <div class="grid grid-cols-2 gap-3">
            <label class="form-control">
              <span class="label-text text-xs mb-1">活动类型</span>
              <select v-model.number="form.activityCategory" class="select select-bordered select-sm">
                <option :value="1">超级红包</option>
                <option :value="2">福利购</option>
                <option :value="4">福利直降</option>
              </select>
            </label>
            <label class="form-control">
              <span class="label-text text-xs mb-1">查询间隔（秒）</span>
              <input v-model.number="form.intervalSec" type="number" min="10" class="input input-bordered input-sm">
            </label>
          </div>
          <label class="form-control">
            <span class="label-text text-xs mb-1">通知邮箱（逗号或换行分隔）</span>
            <textarea v-model="form.notifyEmails" rows="3" class="textarea textarea-bordered textarea-sm"></textarea>
          </label>
          <label class="label cursor-pointer justify-start gap-3">
            <input type="checkbox" class="toggle toggle-primary toggle-sm" v-model="form.enabled">
            <span class="label-text">启用</span>
          </label>
        </div>
      </Modal>

      <Modal :show="!!deleteTarget" title="删除查询配置" :confirm-text="deleting ? '删除中…' : '确认删除'" :confirm-disabled="deleting"
        @close="!deleting && (deleteTarget=null)" @confirm="remove">
        <p>确认删除「{{ deleteTarget?.name }}」？此操作不可恢复。</p>
      </Modal>
    </div>
  `,
}
