package sqltrace

import (
	"fmt"

	"github.com/Voskan/flarego/internal/plugins"
)

// Plugin struct
type SQLTracePlugin struct{}

// Kind returns the plugin kind
func (p *SQLTracePlugin) Kind() plugins.Kind {
	return "sampler"
}

// Name returns the plugin name
func (p *SQLTracePlugin) Name() string {
	return "sqltrace"
}

// Init is called when the plugin is registered
func (p *SQLTracePlugin) Init() (any, error) {
	fmt.Println("[sqltrace] SQLTracePlugin initialized")
	return nil, nil
}

// init automatically registers the plugin when the package is imported
func init() {
	plugins.Register(&SQLTracePlugin{})
}
