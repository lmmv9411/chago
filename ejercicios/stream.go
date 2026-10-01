package ejercicios

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	"runtime/debug"
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
	barWith := 30

	filled := int((percentage / 100.0) * float64(barWith))
	currentStr := formatBytes(p.written)
	totalStr := formatBytes(p.total)

	fmt.Printf("\r[")

	for i := range barWith {
		if i < filled {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}

	fmt.Printf("] %.0f%% | %s / %s", percentage, currentStr, totalStr)

	return n, err
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	return fmt.Sprintf("%.2f %s", float64(b)/float64(div), units[exp])
}

func (fs *FileServer) start() {
	listener, err := net.Listen("tcp", ":3000")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Server start on port :3000.")

	for {

		fmt.Println("waiting connection...")
		conn, err := listener.Accept()
		fmt.Println("connected to client!")

		if err != nil {
			log.Fatal(err)
		}
		go fs.readLoop(conn)
	}

}

func (f *FileServer) readLoop(conn net.Conn) {

	for {
		//var buff bytes.Buffer
		var size int64

		err := binary.Read(conn, binary.LittleEndian, &size)

		progress := &ProgressWriter{
			total:  size,
			writer: io.Discard,
		}

		if err != nil {
			if err == io.EOF {
				fmt.Printf("end connection: %s\n", err)
			} else {
				fmt.Printf("error reading: %s\n", err)
			}
			break
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
		err := sendFile(1024 * 1024 * 500)
		if err != nil {
			log.Fatal(err)
		}
		runtime.GC()
		debug.FreeOSMemory()
	}()

	server := FileServer{}
	server.start()

}
