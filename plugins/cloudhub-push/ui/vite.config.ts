import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import externalGlobals from 'rollup-plugin-external-globals'
import fs from 'fs'
import path from 'path'
import { fileURLToPath } from 'url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// 自定义插件：在构建结束后，自动将带有 hash 值的产物名称更新到 manifest.json 中
function updateManifestPlugin() {
  return {
    name: 'update-manifest',
    writeBundle(options: any, bundle: any) {
      const manifestPath = path.resolve(__dirname, './manifest.json')
      const mainManifestPath = path.resolve(__dirname, '../manifest.json')
      const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf-8'))
      const mainManifest = JSON.parse(fs.readFileSync(mainManifestPath, 'utf-8'))

      // 同步版本号
      manifest.version = mainManifest.version

      let jsFile = ''
      let cssFile = ''

      for (const fileName in bundle) {
        const chunk = bundle[fileName]
        // 寻找入口 JS
        if (chunk.type === 'chunk' && chunk.isEntry) {
          jsFile = fileName
        }
        // 寻找样式 CSS
        else if (chunk.type === 'asset' && fileName.endsWith('.css')) {
          cssFile = fileName
        }
      }

      manifest.entry = `out/${jsFile}`
      if (cssFile) {
        manifest.style = `out/${cssFile}`
      } else {
        delete manifest.style
      }

      fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, 2))
      console.log(`\n✅ 已自动更新 ui/manifest.json: version=${manifest.version}, entry=${jsFile}, style=${cssFile}`)
    },
  }
}

export default defineConfig({
  plugins: [vue(), updateManifestPlugin()],
  build: {
    outDir: 'out',
    emptyOutDir: true, // 构建前自动清空 out 目录下的旧 hash 文件
    rollupOptions: {
      input: './src/index.ts',
      preserveEntrySignatures: 'strict',
      output: {
        format: 'es',
        entryFileNames: 'index-[hash].js',
        assetFileNames: 'style-[hash].[ext]',
      },
      // 不将 vue 和 naive-ui 打包进去，而是声明为外部依赖
      external: ['vue', 'naive-ui'],
      plugins: [
        // 关键：拦截外部依赖，将其映射到宿主注入的全局变量
        externalGlobals({
          vue: 'window.__SHARED_VUE__',
          'naive-ui': 'window.__SHARED_NAIVE__',
        }),
      ],
    },
  },
})
