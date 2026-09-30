package ejercicios

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

type FileServer struct{}
type ProgressWriter struct {
	writer  io.Writer
	total   int64
	written int64
}

func (p *ProgressWriter) Write(data []byte) (int, error) {
	n, err := p.writer.Write(data)

	p.written += int64(n)

	percentage := float64(p.written) / float64(p.total) * 100

	filled := int(percentage)

	fmt.Printf("\r[")

	for i := range 100 {
		if i < filled {
			fmt.Printf("█")
		} else {
			fmt.Printf("░")
		}
	}

	fmt.Printf("] %.0f%%", percentage)

	return n, err
}

func (fs *FileServer) start() {
	listener, err := net.Listen("tcp", ":3000")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server start on port :3000.")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go fs.readLoop(conn)
	}

}

func (fs *FileServer) readLoop(conn net.Conn) {

	for {
		var buff bytes.Buffer
		var size int64
		err := binary.Read(conn, binary.LittleEndian, &size)

		progress := &ProgressWriter{
			total:  size,
			writer: &buff,
		}

		if err != nil {
			if err == io.EOF {
				fmt.Printf("end connection: %s\n", err)
			} else {
				fmt.Printf("error reading: %s\n", err)
			}
			return
		}

		n, err := io.CopyN(progress, conn, size)
		fmt.Println("")

		if err != nil {
			if err == io.EOF {
				fmt.Printf("end of reading: %s\n", err)
				break
			}
			log.Fatal(err)
		}

		//fmt.Println(buff.Bytes())
		fmt.Printf("received %d bytes over the network\n", n)

	}

}

func sendFile(size int) error {
	file := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, file)

	if err != nil {
		return err
	}

	conn, err := net.Dial("tcp", ":3000")

	if err != nil {
		return err
	}

	defer conn.Close()

	err = binary.Write(conn, binary.LittleEndian, int64(size))

	if err != nil {
		return err
	}

	_, err = io.CopyN(conn, bytes.NewReader(file), int64(size))

	if err != nil {
		return err
	}

	//fmt.Printf("written %d bytes over the network\n", n)
	return nil
}

func Stream() {

	go func() {
		time.Sleep(2 * time.Second)
		err := sendFile(1024 * 1024 * 500) //500MiB
		if err != nil {
			log.Fatal(err)
		}
	}()

	server := FileServer{}
	server.start()

}
