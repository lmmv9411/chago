package server

import (
	"fmt"
	"io"
	"net"
	"os"
)

type Outgoing struct {
	message string
	file    FileEvent
	isFile  bool
}

type Client struct {
	conn    net.Conn
	out     chan Outgoing
	address string
	done    chan struct{}
}

func worker(client *Client, events chan<- Event) {

	defer close(client.out)

	for {

		select {

		case <-client.done:
			fmt.Printf("Worker para %s finalizando por señal de 'done'\n", client.address)
			return
		case o, ok := <-client.out:

			if !ok {
				fmt.Printf("Worker para %s finalizando porque 'out' fue cerrado\n", client.address)
				return
			}

			if !o.isFile {
				if _, err := client.conn.Write([]byte(o.message)); err != nil {
					fmt.Println("Error enviando stream message to client: ", err)
					events <- Event{kind: disconnection, address: client.address}
					break
				}
			} else {
				file, err := os.Open(o.file.path)
				if err != nil {
					fmt.Println("Error abriendo fichero", err)
					continue
				}

				_, err = client.conn.Write([]byte(o.file.headers))
				if err != nil {
					fmt.Println("Error enviando stream header to client: ", err)
					events <- Event{kind: disconnection, address: client.address}
					file.Close()
					break
				}

				_, err = io.CopyN(client.conn, file, o.file.size)
				if err != nil {
					fmt.Println("Error enviando stream body to client: ", err)
					file.Close()
					events <- Event{kind: disconnection, address: client.address}
					break
				}
				file.Close()
			}
		}
	}
}
