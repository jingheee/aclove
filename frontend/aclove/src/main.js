import { createApp } from 'vue'
import './style.css'
import App from './App.vue'
import NaiveUI from 'naive-ui'
import { createDiscreteApi } from 'naive-ui'
import { VueQueryPlugin } from './queryClient'

const app = createApp(App)

app.use(NaiveUI)
app.use(VueQueryPlugin)

const { message, notification, dialog, loadingBar } = createDiscreteApi([
  'message',
  'dialog',
  'notification',
  'loadingBar'
])

app.config.globalProperties.$message = message
app.config.globalProperties.$dialog = dialog

app.mount('#app')
