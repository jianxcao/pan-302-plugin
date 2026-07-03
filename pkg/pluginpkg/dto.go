package pluginpkg

// 这里定义了 pan-302 插件系统与 Wasm 虚拟机宿主（Host）之间的底层 ABI 通信协议和内存共享契约。

const (
	// ABIVersion 插件和宿主通信的数据契约 ABI 大版本。
	// 宿主在加载 Wasm 插件时会首先校验此版本号，如果插件编译时依赖的 ABI 版本与宿主当前版本不匹配，宿主将拒绝加载该插件。
	ABIVersion     = "pan302-plugin-abi/v2"

	// UIProtocol 前端 UI 扩展的通信协议大版本。
	// 用于规范宿主管理后台与插件自定义前端界面（如果有）之间的数据交互协议。
	UIProtocol     = "pan302-plugin-ui/v1"

	// HostModuleName 宿主（主程序）向 Wasm 插件沙箱中导出的核心模块名称（Wasm 命名空间）。
	// 在 Wasm 中，插件导入（import）外部函数时，必须指定模块名为 "pan302_v1"。
	HostModuleName = "pan302_v1"

	// HostCallName 插件调用宿主核心方法时的导入函数名（Wasm Guest -> Host 的唯一通信网关）。
	// 插件通过调用 `pan302_v1.host_call`，将请求操作名、参数内存指针和长度传递给宿主，从而使用宿主提供的各种系统级能力（如网络请求、数据库读写、strm 文件生成等）。
	HostCallName   = "host_call"

	// WASIModuleName WASI 核心系统调用规范在 Wasm 中导入的模块段名。
	// 提供标准的系统调用支持（例如时间、随机数、环境变量等标准接口）。
	WASIModuleName = "wasi_snapshot_preview1"

	// --- 以下为 Wasm 插件必须向宿主导出（export）的 FFI 函数名称 ---

	// ExportAlloc Wasm 插件导出的内存分配函数。
	// 由于 Wasm 的线性内存是隔离的，宿主如果想向插件传递数据（如请求参数、配置等），
	// 必须先调用插件的 `pan302_alloc` 申请一块内存，然后向该地址写入数据，最后把指针传给插件的处理函数。
	ExportAlloc      = "pan302_alloc"

	// ExportFree Wasm 插件导出的内存释放函数。
	// 宿主在使用完插件返回的数据或完成数据写入后，调用 `pan302_free` 释放对应的内存空间，避免插件沙箱发生内存泄露。
	ExportFree       = "pan302_free"

	// ExportInit Wasm 插件导出的初始化入口函数。
	// 宿主在实例化插件并准备就绪后，会首先调用 `pan302_init` 钩子，传递初始化配置，通知插件开始工作。
	ExportInit       = "pan302_init"

	// ExportOnEvent Wasm 插件导出的事件监听处理函数。
	// 当宿主侧发生网盘文件变动、同步状态改变等各种系统事件时，会通过此函数将事件 payload 投递给插件。
	ExportOnEvent    = "pan302_on_event"

	// ExportHandleHTTP Wasm 插件导出的 Web 路由请求处理函数。
	// 当用户在宿主后台访问该插件注册的 Web 路由（如配置界面 API）时，宿主会将 HTTP 请求报文打包后，调用此函数交给插件处理。
	ExportHandleHTTP = "pan302_handle_http"

	// ExportMigrate Wasm 插件导出的配置升级迁移函数。
	// 当插件版本更新、配置 Schema 结构发生改变时，宿主会在启动时调用该函数，允许插件将旧配置格式自动迁移到新格式。
	ExportMigrate    = "pan302_migrate"
)

// CoreHostOperations 列出了宿主向 Wasm 插件开放的可用核心操作（系统调用）。
// 插件通过 `host_call` 传入这些操作标识，调用宿主提供的底层功能。
var CoreHostOperations = map[string]struct{}{
	"log.write":       {}, // 向宿主日志系统写入一条日志（支持 stdout/file 等配置）
	"config.read":     {}, // 读取插件的持久化配置数据
	"config.write":    {}, // 写入/更新插件的持久化配置数据
	"driver.list":     {}, // 获取宿主中当前已注册的网盘驱动列表
	"driver.read":     {}, // 获取指定网盘驱动的详细元数据信息
	"driver.link":     {}, // 请求解析某个网盘文件的直链 (302 播放地址)
	"driver.mkdir":    {}, // 在网盘驱动中创建文件夹
	"driver.rename":   {}, // 重命名网盘中的文件或文件夹
	"driver.move":     {}, // 移动网盘中的文件或文件夹
	"driver.copy":     {}, // 复制网盘中的文件或文件夹
	"driver.delete":   {}, // 删除网盘中的文件或文件夹
	"strm.write":      {}, // 请求宿主在本地生成 STRM 播放文件
	"strm.delete":     {}, // 请求宿主删除本地的 STRM 播放文件
	"route.handle":    {}, // 向宿主注册插件专有的 Web API 路由路径
	"event.strm.read": {}, // 读取/拉取 STRM 生成相关的事件
	"event.media.read": {}, // 读取/拉取媒体服务器 webhook 相关的事件
	"notify.send":     {}, // 通过宿主通道发送系统级通知（如 Telegram, Webhook 等）
	"http.request":    {}, // 利用宿主的网络代理发起外部网络 HTTP 请求（突破沙箱限制）
	"media.server_config.read": {}, // 读取宿主配置的 Emby/Jellyfin 连接信息
}
