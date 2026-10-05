package serverchat

import (
	"fmt"
	"net"
)

type Outgoing struct {
	message string
}

type Client struct {
	conn    net.Conn
	out     chan *Outgoing
	address string
	done    chan struct{}
}

func worker(client *Client, events chan<- Event) {

	for {
		select {
		case <-client.done:
			fmt.Printf("Worker para %s finalizando porque 'done' fue cerrado\n", client.address)
			return
		case o, ok := <-client.out:

			if !ok {
				fmt.Printf("Worker para %s finalizando porque 'out' fue cerrado\n", client.address)
				return
			}

			if _, err := client.conn.Write([]byte(o.message)); err != nil {
				fmt.Println("Error enviando stream message to client: ", err)
				events <- Event{kind: disconnection, address: client.address}
				return
			}

		}
	}
}
