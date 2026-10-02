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
	conn net.Conn
	out  chan Outgoing
}

func worker(client Client, events chan<- Event) {

	for o := range client.out {
		if !o.isFile {
			if _, err := client.conn.Write([]byte(o.message)); err != nil {
				fmt.Println("Error enviando stream message to client: ", err)
				events <- Event{kind: disconnection}
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
				events <- Event{kind: disconnection}
				break
			}

			_, err = io.CopyN(client.conn, file, o.file.size)
			if err != nil {
				fmt.Println("Error enviando stream body to client: ", err)
				events <- Event{kind: disconnection}
				break
			}
			file.Close()
		}
	}
}
