// internal/agent/encoder/proto.go
// Helper utilities that wrap the generic encoder logic for callers that
// explicitly need a protobuf‐marshalled FlamegraphChunk.  This file keeps the
// public surface of internal/agent/encoder minimal while avoiding import
// cycles in packages that prefer direct helper functions over the generic
// Encoder interface.
package encoder

import (
	agentpb "github.com/Voskan/flarego/internal/proto"
	"github.com/Voskan/flarego/pkg/flamegraph"
	"google.golang.org/protobuf/proto"
)

// EncodeToProto marshals root into an agentpb.FlamegraphChunk protobuf binary.
// It is functionally identical to New(Proto).Encode but avoids an allocation
// of the protoEncoder instance when used in hot paths.
func EncodeToProto(root *flamegraph.Frame) ([]byte, error) {
    if root == nil {
        return nil, nil
    }
    jsonBytes, err := root.ToJSON()
    if err != nil {
        return nil, err
    }
    msg := &agentpb.FlamegraphChunk{Payload: jsonBytes}
    return proto.Marshal(msg)
}
