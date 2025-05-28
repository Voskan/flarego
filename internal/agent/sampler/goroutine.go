// internal/agent/sampler/goroutine.go
// GoroutineSampler captures full stack traces of *all* goroutines at a fixed
// frequency and feeds them into flamegraph.Builder.  Each goroutine produces a
// Sample whose Stack is the sequence of function names bottom→top (root first)
// with the Weight equal to 1, so the flame graph visualises how many
// goroutines share the same call path in real time.
//
// Implementation details:
//   - Uses runtime.GoroutineProfile which is cheaper than stopping the world
//     via runtime.Stack.  It allocates a []runtime.StackRecord buffer that is
//     grown to fit (doubling strategy).
//   - For each StackRecord we translate PCs into function names via
//     runtime.FuncForPC and drop wrapper frames such as "runtime.goexit".
//   - The sampler intentionally **does not** de‐duplicate stacks before adding
//     to the builder because the builder’s tree structure naturally merges
//     identical paths.
package sampler

import (
	"runtime"
	"strings"
	"time"

	"github.com/Voskan/flarego/pkg/flamegraph"
)

// GoroutineSampler polls the runtime for goroutine stacks.
type GoroutineSampler struct {
    b  *flamegraph.Builder
    hz int

    quit chan struct{}
    done chan struct{}
}

// NewGoroutineSampler constructs a sampler clamped to [10, 200] Hz.
func NewGoroutineSampler(b *flamegraph.Builder, hz int) *GoroutineSampler {
    if hz < 10 {
        hz = 10
    }
    if hz > 200 {
        hz = 200
    }
    return &GoroutineSampler{
        b:    b,
        hz:   hz,
        quit: make(chan struct{}),
        done: make(chan struct{}),
    }
}

// Start launches the background goroutine.
func (s *GoroutineSampler) Start() {
    select {
    case <-s.done:
        return // cannot restart
    default:
    }
    go s.loop()
}

func (s *GoroutineSampler) loop() {
    defer close(s.done)

    interval := time.Second / time.Duration(s.hz)
    ticker := time.NewTicker(interval)
    defer ticker.Stop()

    buf := make([]runtime.StackRecord, 256)

    for {
        select {
        case <-ticker.C:
            // Grow buffer until it fits.
            for {
                n, ok := runtime.GoroutineProfile(buf)
                if ok {
                    buf = buf[:n]
                    break
                }
                buf = make([]runtime.StackRecord, len(buf)*2)
            }

            for _, rec := range buf {
                pcs := rec.Stack()
                if len(pcs) == 0 {
                    continue
                }
                stack := pcSliceToNames(pcs)
                if len(stack) == 0 {
                    continue
                }
                s.b.Add(flamegraph.Sample{Stack: stack, Weight: 1})
            }
        case <-s.quit:
            return
        }
    }
}

// Stop signals the sampler to finish.
func (s *GoroutineSampler) Stop() {
    select {
    case <-s.done:
        return
    default:
        close(s.quit)
        <-s.done
    }
}

//--------------------------------------------------------------------
// helpers
//--------------------------------------------------------------------

func pcSliceToNames(pcs []uintptr) []string {
    names := make([]string, 0, len(pcs))
    // Iterate from deepest (root) to shallowest (leaf) so the flamegraph tree
    // orders frames naturally top‐down.
    for i := len(pcs) - 1; i >= 0; i-- {
        if pcs[i] == 0 {
            continue
        }
        fn := runtime.FuncForPC(pcs[i] - 1) // -1 to within function
        if fn == nil {
            continue
        }
        name := trimPkgPath(fn.Name())
        if name == "runtime.goexit" || name == "runtime.main" {
            continue // drop runtime wrappers
        }
        names = append(names, name)
    }
    return names
}

// trimPkgPath removes the full import path, leaving pkg.Func.
func trimPkgPath(full string) string {
    // runtime.Func.Name returns "github.com/user/pkg.function".
    if idx := strings.LastIndex(full, "/"); idx != -1 {
        return full[idx+1:]
    }
    return full
}
