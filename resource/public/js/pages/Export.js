import { ref, inject, onMounted } from 'vue'
import { api, getToken } from '../api.js'

export default {
  setup() {
    const filter = ref({ platform: '', field: '' })
    const files = ref([])
    const toast = inject('toast')

    const refresh = async () => {
      try {
        const res = await api.exportFiles(filter.value)
        files.value = res.files || []
      } catch (e) {
        files.value = []
      }
    }

    const formatSize = (bytes) => {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
      return (bytes / (1024 * 1024 * 1024)).toFixed(2) + ' GB'
    }

    const download = async (filename) => {
      try {
        const resp = await fetch(api.exportDownloadURL(filename), {
          headers: { 'Authorization': 'Bearer ' + getToken() },
        })
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`)
        const blob = await resp.blob()
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = filename
        a.click()
        URL.revokeObjectURL(url)
      } catch (e) {
        toast.show('下载失败: ' + e.message, 'error')
      }
    }

    const deleteFile = async (filename) => {
      if (!confirm(`确认删除 ${filename}?`)) return
      try {
        await api.exportDelete(filename)
        toast.show('已删除', 'success')
        await refresh()
      } catch (e) {
        toast.show('删除失败: ' + e.message, 'error')
      }
    }

    onMounted(refresh)
    return { filter, files, refresh, formatSize, download, deleteFile }
  },
  template: `
    <div class="space-y-4">
      <h1 class="text-2xl font-bold flex items-center gap-2">
        <i data-lucide="file-down" class="w-6 h-6"></i> CSV 导出
      </h1>
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <div class="flex gap-2 mb-4 flex-wrap">
            <select v-model="filter.platform" class="select select-bordered select-sm">
              <option value="">所有平台</option>
              <option value="weibo">微博</option>
              <option value="xiaohongshu">小红书</option>
            </select>
            <select v-model="filter.field" class="select select-bordered select-sm">
              <option value="">所有设备号</option>
              <option value="imei_md5">imei_md5</option>
              <option value="idfa_md5">idfa_md5</option>
              <option value="oaid_md5">oaid_md5</option>
              <option value="caid_md5">caid_md5</option>
            </select>
            <button class="btn btn-primary btn-sm" @click="refresh">
              <i data-lucide="search" class="w-4 h-4"></i> 查询
            </button>
          </div>
          <table class="table">
            <thead>
              <tr><th>文件名</th><th>大小</th><th>操作</th></tr>
            </thead>
            <tbody>
              <tr v-for="f in files" :key="f.name">
                <td><code class="text-sm">{{ f.name }}</code></td>
                <td class="text-sm text-gray-500">{{ formatSize(f.size) }}</td>
                <td class="flex gap-1">
                  <button class="btn btn-xs btn-primary" @click="download(f.name)">
                    <i data-lucide="download" class="w-3.5 h-3.5"></i> 下载
                  </button>
                  <button class="btn btn-xs btn-ghost text-error" @click="deleteFile(f.name)">
                    <i data-lucide="trash-2" class="w-3.5 h-3.5"></i>
                  </button>
                </td>
              </tr>
              <tr v-if="!files.length"><td colspan="3" class="text-center text-gray-400">暂无文件</td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `,
}
