package main

import (
	"fmt"
	"os"

	"github.com/rkqze10/mt-tui/pkg/ui"
)

func main() {
	if err := ui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		os.Exit(1)
	}
}
