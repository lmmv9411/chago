package serverchat

import (
	"bufio"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/protocolfile"
)

func TestValidateFileNotification(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]string
		wantErr bool
	}{
		{
			name: "valid",
			headers: map[string]string{
				"filename": "document.pdf",
				"size":     "128",
			},
		},
		{name: "missing filename", headers: map[string]string{"size": "128"}, wantErr: true},
		{name: "path traversal", headers: map[string]string{"filename": "../secret", "size": "128"}, wantErr: true},
		{name: "terminal escape", headers: map[string]string{"filename": "file\x1b[31m", "size": "128"}, wantErr: true},
		{name: "invalid size", headers: map[string]string{"filename": "document.pdf", "size": "invalid"}, wantErr: true},
		{name: "negative size", headers: map[string]string{"filename": "document.pdf", "size": "-1"}, wantErr: true},
		{
			name: "oversized",
			headers: map[string]string{
				"filename": "document.pdf",
				"size":     strconv.FormatInt(protocolfile.MaxBodySize+1, 10),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFileNotification(tt.headers)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateFileNotification() error = %v, wantErr %t", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizeMessageRemovesTerminalControls(t *testing.T) {
	input := "hello\x1b[31mworld\r\nnext"
	want := "hello [31mworld \nnext"

	if got := sanitizeMessage(input); got != want {
		t.Fatalf("sanitizeMessage(%q) = %q, want %q", input, got, want)
	}
}

func TestHandleConnectionKeepsSanitizedMessageFramingValid(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	conn := &Connection{
		conn:   serverConn,
		Reader: bufio.NewReader(serverConn),
		Writer: bufio.NewWriter(serverConn),
	}
	events := make(chan Event, 4)
	done := make(chan struct{})
	go func() {
		handleConnection(conn, events)
		close(done)
	}()

	body := "hello\x1b[31m"
	headers := protocolchat.BuildHeader(map[string]string{
		"content-type": "text/plain",
		"sender":       "alice",
		"size":         strconv.Itoa(len(body)),
	})
	if _, err := io.WriteString(clientConn, headers+body); err != nil {
		t.Fatal(err)
	}

	message := <-events
	relayedReader := bufio.NewReader(strings.NewReader(message.message))
	receivedHeaders, _, err := protocolchat.ReadHeaders(relayedReader)
	if err != nil {
		t.Fatalf("read relayed message headers: %v", err)
	}
	size, err := strconv.Atoi(receivedHeaders["size"])
	if err != nil {
		t.Fatalf("parse relayed body size: %v", err)
	}
	relayedBody, err := io.ReadAll(relayedReader)
	if err != nil {
		t.Fatalf("read relayed message body: %v", err)
	}
	if len(relayedBody) != size {
		t.Fatalf("relayed header size = %d, body has %d bytes", size, len(relayedBody))
	}
	if string(relayedBody) != "hello [31m" {
		t.Fatalf("relayed body = %q, want %q", relayedBody, "hello [31m")
	}
	if _, _, err := protocolchat.ReadHeaders(bufio.NewReader(clientConn)); err != nil {
		t.Fatalf("read message acknowledgment: %v", err)
	}

	if err := clientConn.Close(); err != nil {
		t.Fatal(err)
	}
	<-done
}
