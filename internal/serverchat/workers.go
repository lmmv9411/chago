package serverchat

import (
	"fmt"
)

type Outgoing struct {
	message string
}

const maxQueue = 100

type Client struct {
	conn    *Connection
	in      chan *Outgoing
	address string
	done    chan struct{}
}

func worker(c *Client, e chan<- Event) {

	for {
		select {
		case <-c.done:
			fmt.Printf("Worker para %s finalizando porque 'done' fue cerrado\n", c.address)
			return
		case o, ok := <-c.in:

			if !ok {
				fmt.Printf("Worker para %s finalizando porque 'out' fue cerrado\n", c.address)
				return
			}

			if _, err := c.conn.Write([]byte(o.message)); err != nil {
				select {
				case <-c.done:
					//la desconexión ya estaba encurso
					return
				default:
				}
				fmt.Println("Error enviando stream message to client: ", err)
				e <- Event{kind: disconnection, address: c.address}
				return
			}

		}
	}
}
