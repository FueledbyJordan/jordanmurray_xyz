package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
)

func main() {
	formatter := html.New(html.WithClasses(true))
	style := styles.Get("monokai")
	if style == nil {
		fmt.Fprintf(os.Stderr, "Style not found\n")
		os.Exit(1)
	}
	formatter.WriteCSS(os.Stdout, style)
}
