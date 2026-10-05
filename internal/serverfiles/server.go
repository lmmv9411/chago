package serverfiles

import (
	"bufio"
	"net"
	"strconv"

	"github.com/lmmv9411/chago/internal/protocolfile"
)

type Status int

const (
	Ok Status = iota
	RequestError
	InternalError
)

type Server struct {
	Listener net.Listener
}

type Connection struct {
	Conn   net.Conn
	Writer *bufio.Writer
	Reader *bufio.Reader
}

func NewServer(network string, address string) (*Server, error) {

	listener, err := net.Listen(network, address)

	if err != nil {
		return nil, err
	}

	return &Server{Listener: listener}, nil
}

func (s *Server) CloseListener() error {
	if s.Listener != nil {
		return s.Listener.Close()
	}
	return nil
}

func (s *Connection) CloseConn() error {
	if s.Conn != nil {
		return s.Conn.Close()
	}
	return nil
}

func (s *Server) Accept() (*Connection, error) {

	conn, err := s.Listener.Accept()

	if err != nil {
		return nil, err
	}

	connection := &Connection{
		Conn:   conn,
		Writer: bufio.NewWriter(conn),
		Reader: bufio.NewReader(conn),
	}

	return connection, nil
}

func (s *Connection) SendError(msg string, code Status) error {
	return s.send(msg, code)
}

func (s *Connection) SendOk(msg string) error {
	return s.send(msg, Ok)
}

func (s *Connection) send(msg string, code Status) error {

	headers := make(map[string]string)

	headers["status"] = strconv.Itoa(int(code))
	headers["message"] = msg

	header := protocolfile.BuildHeader(headers)

	_, err := s.Writer.Write([]byte(header))

	if err != nil {
		return err
	}

	return s.Writer.Flush()
}
