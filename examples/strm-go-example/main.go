package main

import (
	"encoding/json"
	"fmt"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	pan302plugin "github.com/jianxcao/pan-302-plugin/sdk/go"
)

type ExampleConfig struct {
	ExampleVal string `json:"example_val"`
}

func main() {}

//go:wasmexport pan302_alloc
func pan302Alloc(size uint32) uint32 {
	return pan302plugin.Allocate(size)
}

//go:wasmexport pan302_free
func pan302Free(ptr, _ uint32) {
	pan302plugin.Free(ptr)
}

//go:wasmexport pan302_init
func pan302Init(ptr, length uint32) uint64 {
	var request pb.InitRequest
	if err := pan302plugin.DecodeRequest(ptr, length, &request); err != nil {
		return errorResponse(err)
	}
	pan302plugin.Logger.Info("Go 示例插件已初始化并加载成功！", nil)
	return successResponse()
}

//go:wasmexport pan302_on_event
func pan302OnEvent(ptr, length uint32) uint64 {
	var event pb.StrmEvent
	if err := pan302plugin.DecodeRequest(ptr, length, &event); err != nil {
		return errorResponse(err)
	}

	// 打印收到的事件日志，以此证明监听到了宿主事件
	pan302plugin.Logger.Info(fmt.Sprintf("[Go Example Plugin] 监听到系统事件: %s", event.Event), map[string]string{
		"eventId": event.EventId,
	})

	// 演示读取一下自己的配置
	configResp, err := pan302plugin.Config.Read()
	if err == nil && configResp.Config != nil {
		configJSON, _ := configResp.Config.MarshalJSON()
		var config ExampleConfig
		if err := json.Unmarshal(configJSON, &config); err == nil {
			pan302plugin.Logger.Info(fmt.Sprintf("[Go Example Plugin] 读取到的配置值: example_val=%s", config.ExampleVal), nil)
		}
	}

	return successResponse()
}

func successResponse() uint64 {
	return pan302plugin.EncodeResponse(&pb.LifecycleResponse{Ok: true})
}

func errorResponse(err error) uint64 {
	return pan302plugin.EncodeResponse(&pb.LifecycleResponse{Ok: false, Error: err.Error()})
}
