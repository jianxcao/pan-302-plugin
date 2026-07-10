<template>
  <NConfigProvider
    :theme="runtime.theme.value"
    :theme-overrides="runtime.themeOverrides.value"
    :locale="runtime.locale"
    :date-locale="runtime.dateLocale"
  >
    <section class="cloudhub-push-settings">
      <header class="plugin-heading">
        <div>
          <NText depth="3" class="plugin-kicker">CloudHub Push</NText>
          <h2>CloudHub 资源推送</h2>
        </div>
        <NText depth="3">
          默认在 STRM 创建 120 秒后推送资源，删除始终跟随 STRM 删除事件。
        </NText>
      </header>

      <NAlert type="warning" :show-icon="false" class="setup-alert">
        仅当开启“使用媒体库新增事件”时，才需要在 Emby/Jellyfin Webhook 中启用媒体库新增事件；无需监听媒体删除事件。
      </NAlert>

      <NAlert v-if="status" :type="statusType" closable @close="status = ''" class="status-alert">
        {{ status }}
      </NAlert>

      <NForm :model="config" label-placement="top" :disabled="loading" class="plugin-form">
        <NFormItem label="节点 ID" path="node_id">
          <NInput v-model:value="config.node_id" placeholder="例如 media302" autocomplete="off" />
        </NFormItem>

        <NFormItem label="CloudHub API URL" path="base_url">
          <NInput v-model:value="config.base_url" placeholder="https://cloudhub.example.com" />
        </NFormItem>

        <NFormItem label="节点 API Key" path="api_key">
          <NInput
            v-model:value="config.api_key"
            type="password"
            show-password-on="click"
            autocomplete="new-password"
          />
        </NFormItem>

        <NFormItem label="节点公开 URL（可选）" path="public_base_url">
          <NInput v-model:value="config.public_base_url" placeholder="https://pan.example.com" />
        </NFormItem>

        <NFormItem label="单批数量" path="batch_size">
          <NInputNumber v-model:value="config.batch_size" :min="1" :max="500" />
        </NFormItem>

        <NFormItem label="包含路径" path="include_paths" feedback="只有包含在这些路径内的文件或 STRM 才会推送到 CloudHub，留空代表全部推送。">
          <NDynamicInput v-model:value="config.include_paths" placeholder="例如 /电影" />
        </NFormItem>

        <NFormItem label="使用媒体库新增事件" path="use_media_added_event">
          <NSwitch v-model:value="config.use_media_added_event" />
          <template #feedback>
            默认关闭：STRM 创建 120 秒后推送。开启后仅在 Emby/Jellyfin 新增媒体时推送；删除始终使用 STRM 删除事件。
          </template>
        </NFormItem>

        <NSpace justify="end">
          <NButton type="primary" :loading="loading" @click="saveConfig">保存设置</NButton>
        </NSpace>
      </NForm>
    </section>
  </NConfigProvider>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  NConfigProvider,
  NText,
  NAlert,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NSpace,
  NButton,
  NDynamicInput,
  NSwitch,
} from 'naive-ui'

const props = defineProps<{
  runtime: any
  configApi: string
  apiBase: string
  pluginName: string
  onClose?: () => void
}>()

const config = ref({
  node_id: '',
  base_url: '',
  api_key: '',
  public_base_url: '',
  batch_size: 500,
  include_paths: [] as string[],
  use_media_added_event: false,
})

const loading = ref(false)
const status = ref('')
const statusType = ref<'info' | 'success' | 'error'>('info')

function authHeaders(): string {
  const credentials = localStorage.getItem('pan302_auth')
  return credentials ? `Basic ${credentials}` : ''
}

async function request(url: string, options: RequestInit = {}) {
  const headers = new Headers(options.headers)
  headers.set('content-type', 'application/json')
  const auth = authHeaders()
  if (auth) {
    headers.set('Authorization', auth)
  }

  const response = await fetch(url, {
    ...options,
    headers,
  })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok || payload?.code !== 0) {
    throw new Error(payload?.msg || `请求失败: ${response.status}`)
  }
  return payload.data
}

async function loadConfig() {
  loading.value = true
  status.value = ''
  try {
    const data = await request(props.configApi)
    if (data) {
      config.value = { ...config.value, ...data }
    }
  } catch (error: any) {
    status.value = error.message || String(error)
    statusType.value = 'error'
  } finally {
    loading.value = false
  }
}

async function saveConfig() {
  loading.value = true
  status.value = ''
  try {
    await request(props.configApi, {
      method: 'PUT',
      body: JSON.stringify(config.value),
    })
    status.value = '设置已保存'
    statusType.value = 'success'
    if (props.onClose) {
      setTimeout(() => {
        props.onClose?.()
      }, 1000)
    }
  } catch (error: any) {
    status.value = error.message || String(error)
    statusType.value = 'error'
  } finally {
    loading.value = false
  }
}

onMounted(loadConfig)
</script>

<style scoped>
.cloudhub-push-settings {
  padding: 24px;
  max-width: 600px;
  margin: 0 auto;
}
.plugin-heading {
  margin-bottom: 24px;
}
.plugin-heading h2 {
  margin: 4px 0 8px 0;
  font-size: 20px;
  font-weight: 600;
}
.plugin-form {
  margin-top: 16px;
}
.setup-alert {
  margin-bottom: 16px;
}
.status-alert {
  margin-bottom: 16px;
}
</style>
