// 通用弹窗组件
// 用法: <Modal :show="visible" title="标题" @close="visible=false" @confirm="save">
//         <template #default>...表单内容...</template>
//       </Modal>
export default {
  props: {
    show: { type: Boolean, default: false },
    title: { type: String, default: '' },
    width: { type: String, default: 'max-w-lg' },
    confirmText: { type: String, default: '保存' },
    confirmDisabled: { type: Boolean, default: false },
    showFooter: { type: Boolean, default: true },
  },
  emits: ['close', 'confirm'],
  template: `
    <dialog class="modal" :class="{'modal-open': show}">
      <div class="modal-box" :class="width">
        <h3 class="font-bold text-lg" v-if="title">{{ title }}</h3>
        <div class="mt-4">
          <slot/>
        </div>
        <div class="modal-action" v-if="showFooter">
          <button class="btn btn-sm" @click="$emit('close')">取消</button>
          <button class="btn btn-primary btn-sm" :disabled="confirmDisabled" @click="$emit('confirm')">{{ confirmText }}</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button @click="$emit('close')">close</button></form>
    </dialog>
  `,
}
