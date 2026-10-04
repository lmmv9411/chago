package protocolfile

import (
	"bufio"
	"errors"
	"net"
	"strings"
)

func ReadHeaders(conn net.Conn) (map[string]string, error) {

	reader := bufio.NewReader(conn)
	headers := make(map[string]string)

	for {
		line, err := reader.ReadString('\n')

		if err != nil {
			return nil, errors.New("Error en lectura de header. " + err.Error())
		}

		if line == "\n" {
			break
		}

		split := strings.SplitN(line, ":", 2)

		if len(split) != 2 {
			return nil, errors.New("Formato incorrecto header key:value")
		}

		key := split[0]
		value := split[1]

		if key == "" || value == "" {
			return nil, errors.New("Par key:value vacios en linea header")
		}

		headers[key] = value

	}

	return headers, nil
}
