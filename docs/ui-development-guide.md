# 插件前端 UI 扩展开发指南

pan-302 允许插件提供自定义的前端设置/控制面板，显示在宿主系统的“插件管理 -> 插件设置”弹窗中。
为了保持统一的界面交互和极致的加载性能，插件前端 UI 应该**直接共享宿主运行的 Vue 3 和 Naive UI 依赖**，保证页面主题（暗黑/明亮色）和中文语言包能够平滑同步跟随切换。

---

## 核心工作原理（宿主与插件 UI 共享机制）

1. **运行时依赖共享（Vite External）**：
   宿主在加载插件前端时，已经将当前的 `Vue` 实例与 `Naive UI` 组件包暴露在全局：
   - `window.__SHARED_VUE__`
   - `window.__SHARED_NAIVE__`
   插件的前端在打包（pnpm run build）时，使用 `rollup-plugin-external-globals` 插件拦截对 `vue` 和 `naive-ui` 的 import 依赖，直接重定向到这两个 window 全局变量。因此，即使插件使用单文件组件（.vue）开发，打包出来的 JS 大小通常也只有约 **5KB** 到 **10KB**，没有任何冗余体积！

2. **生命周期挂载契约**：
   插件的 `ui/manifest.json` 声明的 entry JS 文件（ESM 格式）必须默认导出（export default）三个核心生命周期钩子函数：
   ```typescript
   export default {
     // 当用户打开插件设置弹窗时触发，负责将插件前端挂载到 DOM 容器中
     mount(container: HTMLElement, props: Record<string, any>) { ... },
     // 当传入的 props 发生变化时调用（可选）
     update(props: Record<string, any>) { ... },
     // 当关闭插件设置弹窗时触发，负责销毁子 Vue 实例并清空 DOM 容器
     unmount(container: HTMLElement) { ... }
   }
   ```

3. **宿主注入的 Runtime 资源**：
   宿主在调用插件的 `mount(container, props)` 时，会通过 `props` 传入以下核心对象：
   - `configApi`：插件配置接口的后端 URL（支持 GET/PUT 请求，GET 读取配置，PUT 保存配置，内部已自动处理跨域和 Basic 鉴权）。
   - `runtime`：宿主注入的前端运行状态。
     - `runtime.theme`：Naive UI 的响应式暗黑主题（`naive.darkTheme` 或 `null`）。
     - `runtime.themeOverrides`：宿主自定义的主题色调覆盖配置。
     - `runtime.locale` / `runtime.dateLocale`：宿主自带的中文化 Naive UI 语言包。

---

## 插件前端目录推荐

插件前端一般存放在插件目录下的 `ui/` 文件夹中：
```
ui/
├── src/
│   ├── index.ts        # 前端挂载入口
│   └── SettingsForm.vue # 用 Vue SFC 单文件组件编写的设置表单组件
├── vite.config.ts      # Vite 打包配置
├── package.json        # 依赖与打包脚本
├── tsconfig.json       # TS 配置
├── manifest.json       # UI 描述符
└── .gitignore          # Git 忽略文件（忽略 node_modules/ 和 out/ 缓存）
```

---

## 关键代码实现示例

### 1. `vite.config.ts` 配置
关键点在于剔除 `vue` 和 `naive-ui` 并将其映射到 `__SHARED_` 全局对象；并且在打包出带 Hash 的 JS / CSS 后，自动把最新的文件名写回 `ui/manifest.json`。
```typescript
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import externalGlobals from 'rollup-plugin-external-globals'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// 自动把带 hash 的文件名更新回 ui/manifest.json 供宿主读取的构建插件
function updateManifestPlugin() {
  return {
    name: 'update-manifest',
    writeBundle(options: any, bundle: any) {
      const manifestPath = path.resolve(__dirname, './manifest.json')
      const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
      let jsFile = ''
      let cssFile = ''
      for (const fileName in bundle) {
        if (fileName.endsWith('.js')) jsFile = fileName
        if (fileName.endsWith('.css')) cssFile = fileName
      }
      manifest.entry = `out/${jsFile}`
      if (cssFile) {
        manifest.style = `out/${cssFile}`
      }
      fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2))
    },
  }
}

export default defineConfig({
  plugins: [vue(), updateManifestPlugin()],
  build: {
    outDir: 'out', // 必须使用 out 避开打包 CLI 过滤的 dist 关键字
    emptyOutDir: true,
    rollupOptions: {
      input: './src/index.ts',
      preserveEntrySignatures: 'strict',
      output: {
        format: 'es',
        entryFileNames: 'index-[hash].js',
        assetFileNames: 'style-[hash].[ext]',
      },
      external: ['vue', 'naive-ui'],
      plugins: [
        externalGlobals({
          vue: 'window.__SHARED_VUE__',
          'naive-ui': 'window.__SHARED_NAIVE__',
        }),
      ],
    },
  },
})
```

### 2. `src/index.ts` 挂载入口
```typescript
import { createApp, type App } from 'vue'
import SettingsForm from './SettingsForm.vue'

let app: App | null = null

export default {
  mount(container: HTMLElement, props: Record<string, any> = {}) {
    app = createApp(SettingsForm, { ...props })
    app.mount(container)
  },
  unmount(container: HTMLElement) {
    if (app) {
      app.unmount()
      app = null
    }
    container.replaceChildren()
  },
}
```

### 3. `src/SettingsForm.vue` 设置组件
在 `<NConfigProvider>` 包装器中传入宿主的 `runtime`，并在 TypeScript 中**显式导入** Naive UI 的组件，以避开 IDE 红线并确保运行时能够完全渲染：
```vue
<template>
  <NConfigProvider
    :theme="runtime.theme.value"
    :theme-overrides="runtime.themeOverrides.value"
    :locale="runtime.locale"
    :date-locale="runtime.dateLocale"
  >
    <section class="my-settings">
      <NForm :model="config" label-placement="top">
        <NFormItem label="推送地址" path="url">
          <NInput v-model:value="config.url" placeholder="https://..." />
        </NFormItem>
        <NSpace justify="end">
          <NButton type="primary" :loading="loading" @click="saveConfig">
            保存设置
          </NButton>
        </NSpace>
      </NForm>
    </section>
  </NConfigProvider>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { NConfigProvider, NForm, NFormItem, NInput, NSpace, NButton } from 'naive-ui'

const props = defineProps<{
  runtime: any
  configApi: string
}>()

const config = ref({ url: '' })
const loading = ref(false)

async function loadConfig() {
  loading.value = true
  try {
    const response = await fetch(props.configApi)
    const payload = await response.json()
    if (payload.code === 0) {
      config.value = payload.data
    }
  } finally {
    loading.value = false
  }
}

onMounted(loadConfig)
</script>
```

---

## 最佳参考实例

- 完整采用 Vite + Vue SFC 编译、共享宿主依赖、并且自动将 Hash 产物回写 manifest 文件的官方插件参考：
  👉 [plugins/cloudhub-push](file:///Users/jianxiong.cao/work/fun/pan-302/pan-302-plugin/plugins/cloudhub-push)
- 零本地前端打包工具链，通过手写原生 HTML/JS 渲染界面的极简 Go 教学插件参考：
  👉 [examples/strm-go-example](file:///Users/jianxiong.cao/work/fun/pan-302/pan-302-plugin/examples/strm-go-example)
