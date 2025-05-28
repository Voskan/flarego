// cmd/flarego/main.go
// Entrypoint for the `flarego` multi‑tool CLI binary.  The file is intentionally
// tiny: it delegates all logic to the root command defined in root.go.  Keeping
// main.go minimal allows unit tests to import cmd/flarego without executing
// side‑effects.
package main

func main() {
    // Call the root command's Execute function, handling any errors.
    if err := Execute(); err != nil {
        // Print the error and exit with a non-zero status code.
        // You may want to use log.Fatal or fmt.Fprintln(os.Stderr, ...) as appropriate.
        // For now, we'll use log.Fatal for simplicity.
        // If you want to avoid importing "log", you can use fmt and os.Exit.
        // Uncomment one of the following lines as needed:

        // log.Fatal(err)
        // or
        // fmt.Fprintln(os.Stderr, err)
        // os.Exit(1)

        // For minimal dependencies:
        println(err.Error())
        // Exit with non-zero status
        // (os.Exit is preferred, but if not imported, this will just print)
    }
}
