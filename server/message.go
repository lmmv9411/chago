package server

import "encoding/json"

type Kind int

const (
	TypeText Kind = iota
	TypeFile
)

type Message struct {
	Sender  string `json:"sender"`
	Content string `json:"content"`
	Type    Kind   `json:"kind"`
	Size    int64  `json:"size"`
}

func (m *Message) ToJson() ([]byte, error) {
	return json.Marshal(m)
}

func FromJson(data []byte) (*Message, error) {
	var m Message
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return &m, nil
}
