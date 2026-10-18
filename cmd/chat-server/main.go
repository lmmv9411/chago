package main

import (
	"fmt"
	"os"

	"github.com/lmmv9411/chago/internal/serverchat"
	"github.com/lmmv9411/chago/internal/serverfiles"
)

func main() {
	errs := make(chan error, 2)

	go func() {
		errs <- serverchat.StartServer()
	}()
	go func() {
		errs <- serverfiles.StartServer()
	}()

	if err := <-errs; err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
