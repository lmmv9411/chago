package protocolfile

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
)

const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
)

const (
	MaxHeaderLineSize = KiB
	MaxHeaderCount    = 16
	MaxBodySize       = GiB
)

func ReadHeaders(reader *bufio.Reader) (map[string]string, error) {

	headers := make(map[string]string)
	count := 0

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			return nil, errors.New("Error en lectura de header. " + err.Error())
		}

		if len(line) > MaxHeaderLineSize {
			return nil, errors.New("linea header supera tamaño permitido.")
		}

		count++

		if line == "\n" {
			break
		}

		if count > MaxHeaderCount {
			return nil, errors.New("lineas de header superan el maximo permitido.")
		}

		split := strings.SplitN(line, ":", 2)

		if len(split) != 2 {
			return nil, errors.New("Formato incorrecto header key:value")
		}

		key := strings.TrimSpace(split[0])
		value := strings.TrimSpace(split[1])

		if key == "" || value == "" {
			return nil, errors.New("Par key:value vacios en linea header")
		}

		headers[key] = value

	}

	return headers, nil
}

func BuildHeader(headers map[string]string) string {
	var header strings.Builder

	for key, value := range headers {
		fmt.Fprintf(&header, "%s:%s\n", key, value)
	}

	header.WriteString("\n")

	return header.String()
}
