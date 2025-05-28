//go:build cli
// +build cli

// cmd/flarego/main.go
// Entrypoint for the `flarego` multi‑tool CLI binary.  The file is intentionally
// tiny: it delegates all logic to the root command defined in root.go.  Keeping
// main.go minimal allows unit tests to import cmd/flarego without executing
// side‑effects.
package main

func main() {
    Execute()
}
