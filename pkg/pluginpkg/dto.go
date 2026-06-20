package pluginpkg

const (
	ABIVersion     = "pan302-plugin-abi/v2"
	UIProtocol     = "pan302-plugin-ui/v1"
	HostModuleName = "pan302_v1"
	HostCallName   = "host_call"
	WASIModuleName = "wasi_snapshot_preview1"

	ExportAlloc      = "pan302_alloc"
	ExportFree       = "pan302_free"
	ExportInit       = "pan302_init"
	ExportOnEvent    = "pan302_on_event"
	ExportHandleHTTP = "pan302_handle_http"
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
