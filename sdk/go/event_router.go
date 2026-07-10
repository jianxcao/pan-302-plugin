package pan302plugin

import (
	"fmt"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	"google.golang.org/protobuf/proto"
)

type EventRouter struct {
	onStrm          func(*pb.StrmEvent) error
	onMedia         func(*pb.MediaEvent) error
	onResourceReady func(*pb.ResourceReadyEvent) error
}

func NewEventRouter() *EventRouter {
	return &EventRouter{}
}

func (r *EventRouter) OnStrm(handler func(*pb.StrmEvent) error) *EventRouter {
	r.onStrm = handler
	return r
}

func (r *EventRouter) OnMedia(handler func(*pb.MediaEvent) error) *EventRouter {
	r.onMedia = handler
	return r
}

func (r *EventRouter) OnResourceReady(handler func(*pb.ResourceReadyEvent) error) *EventRouter {
	r.onResourceReady = handler
	return r
}

func (r *EventRouter) Dispatch(ptr, length uint32) (string, error) {
	event, err := DecodeEventRequest(ptr, length)
	if err != nil {
		return "", err
	}
	return r.DispatchEvent(event)
}

func (r *EventRouter) DispatchEvent(event proto.Message) (string, error) {
	switch value := event.(type) {
	case *pb.StrmEvent:
		return dispatchTypedEvent(value.EventId, value.Event, r.onStrm, value)
	case *pb.MediaEvent:
		return dispatchTypedEvent(value.EventId, value.Event, r.onMedia, value)
	case *pb.ResourceReadyEvent:
		return dispatchTypedEvent(value.EventId, value.Event, r.onResourceReady, value)
	default:
		return "", fmt.Errorf("不支持的插件事件类型 %T", event)
	}
}

func dispatchTypedEvent[T proto.Message](eventID, eventName string, handler func(T) error, event T) (string, error) {
	if handler == nil {
		return eventID, fmt.Errorf("事件 %q 未注册处理函数", eventName)
	}
	return eventID, handler(event)
}
