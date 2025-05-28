// cmd/flarego/main.go
// Entrypoint for the `flarego` multi‑tool CLI binary.  The file is intentionally
// tiny: it delegates all logic to the root command defined in root.go.  Keeping
// main.go minimal allows unit tests to import cmd/flarego without executing
// side‑effects.
package main

func main() {
    // Call the root command's Execute function, handling any errors.
    if err := main.Execute(); err != nil {
        println(err.Error())
    }
}
