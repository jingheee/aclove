import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
// 全局引入（学习阶段先这样，后面可改按需）
import NaiveUI from 'naive-ui'
import { createDiscreteApi } from 'naive-ui'  // 用于 message/dialog 等独立使用
const app=createApp(App)
app.use(NaiveUI)
// 可选：全局 message / notification / dialog / loadingBar
const { message, notification, dialog, loadingBar } = createDiscreteApi([
  'message',
  'dialog',
  'notification',
  'loadingBar'
])

// 挂载到 app.config.globalProperties（可选，方便 this.$message）
app.config.globalProperties.$message = message
app.config.globalProperties.$dialog = dialog
app.mount('#app')
