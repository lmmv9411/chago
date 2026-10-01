package server

import (
	"fmt"
	"io"
	"net"
)

type Outgoing struct {
	header  string
	message string
	reader  io.Reader
	size    int64
	isFile  bool
}

type Client struct {
	conn net.Conn
	out  chan Outgoing
}

func Worker(client Client) {

	for o := range client.out {
		if !o.isFile {
			if _, err := client.conn.Write([]byte(o.message)); err != nil {
				fmt.Println("Error enviando stream message to client: ", err)
				break
			}
		} else {
			_, err := client.conn.Write([]byte(o.header))
			if err != nil {
				fmt.Println("Error enviando stream header to client: ", err)
				break
			}
			_, err = io.CopyN(client.conn, o.reader, o.size)
			if err != nil {
				fmt.Println("Error enviando stream body to client: ", err)
				break
			}
		}
	}
}
