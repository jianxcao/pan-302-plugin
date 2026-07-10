package pan302plugin

import (
	"testing"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestEventRouterDispatchesRegisteredHandler(t *testing.T) {
	called := false
	router := NewEventRouter().OnResourceReady(func(event *pb.ResourceReadyEvent) error {
		called = true
		require.Equal(t, "ready-1", event.EventId)
		return nil
	})

	eventID, err := router.DispatchEvent(&pb.ResourceReadyEvent{EventId: "ready-1", Event: "resource.ready"})
	require.NoError(t, err)
	require.Equal(t, "ready-1", eventID)
	require.True(t, called)
}

func TestEventRouterRejectsUnregisteredHandler(t *testing.T) {
	eventID, err := NewEventRouter().DispatchEvent(&pb.MediaEvent{EventId: "media-1", Event: "media.item.added"})
	require.Equal(t, "media-1", eventID)
	require.ErrorContains(t, err, "未注册")
}

func TestEventRouterSupportsAllEventFamilies(t *testing.T) {
	router := NewEventRouter().
		OnStrm(func(*pb.StrmEvent) error { return nil }).
		OnMedia(func(*pb.MediaEvent) error { return nil }).
		OnResourceReady(func(*pb.ResourceReadyEvent) error { return nil })

	for _, event := range []proto.Message{
		&pb.StrmEvent{EventId: "strm-1", Event: "strm.created"},
		&pb.MediaEvent{EventId: "media-1", Event: "media.item.added"},
		&pb.ResourceReadyEvent{EventId: "ready-1", Event: "resource.ready"},
	} {
		_, err := router.DispatchEvent(event)
		require.NoError(t, err)
	}
}
