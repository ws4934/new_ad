// Lucide ESM shim: 将 UMD 全局对象桥接为 ES module export
// 解决: 如果有代码残余 `import { createIcons } from 'lucide'`，
// importmap 会指到这个 shim, 从 window.lucide 透传
const lucide = window.lucide || {}
export const createIcons = lucide.createIcons || (() => {})
export default lucide
