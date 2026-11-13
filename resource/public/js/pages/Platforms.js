import { ref, inject, onMounted } from 'vue'
import { api } from '../api.js?v=6'

export default {
  setup() {
    const deleteTarget = ref(null)
    const deleting = ref(false)
    const list = ref([])
    const showEdit = ref(false)
    const editForm = ref({ id: 0, displayName: '', reportUrl: '', status: 1 })
    const toast = inject('toast')

    const refresh = async () => {
      const res = await api.platformList()
      list.value = res.list || []
    }

    const openEdit = (p) => {
      // 从 config JSON 解出 reportUrl
      let cfg = {}
      try { cfg = typeof p.config === 'string' ? JSON.parse(p.config) : (p.config || {}) } catch {}
      editForm.value = {
        id: p.id,
        displayName: p.displayName,
        reportUrl: cfg.reportUrl || '',
        status: p.status,
      }
      showEdit.value = true
    }

    const saveEdit = async () => {
      try {
        // 把表单字段组装回 config JSON
        const config = JSON.stringify({ reportUrl: editForm.value.reportUrl })
        await api.platformUpdate(editForm.value.id, {
          displayName: editForm.value.displayName,
          config: config,
          status: editForm.value.status,
        })
        showEdit.value = false
        toast.show('保存成功', 'success')
        await refresh()
      } catch (e) { toast.show('失败: ' + e.message, 'error') }
    }

    const remove = async () => {
      if (!deleteTarget.value || deleting.value) return
      deleting.value = true
      try {
        await api.platformDelete(deleteTarget.value.id)
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
    return { deleteTarget, deleting, remove, list, showEdit, editForm, openEdit, saveEdit }
  },
  template: `
    <div class="space-y-4">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <i data-lucide="settings" class="w-6 h-6"></i> 平台管理
      </h1>

      <Modal :show="!!deleteTarget" title="删除平台" :confirm-text="deleting ? '删除中…' : '确认删除'" :confirm-disabled="deleting" @close="!deleting && (deleteTarget=null)" @confirm="remove">
        <p>确认删除平台「{{ deleteTarget?.displayName }}」？此操作不可恢复。</p>
      </Modal>

      <div class="card bg-base-100 shadow">
        <div class="card-body p-0">
          <table class="table">
            <thead>
              <tr><th>ID</th><th>标识</th><th>显示名</th><th>状态</th><th>操作</th></tr>
            </thead>
            <tbody>
              <tr v-for="p in list" :key="p.id">
                <td>{{ p.id }}</td>
                <td><code>{{ p.name }}</code></td>
                <td>{{ p.displayName }}</td>
                <td><span :class="p.status===1?'badge badge-primary badge-sm':'badge badge-ghost badge-sm'">{{ p.status===1?'启用':'禁用' }}</span></td>
                <td>
                  <button class="btn btn-xs btn-ghost" @click="openEdit(p)">
                    <i data-lucide="edit" class="w-3.5 h-3.5"></i> 编辑
                  </button>
                    <!-- 
                  <button class="btn btn-xs btn-ghost text-error" @click="deleteTarget=p" :disabled="deleting" title="删除平台">
                    <i data-lucide="trash-2" class="w-3.5 h-3.5"></i> 删除
                  </button>
                   -->
                </td>
              </tr>
              <tr v-if="!list.length"><td colspan="5" class="text-center text-gray-400">暂无平台</td></tr>
            </tbody>
          </table>
        </div>
      </div>

      <Modal :show="showEdit" title="编辑平台" @close="showEdit=false" @confirm="saveEdit">
        <div class="space-y-3">
          <div class="form-control">
            <label class="label"><span class="label-text">显示名称</span></label>
            <input v-model="editForm.displayName" class="input input-bordered input-sm">
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">回传接口地址</span></label>
            <input v-model="editForm.reportUrl" class="input input-bordered input-sm"
              placeholder="https://appmonitor.biz.weibo.com/sdkserver/active">
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">状态</span></label>
            <select v-model.number="editForm.status" class="select select-bordered select-sm">
              <option :value="1">启用</option>
              <option :value="0">禁用</option>
            </select>
          </div>
        </div>
      </Modal>
    </div>
  `,
}
