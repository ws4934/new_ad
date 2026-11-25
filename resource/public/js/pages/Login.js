import { ref } from 'vue'
import { api, setToken } from '../api.js?v=19'
import { useRouter } from 'vue-router'

export default {
  setup() {
    const password = ref('')
    const error = ref('')
    const loading = ref(false)
    const router = useRouter()

    const login = async () => {
      error.value = ''
      loading.value = true
      try {
        const res = await api.login(password.value)
        setToken(res.token)
        router.push('/')
      } catch (e) {
        error.value = e.message
      } finally {
        loading.value = false
      }
    }

    return { password, error, loading, login }
  },
  template: `
    <div class="min-h-screen flex items-center justify-center bg-base-200">
      <div class="card w-96 bg-base-100 shadow-xl">
        <div class="card-body">
          <p class="text-center text-gray-500 text-sm mt-1">请输入管理密码登录</p>

          <div v-if="error" class="alert alert-error mt-3 text-sm py-2">{{ error }}</div>

          <div class="form-control mt-4">
            <input
              type="password"
              v-model="password"
              class="input input-bordered"
              placeholder="密码"
              @keyup.enter="login"
              autofocus
            >
          </div>

          <div class="form-control mt-4">
            <button
              class="btn btn-primary"
              :class="{'loading': loading}"
              :disabled="!password || loading"
              @click="login"
            >
              登录
            </button>
          </div>
        </div>
      </div>
    </div>
  `,
}
