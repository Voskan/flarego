// cmd/flarego/version.go
// Implements the `flarego version` sub-command, which prints build metadata
// injected via pkg/version.  Supports an optional --json flag for machine
// consumption.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/Voskan/flarego/pkg/version"
)

// newVersionCmd wires the `version` command into the root CLI.
func newVersionCmd() *cobra.Command {
    var outputJSON bool

    cmd := &cobra.Command{
        Use:   "version",
        Short: "Print FlareGo version information",
        RunE: func(cmd *cobra.Command, args []string) error {
            if outputJSON {
                ver, commit, date := version.Components()
                payload := map[string]string{
                    "version": ver,
                    "commit":  commit,
                    "date":    date,
                }
                enc := json.NewEncoder(os.Stdout)
                enc.SetIndent("", "  ")
                return enc.Encode(payload)
            }

            fmt.Println(version.String())
            return nil
        },
    }

    cmd.Flags().BoolVar(&outputJSON, "json", false, "Print version information as JSON")
    return cmd
}
