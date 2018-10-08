package packagetest

import (
	"fmt"
	"net/http"
)

// PrintHello prints hello
func PrintHello(w http.ResponseWriter) {
	fmt.Fprintln(w, "Hello World")
}
