let app = null
let currentProps = {}

function requireRuntime(props) {
  const runtime = props.runtime
  if (!runtime?.vue || !runtime?.naive) {
    throw new Error('插件 UI 缺少宿主提供的 Vue/Naive UI 运行时')
  }
  return runtime
}

async function request(path, options = {}) {
  const credentials = localStorage.getItem('pan302_auth')
  const response = await fetch(`${currentProps.apiBase}${path}`, {
    ...options,
    headers: {
      'content-type': 'application/json',
      ...(credentials ? { Authorization: `Basic ${credentials}` } : {}),
      ...(options.headers || {}),
    },
  })
  const payload = await response.json().catch(() => ({}))
  if (!response.ok || payload?.code > 0) {
    throw new Error(payload?.msg || payload?.error || `请求失败: ${response.status}`)
  }
  return payload?.code === 0 && 'data' in payload ? payload.data : payload
}

function createRootComponent(runtime) {
  const { vue, naive } = runtime
  const { computed, defineComponent, h, onMounted, reactive } = vue
  const {
    NAlert,
    NButton,
    NConfigProvider,
    NForm,
    NFormItem,
    NInput,
    NSelect,
    NSpace,
    NSwitch,
    NText,
  } = naive

  const TestForm = defineComponent({
    name: 'StrmExampleSettings',
    setup() {
      const state = reactive({
        drivers: [],
        driverId: null,
        taskName: '',
        cloudPath: '',
        force: false,
        loading: false,
        status: '准备就绪',
        statusType: 'info',
      })

      const driverOptions = computed(() =>
        state.drivers.map(driver => ({
          label: `${driver.name} (${driver.type})`,
          value: driver.id,
        })),
      )

      async function loadDrivers() {
        state.loading = true
        state.status = ''
        try {
          const drivers = await request('/drivers')
          state.drivers = Array.isArray(drivers) ? drivers : []
          if (!state.drivers.some(driver => driver.id === state.driverId)) {
            state.driverId = state.drivers[0]?.id || null
          }
          state.status = `当前可用 Driver：${state.drivers.length} 个`
          state.statusType = 'success'
        } catch (error) {
          state.status = error instanceof Error ? error.message : String(error)
          state.statusType = 'error'
        } finally {
          state.loading = false
        }
      }

      async function run(operation) {
        if (!state.driverId || !state.cloudPath) {
          state.status = '请选择 Driver 并填写云盘文件路径'
          state.statusType = 'warning'
          return
        }
        state.loading = true
        state.status = ''
        try {
          const result = await request(`/${operation}`, {
            method: 'POST',
            body: JSON.stringify({
              driverId: state.driverId,
              cloudPath: state.cloudPath,
              taskName: state.taskName,
              force: state.force,
              idempotencyKey: crypto.randomUUID(),
            }),
          })
          state.status = typeof result === 'string' ? result : JSON.stringify(result, null, 2)
          state.statusType = 'success'
        } catch (error) {
          state.status = error instanceof Error ? error.message : String(error)
          state.statusType = 'error'
        } finally {
          state.loading = false
        }
      }

      onMounted(loadDrivers)

      return () =>
        h('section', { class: 'strm-example-settings' }, [
          h('header', { class: 'plugin-heading' }, [
            h('div', [
              h(NText, { depth: 3, class: 'plugin-kicker' }, () => 'STRM Example'),
              h('h2', 'STRM 生成与删除测试'),
            ]),
            h(NText, { depth: 3 }, () => '验证 Driver 查询以及宿主 STRM 写入、删除能力。'),
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
                () => h('pre', { class: 'plugin-result' }, state.status),
              )
            : null,
          h(
            NForm,
            {
              labelPlacement: 'top',
              disabled: state.loading,
              class: 'plugin-form',
            },
            () => [
              h(NFormItem, { label: 'Driver' }, () =>
                h(NSelect, {
                  value: state.driverId,
                  options: driverOptions.value,
                  placeholder: '请选择 Driver',
                  filterable: true,
                  'onUpdate:value': value => {
                    state.driverId = value
                  },
                }),
              ),
              h(NFormItem, { label: '任务名（可选）' }, () =>
                h(NInput, {
                  value: state.taskName,
                  'onUpdate:value': value => {
                    state.taskName = value
                  },
                }),
              ),
              h(NFormItem, { label: '云盘文件路径' }, () =>
                h(NInput, {
                  value: state.cloudPath,
                  placeholder: '/media/movie.mkv',
                  'onUpdate:value': value => {
                    state.cloudPath = value
                  },
                }),
              ),
              h(NFormItem, { label: '覆盖已存在 STRM' }, () =>
                h(NSwitch, {
                  value: state.force,
                  'onUpdate:value': value => {
                    state.force = value
                  },
                }),
              ),
              h(NSpace, { justify: 'space-between' }, () => [
                h(NButton, { loading: state.loading, onClick: loadDrivers }, () => '刷新 Driver'),
                h(NSpace, null, () => [
                  h(
                    NButton,
                    { loading: state.loading, onClick: () => run('delete') },
                    () => '删除 STRM',
                  ),
                  h(
                    NButton,
                    {
                      type: 'primary',
                      loading: state.loading,
                      onClick: () => run('write'),
                    },
                    () => '生成 STRM',
                  ),
                ]),
              ]),
            ],
          ),
        ])
    },
  })

  return defineComponent({
    name: 'StrmExamplePluginRoot',
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
          () => h(TestForm),
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
