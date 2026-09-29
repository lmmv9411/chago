package server

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

func ReadHeaders(r *bufio.Reader) (map[string]string, *strings.Builder, error) {

	headers := make(map[string]string)

	var header strings.Builder

	for {
		line, err := r.ReadString('\n')
		header.WriteString(line)

		if err != nil {
			return nil, nil, err
		}

		if line == "\n" {
			break
		}

		data := strings.SplitN(line, ":", 2)

		if len(data) != 2 {
			return nil, nil, errors.New("Error en protocolo de cabecera, no contiene par llave - valor.")
		}

		headers[strings.TrimSpace(data[0])] = strings.TrimSpace(data[1])
	}

	return headers, &header, nil
}

func BuildHeader(headers map[string]string) string {
	var header strings.Builder

	for key, value := range headers {
		fmt.Fprintf(&header, "%s:%s\n", key, value)
	}

	header.WriteString("\n")

	return header.String()

}
