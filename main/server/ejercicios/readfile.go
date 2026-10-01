package ejercicios

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func StartRead() {
	archivo, err := os.Open("./texto.txt")

	if err != nil {
		log.Fatalf("Error al abrir archivo: %v", err)
	}

	defer archivo.Close()

	ReadStream(archivo)
}

func ReadStream(reader io.Reader) {

	r := bufio.NewReader(reader)

	headers := make(map[string]string)

	for {
		line, err := r.ReadString('\n')

		if err != nil {
			log.Fatalln(err)
			break
		}

		if line == "\n" {
			break
		}

		data := strings.SplitN(line, ":", 2)
		headers[strings.TrimSpace(data[0])] = strings.TrimSpace(data[1])

	}

	fmt.Printf("data: %#v\n\n", headers)

	n, err := strconv.Atoi(headers["size"])

	if err != nil {
		log.Fatal(err)
	}

	kind := headers["content-type"]

	switch kind {
	case "text/plain":
		_, err := io.CopyN(os.Stdout, r, int64(n))
		fmt.Println()
		if err != nil {
			log.Fatal(err)
		}
	case "file":
		fmt.Println("Por implementar...")
	}

}
