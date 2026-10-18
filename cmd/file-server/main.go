package main

import (
	"fmt"
	"os"

	"github.com/lmmv9411/chago/internal/serverfiles"
)

func main() {
	if err := serverfiles.StartServer(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
