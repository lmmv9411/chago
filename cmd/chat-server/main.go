package main

import (
	"sync"

	"github.com/lmmv9411/chago/internal/serverchat"
	"github.com/lmmv9411/chago/internal/serverfiles"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		serverchat.StartServer()
	}()
	go func() {
		defer wg.Done()
		serverfiles.StartServer()
	}()

	wg.Wait()
}
