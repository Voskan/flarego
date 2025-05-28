// pkg/trace/reader.go
// Reader utilities for parsing FlareGo runtime traces from either the
// protobuf‐encoded TraceBatch (defined in internal/proto/trace.proto) or a
// newline‐delimited JSON stream of Event objects.  The goal is to make it very
// simple for tooling (CLI diff, offline analysis, tests) to iterate over
// events without duplicating deserialisation boilerplate.
package trace

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"

	agentpb "github.com/Voskan/flarego/internal/proto"
	"google.golang.org/protobuf/proto"
)

// Format enumerates supported on‐disk encodings.
type Format int

const (
    // AutoDetect peeks at the first few bytes to choose between Proto or JSON.
    AutoDetect Format = iota
    Proto
    NDJSON // newline‐delimited JSON [{"ts":...}\n]
)

// ErrUnknownFormat returned when AutoDetect fails.
var ErrUnknownFormat = errors.New("trace: unknown format")

// ReadAll consumes r and returns the decoded events slice.
// When format == AutoDetect it sniffs first byte – if 0x0a (varint field tag)
// assume protobuf, else JSON.
func ReadAll(r io.Reader, format Format) ([]Event, error) {
    if format == AutoDetect {
        br := bufio.NewReader(r)
        b, err := br.Peek(1)
        if err != nil {
            return nil, err
        }
        if b[0] == 0x0a { // likely protobuf – field 1, length‐delimited
            format = Proto
        } else {
            format = NDJSON
        }
        r = br // reuse buffered reader
    }

    switch format {
    case Proto:
        data, err := io.ReadAll(r)
        if err != nil {
            return nil, err
        }
        var batch agentpb.TraceBatch
        if err := proto.Unmarshal(data, &batch); err != nil {
            return nil, err
        }
        return fromProto(batch.Events), nil

    case NDJSON:
        return readNDJSON(r)

    default:
        return nil, ErrUnknownFormat
    }
}

// fromProto maps generated protobuf events to internal model.
func fromProto(in []*agentpb.Event) []Event {
    out := make([]Event, len(in))
    for i, ev := range in {
        out[i] = Event{
            Ts:    ev.Ts,
            G:     ev.G,
            P:     ev.P,
            Type:  EventType(ev.Type),
            Value: ev.Value,
            // stack is []uint64 in proto; convert to []uintptr.
            Stack: uint64SliceToPtr(ev.Stack),
        }
    }
    return out
}

func uint64SliceToPtr(in []uint64) []uintptr {
    if len(in) == 0 {
        return nil
    }
    out := make([]uintptr, len(in))
    for i, v := range in {
        out[i] = uintptr(v)
    }
    return out
}

// readNDJSON decodes newline‐delimited JSON stream.
func readNDJSON(r io.Reader) ([]Event, error) {
    var events []Event
    scanner := bufio.NewScanner(r)
    for scanner.Scan() {
        var ev Event
        if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
            return nil, err
        }
        events = append(events, ev)
    }
    if err := scanner.Err(); err != nil {
        return nil, err
    }
    return events, nil
}
