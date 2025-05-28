// internal/agent/encoder/encoder.go
// Package encoder converts a collected flamegraph tree into a serialised byte
// representation ready for transport by exporters.  The current implementation
// supports two formats:
//   - JSON   – the canonical D3-compatible structure (default)
//   - Proto  – marshals the JSON bytes into the FlamegraphChunk protobuf message
//     so that gateway-side decoding is easier.
//
// Adding additional formats (FlatBuffers, MessagePack) only requires
// implementing the Encoder interface and registering a constructor.
package encoder

import (
	agentpb "github.com/Voskan/flarego/internal/proto"
	"github.com/Voskan/flarego/pkg/flamegraph"
	"google.golang.org/protobuf/proto"
)

// Format enumeration.
const (
    JSON   = "json"
    Proto  = "proto"
)

// Encoder serialises a Frame to bytes.
type Encoder interface {
    Encode(root *flamegraph.Frame) ([]byte, error)
    // ContentType describes the MIME that exporters should set (optional).
    ContentType() string
}

// New returns an encoder for given format; defaults to JSON.
func New(format string) Encoder {
    switch format {
    case Proto:
        return &protoEncoder{}
    case JSON:
        fallthrough
    default:
        return &jsonEncoder{}
    }
}

// ---------------------------------------------------------------------------------------------
// JSON encoder
// ---------------------------------------------------------------------------------------------

type jsonEncoder struct{}

func (j *jsonEncoder) Encode(root *flamegraph.Frame) ([]byte, error) { return root.ToJSON() }
func (j *jsonEncoder) ContentType() string                         { return "application/json" }

// ---------------------------------------------------------------------------------------------
// Proto encoder (wraps JSON payload inside FlamegraphChunk)
// ---------------------------------------------------------------------------------------------

type protoEncoder struct{}

func (p *protoEncoder) Encode(root *flamegraph.Frame) ([]byte, error) {
    data, err := root.ToJSON()
    if err != nil {
        return nil, err
    }
    msg := &agentpb.FlamegraphChunk{Payload: data}
    return proto.Marshal(msg)
}
func (p *protoEncoder) ContentType() string { return "application/x-protobuf" }
