package serverchat

import (
	"fmt"
	"time"
)

type Outgoing struct {
	message string
}

const maxQueue = 500

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

			start := time.Now()

			if _, err := c.conn.Write([]byte(o.message)); err != nil {
				elapsed := time.Since(start)

				if elapsed > 20*time.Millisecond {
					fmt.Printf(
						"Entró a error: [WRITE LENTA] cliente=%s duración=%s cola=%d/%d\n",
						c.address,
						elapsed,
						len(c.in),
						cap(c.in),
					)
				}
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

			elapsed := time.Since(start)

			if elapsed > 20*time.Millisecond {
				fmt.Printf(
					"[WRITE LENTA] cliente=%s duración=%s cola=%d/%d\n",
					c.address,
					elapsed,
					len(c.in),
					cap(c.in),
				)
			}

		}
	}
}
