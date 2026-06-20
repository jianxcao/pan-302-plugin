package pan302plugin

import (
	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

// HTTP 命名空间，提供借由宿主网络代理发送 HTTP 请求的接口（用以突破沙箱网络隔离）
var HTTP = httpAPI{}

type httpAPI struct{}

// Request 通过宿主的网络代理发起外部网络 HTTP 请求
func (httpAPI) Request(request *pb.HTTPRequestArgs) (*pb.HTTPResponseData, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_HttpRequest{HttpRequest: request},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetHttpResponse(), nil
}
