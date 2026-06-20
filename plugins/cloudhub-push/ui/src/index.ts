import { createApp, type App } from 'vue'
import SettingsForm from './SettingsForm.vue'

let app: App | null = null

export default {
  mount(container: HTMLElement, props: Record<string, any> = {}) {
    app = createApp(SettingsForm, {
      ...props,
    })
    app.mount(container)
  },

  update(props: Record<string, any> = {}) {
    // 可以响应组件 props 的变化
  },

  unmount(container: HTMLElement) {
    if (app) {
      app.unmount()
      app = null
    }
    container.replaceChildren()
  },
}
