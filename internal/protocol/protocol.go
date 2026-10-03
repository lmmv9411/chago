package protocol

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
)

const (
	MaxHeaderLineSize = KiB
	MaxHeaderCount    = 16
	MaxSenderSize     = 64
	MaxBodySize       = KiB
)

func ReadHeaders(r *bufio.Reader) (map[string]string, *strings.Builder, error) {

	headers := make(map[string]string)
	count := 0

	var header strings.Builder

	for {
		line, err := r.ReadString('\n')

		if err != nil {
			return nil, nil, err
		}

		if len(line) > MaxHeaderLineSize {
			return nil, nil, errors.New("Máximo tamaño de linea header alcanzado.")
		}

		header.WriteString(line)

		if line == "\n" {
			break
		}

		data := strings.SplitN(line, ":", 2)

		if len(data) != 2 {
			return nil, nil, errors.New("Protocolo mal formado, key:value")
		}

		count++

		if count > MaxHeaderCount {
			return nil, nil, errors.New("Máximo headers alcanzado")
		}

		key := strings.TrimSpace(data[0])
		value := strings.TrimSpace(data[1])

		if key == "" || value == "" {
			return nil, nil, errors.New("header con key y/o value vacíos.")
		}

		if _, exists := headers[key]; exists {
			return nil, nil, errors.New("header duplicado: " + key)
		}

		headers[key] = value
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
