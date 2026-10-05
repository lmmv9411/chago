package client

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/lmmv9411/chago/internal/serverchat"
)

func response(headers map[string]string) (*string, error) {

	status, ok := headers["status"]

	if !ok {
		return nil, errors.New("respuesta del servidor sin status")
	}

	code, err := strconv.Atoi(status)

	if err != nil {
		return nil, fmt.Errorf("status inválido en respuesta del servidor: %v\n", err)
	}

	if serverchat.Status(code) != serverchat.Ok {
		return nil, fmt.Errorf("servidor respondió %s: %s", status, headers["message"])
	}

	msg := headers["message"]
	return &msg, nil
}
