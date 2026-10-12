package main

import (
	"fmt"
	"os"

	"github.com/lmmv9411/chago/internal/client"
	ui "github.com/lmmv9411/chago/internal/ui/terminal"
	"github.com/lmmv9411/chago/internal/ui/terminal/chat"
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() error {
	login, err := ui.LoginMain()

	if err != nil {
		return err
	}

	c, err := client.Connect(login.Address, "8080", login.User)

	if err != nil {
		return err
	}

	defer c.Close()

	return chat.Run(c)
}
