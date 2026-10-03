package main

import (
	"sync"

	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/files"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(2)

	go func() {
		defer wg.Done()
		chat.StartServer()
	}()
	go func() {
		defer wg.Done()
		files.StartServer()
	}()

	wg.Wait()
}
