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
          STRM 创建、移动、重命名或删除后，将数据库中的网盘文件快照同步到 CloudHub。
        </NText>
      </header>

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
} from 'naive-ui'

const props = defineProps<{
  runtime: any
  configApi: string
  apiBase: string
  pluginName: string
}>()

const config = ref({
  node_id: '',
  base_url: '',
  api_key: '',
  public_base_url: '',
  batch_size: 500,
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
.status-alert {
  margin-bottom: 16px;
}
</style>
