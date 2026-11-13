import { ref } from 'vue'

export function useToast() {
  const message = ref('')
  const type = ref('info')
  let timer = null

  const show = (msg, t = 'info', duration = 2500) => {
    message.value = msg
    type.value = t
    clearTimeout(timer)
    timer = setTimeout(() => { message.value = '' }, duration)
  }

  return { message, type, show }
}

export const ToastView = {
  props: ['message', 'type'],
  template: `
    <Transition name="toast">
      <div v-if="message" class="fixed top-4 right-4 z-50">
        <div class="bg-base-100 shadow-lg rounded-lg border-l-4 px-4 py-3 flex items-center gap-2 text-sm min-w-[200px]"
             :class="{
               'border-primary text-primary': type==='success',
               'border-error text-error': type==='error',
               'border-warning text-warning': type==='warning',
               'border-primary/50 text-base-content': type==='info',
             }">
          <i :data-lucide="type==='error' ? 'alert-circle' : type==='warning' ? 'alert-triangle' : 'check-circle'" class="w-4 h-4 shrink-0"></i>
          <span>{{ message }}</span>
        </div>
      </div>
    </Transition>
  `,
}
