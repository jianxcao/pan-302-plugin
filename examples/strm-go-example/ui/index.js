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
    NSpace,
    NText,
  } = naive

  const SettingsForm = defineComponent({
    name: 'GoExampleSettings',
    setup() {
      const state = reactive({
        config: {
          example_val: '',
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
        h('section', { class: 'go-example-settings' }, [
          h('header', { class: 'plugin-heading' }, [
            h('div', [
              h(NText, { depth: 3, class: 'plugin-kicker' }, () => 'Go Example Plugin'),
              h('h2', 'Go 极简示例插件配置'),
            ]),
            h(
              NText,
              { depth: 3 },
              () => '这是一个基于 Go 语言 Wasm 的极简事件监听 Demo，无需本地打包环境即可开发前端面板。',
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
              h(NFormItem, { label: '测试参数输入', path: 'example_val' }, () =>
                textInput('example_val', {
                  placeholder: '输入点什么以测试保存配置',
                  autocomplete: 'off',
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
    name: 'GoExamplePluginRoot',
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
