package pan302plugin

import (
	"testing"

	pb "github.com/jianxcao/pan-302-plugin/gen/go/plugin/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestDecodeEventPayload(t *testing.T) {
	tests := []struct {
		name  string
		event proto.Message
		want  any
	}{
		{name: "strm", event: &pb.StrmEvent{Event: "strm.created"}, want: &pb.StrmEvent{}},
		{name: "media", event: &pb.MediaEvent{Event: "media.item.added"}, want: &pb.MediaEvent{}},
		{name: "resource", event: &pb.ResourceReadyEvent{Event: "resource.ready"}, want: &pb.ResourceReadyEvent{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload, err := proto.Marshal(tt.event)
			require.NoError(t, err)
			decoded, err := decodeEventPayload(payload)
			require.NoError(t, err)
			require.IsType(t, tt.want, decoded)
		})
	}
}

func TestDecodeEventPayloadRejectsUnknownEvent(t *testing.T) {
	payload, err := proto.Marshal(&pb.StrmEvent{Event: "unknown.created"})
	require.NoError(t, err)
	_, err = decodeEventPayload(payload)
	require.ErrorContains(t, err, "unsupported event")
}
