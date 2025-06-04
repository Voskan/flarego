// cmd/flarego/ebpf_attach.go
// Implements the `flarego ebpf-attach` command.  It attaches to a running Go
// process using eBPF uprobes.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newEBPFAttachCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "ebpf-attach <pid>",
        Short: "Attach to a running Go process using eBPF uprobes (Linux only)",
        RunE: func(cmd *cobra.Command, args []string) error {
            return fmt.Errorf("eBPF attach not yet implemented")
        },
    }
}