package main

import (
	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/files"
)

func main() {
	chat.StartServer()
	files.StartServer()
}
