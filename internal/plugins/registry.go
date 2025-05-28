// internal/plugins/registry.go
// Runtime plugin registry.  Allows dynamic discovery and execution of plugin
// callbacks at runtime without hard-coding them in the core binaries.  Go’s
// native plugin support (plugin.Open) only works on Linux/macOS and requires
// plugins to be built with the exact same Go version and compiler flags; this
// registry abstracts those details and offers a fall‑back "static import" mode
// for platforms where .so loading is unavailable.
//
// Plugin authors implement the Plugin interface and call Register() in their
// plugin’s init() function.  At runtime, the Gateway or Agent can iterate over
// all plugins of a given kind and execute hooks.
package plugins

import (
	"plugin"
	"sync"
)

// Kind classifies plugin purpose so callers can filter quickly.
// Examples: "encoder", "sampler", "exporter".
// Custom kinds are allowed; collisions are prevented by separate maps.
type Kind string

// Plugin is the minimal contract a FlareGo plugin must satisfy.
type Plugin interface {
    Kind() Kind        // category
    Name() string      // human‑readable unique name
    // Init is invoked once after registration.  The plugin can perform setup
    // and return an opaque handle for future use.  Returning error aborts
    // registration.  The handle may be nil if unused.
    Init() (any, error)
}

// registry is a global map: kind → name → plugin instance.
var (
    regMu    sync.RWMutex
    registry = make(map[Kind]map[string]Plugin)
)

// Register adds p to the global registry.  Should be called from plugin init().
// Duplicate (kind,name) pair panics to surface programmer error early.
func Register(p Plugin) {
    regMu.Lock()
    defer regMu.Unlock()
    kindMap, ok := registry[p.Kind()]
    if !ok {
        kindMap = make(map[string]Plugin)
        registry[p.Kind()] = kindMap
    }
    if _, exists := kindMap[p.Name()]; exists {
        panic("plugins: duplicate plugin " + string(p.Kind()) + "/" + p.Name())
    }
    if _, err := p.Init(); err != nil {
        panic("plugins: init failed for " + p.Name() + ": " + err.Error())
    }
    kindMap[p.Name()] = p
}

// ByKind returns a slice of plugins matching kind.
func ByKind(k Kind) []Plugin {
    regMu.RLock()
    defer regMu.RUnlock()
    m := registry[k]
    out := make([]Plugin, 0, len(m))
    for _, p := range m {
        out = append(out, p)
    }
    return out
}

// LoadShared dynamically loads a Go plugin (.so) file and expects it to call
// plugins.Register() in its init() function.  On unsupported platforms or if
// the plugin Open fails, an error is returned.
func LoadShared(path string) error {
    _, err := plugin.Open(path)
    return err
}
