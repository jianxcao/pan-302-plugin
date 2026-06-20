let app = null
let currentProps = {}

function requireRuntime(props) {
  const runtime = props.runtime
  if (!runtime?.vue || !runtime?.naive) {
    throw new Error('插件 UI 缺少宿主提供的 Vue/Naive UI 运行时')
  }
  return runtime
}

function authHeaders() {
  const credentials = localStorage.getItem('pan302_auth')
  return credentials ? { Authorization: `Basic ${credentials}` } : {}
}

async function request(url, options = {}) {
  const response = await fetch(url, {
    ...options,
    headers: {
      'content-type': 'application/json',
      ...authHeaders(),
      ...(options.headers || {}),
    },
  })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok || payload?.code !== 0) {
    throw new Error(payload?.msg || `请求失败: ${response.status}`)
  }
  return payload.data
}

function createRootComponent(runtime) {
  const { vue, naive } = runtime
  const { defineComponent, h, onMounted, reactive } = vue
  const {
    NAlert,
    NButton,
    NConfigProvider,
    NForm,
    NFormItem,
    NInput,
    NInputNumber,
    NSpace,
    NText,
  } = naive

  const SettingsForm = defineComponent({
    name: 'CloudHubPushSettings',
    setup() {
      const state = reactive({
        config: {
          node_id: '',
          base_url: '',
          api_key: '',
          public_base_url: '',
          batch_size: 500,
        },
        loading: false,
        status: '',
        statusType: 'info',
      })

      async function loadConfig() {
        state.loading = true
        state.status = ''
        try {
          Object.assign(state.config, (await request(currentProps.configApi)) || {})
        } catch (error) {
          state.status = error instanceof Error ? error.message : String(error)
          state.statusType = 'error'
        } finally {
          state.loading = false
        }
      }

      async function saveConfig() {
        state.loading = true
        state.status = ''
        try {
          await request(currentProps.configApi, {
            method: 'PUT',
            body: JSON.stringify(state.config),
          })
          state.status = '设置已保存'
          state.statusType = 'success'
        } catch (error) {
          state.status = error instanceof Error ? error.message : String(error)
          state.statusType = 'error'
        } finally {
          state.loading = false
        }
      }

      onMounted(loadConfig)

      const textInput = (key, props = {}) =>
        h(NInput, {
          value: state.config[key],
          ...props,
          'onUpdate:value': value => {
            state.config[key] = value
          },
        })

      return () =>
        h('section', { class: 'cloudhub-push-settings' }, [
          h('header', { class: 'plugin-heading' }, [
            h('div', [
              h(NText, { depth: 3, class: 'plugin-kicker' }, () => 'CloudHub Push'),
              h('h2', 'CloudHub 资源推送'),
            ]),
            h(
              NText,
              { depth: 3 },
              () => 'STRM 创建或删除后，将数据库中的网盘文件快照同步到 CloudHub。',
            ),
          ]),
          state.status
            ? h(
                NAlert,
                {
                  type: state.statusType,
                  closable: true,
                  onClose: () => {
                    state.status = ''
                  },
                },
                () => state.status,
              )
            : null,
          h(
            NForm,
            {
              model: state.config,
              labelPlacement: 'top',
              disabled: state.loading,
              class: 'plugin-form',
            },
            () => [
              h(NFormItem, { label: '节点 ID', path: 'node_id' }, () =>
                textInput('node_id', {
                  placeholder: '例如 media302',
                  autocomplete: 'off',
                }),
              ),
              h(NFormItem, { label: 'CloudHub API URL', path: 'base_url' }, () =>
                textInput('base_url', { placeholder: 'https://cloudhub.example.com' }),
              ),
              h(NFormItem, { label: '节点 API Key', path: 'api_key' }, () =>
                textInput('api_key', {
                  type: 'password',
                  showPasswordOn: 'click',
                  autocomplete: 'new-password',
                }),
              ),
              h(NFormItem, { label: '节点公开 URL（可选）', path: 'public_base_url' }, () =>
                textInput('public_base_url', { placeholder: 'https://pan.example.com' }),
              ),
              h(NFormItem, { label: '单批数量', path: 'batch_size' }, () =>
                h(NInputNumber, {
                  value: state.config.batch_size,
                  min: 1,
                  max: 500,
                  'onUpdate:value': value => {
                    state.config.batch_size = value || 500
                  },
                }),
              ),
              h(NSpace, { justify: 'end' }, () =>
                h(
                  NButton,
                  {
                    type: 'primary',
                    loading: state.loading,
                    onClick: saveConfig,
                  },
                  () => '保存设置',
                ),
              ),
            ],
          ),
        ])
    },
  })

  return defineComponent({
    name: 'CloudHubPushPluginRoot',
    setup() {
      return () =>
        h(
          NConfigProvider,
          {
            theme: runtime.theme.value,
            themeOverrides: runtime.themeOverrides.value,
            locale: runtime.locale,
            dateLocale: runtime.dateLocale,
          },
          () => h(SettingsForm),
        )
    },
  })
}

export default {
  mount(container, props = {}) {
    const runtime = requireRuntime(props)
    currentProps = props
    app = runtime.vue.createApp(createRootComponent(runtime))
    app.mount(container)
  },

  update(props = {}) {
    currentProps = { ...currentProps, ...props }
  },

  unmount(container) {
    app?.unmount()
    app = null
    currentProps = {}
    container.replaceChildren()
  },
}
