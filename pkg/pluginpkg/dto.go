package pluginpkg

const (
	// ABIVersion 插件和宿主通信的数据契约 ABI 大版本。不匹配时宿主拒绝加载。
	ABIVersion     = "pan302-plugin-abi/v2"
	// UIProtocol 前端 UI 扩展的通信协议大版本，用于规范宿主和插件前端交互。
	UIProtocol     = "pan302-plugin-ui/v1"
	// HostModuleName 宿主（主程序）向 Wasm 插件沙箱中导出的核心模块名称（Wasm 命名空间）。
	HostModuleName = "pan302_v1"
	// HostCallName 插件调用宿主核心方法时的导入函数名（Wasm Guest -> Host 通信网关）。
	HostCallName   = "host_call"
	// WASIModuleName WASI 核心系统调用规范在 Wasm 中导入的模块段名。
	WASIModuleName = "wasi_snapshot_preview1"

	// ExportAlloc Wasm 插件导出的内存分配函数。宿主通过它为插件申请虚拟内存空间以传递数据。
	ExportAlloc      = "pan302_alloc"
	// ExportFree Wasm 插件导出的内存释放函数。宿主用其释放传入数据后不再使用的插件内存，防止泄露。
	ExportFree       = "pan302_free"
	// ExportInit Wasm 插件导出的初始化入口函数。宿主实例化插件后会首先调用此钩子进行初始化。
	ExportInit       = "pan302_init"
	// ExportOnEvent Wasm 插件导出的事件监听处理函数。当宿主侧发生网盘文件变动时通过此方法投递事件。
	ExportOnEvent    = "pan302_on_event"
	// ExportHandleHTTP Wasm 插件导出的 Web 路由请求处理函数。当有用户访问插件配置的 Web 路由时调用。
	ExportHandleHTTP = "pan302_handle_http"
	// ExportMigrate Wasm 插件导出的配置升级迁移函数。当插件版本更新导致配置 Schema 改变时调用。
	ExportMigrate    = "pan302_migrate"
)

var CoreHostOperations = map[string]struct{}{
	"log.write":       {},
	"config.read":     {},
	"config.write":    {},
	"driver.list":     {},
	"driver.read":     {},
	"driver.link":     {},
	"driver.mkdir":    {},
	"driver.rename":   {},
	"driver.move":     {},
	"driver.copy":     {},
	"driver.delete":   {},
	"strm.write":      {},
	"strm.delete":     {},
	"route.handle":    {},
	"event.strm.read": {},
	"notify.send":     {},
	"http.request":    {},
}
