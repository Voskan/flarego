// cmd/flarego/kubectl.go
// Implements the `flarego kubectl` command.  It port‑forwards a Kubernetes
// Pod and attaches to it using the `flarego attach` command.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newKubectlCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "kubectl attach -n <namespace> <resource>",
        Short: "Port-forward and attach to a Kubernetes Pod",
        RunE: func(cmd *cobra.Command, args []string) error {
            return fmt.Errorf("kubectl attach not yet implemented")
        },
    }
}