package pan302plugin

import (
	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
)

// Logger 命名空间，向宿主日志系统写入结构化日志
var Logger = loggerAPI{}

type loggerAPI struct{}

// Write 写入一条指定等级的日志，支持附加结构化字段
func (loggerAPI) Write(level, message string, fields map[string]string) {
	_, _ = call(&pb.HostRequest{
		Id: nextID(),
		Request: &pb.HostRequest_LogWrite{
			LogWrite: &pb.LogWriteRequest{
				Level:   level,
				Message: message,
				Fields:  fields,
			},
		},
	})
}

// Info 写入一条 Info 等级日志
func (l loggerAPI) Info(message string, fields map[string]string) {
	l.Write("info", message, fields)
}

// Warn 写入一条 Warn 等级日志
func (l loggerAPI) Warn(message string, fields map[string]string) {
	l.Write("warn", message, fields)
}

// Error 写入一条 Error 等级日志
func (l loggerAPI) Error(message string, fields map[string]string) {
	l.Write("error", message, fields)
}

// Notify 命名空间，通过宿主预设通道发送通知
var Notify = notifyAPI{}

type notifyAPI struct{}

// Send 发送系统级通知（如推送）
func (notifyAPI) Send(req *pb.NotifySendRequest) (*pb.NotifySendResponse, error) {
	resp, err := call(&pb.HostRequest{
		Id:      nextID(),
		Request: &pb.HostRequest_NotifySend{NotifySend: req},
	})
	if err != nil {
		return nil, err
	}
	return resp.GetNotifySend(), nil
}
