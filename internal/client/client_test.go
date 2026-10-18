package client

import (
	"errors"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/lmmv9411/chago/internal/chat"
)

func TestReadDisconnectClosesClient(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.Close()
		}
	}()

	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split listener address: %v", err)
	}

	client, err := Connect("127.0.0.1", port, "test")
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	select {
	case event := <-client.Events():
		if event.Kind != chat.Error {
			t.Fatalf("event kind = %v, want chat.Error", event.Kind)
		}
		if !strings.Contains(event.Message, "el servidor cerró la conexión") {
			t.Fatalf("disconnect reason = %q, want explanation that server closed connection", event.Message)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for disconnect error event")
	}

	select {
	case <-client.Done():
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for client shutdown")
	}

	if _, err := client.conn.conn.Read(make([]byte, 1)); !errors.Is(err, net.ErrClosed) {
		t.Fatalf("connection read error = %v, want net.ErrClosed", err)
	}
}

func TestConnectRejectsInvalidUsernamesBeforeDial(t *testing.T) {
	tests := []struct {
		name string
		user string
	}{
		{name: "empty"},
		{name: "too long", user: strings.Repeat("a", 65)},
		{name: "newline", user: "name\ninjected"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := Connect("", "", tt.user); err == nil {
				t.Fatal("Connect accepted an invalid username")
			}
		})
	}
}

func TestTransferIDsAreUniqueAcrossDirections(t *testing.T) {
	c := &Client{}

	first := c.newTransferID("upload")
	second := c.newTransferID("download")
	third := c.newTransferID("upload")

	if first == second || first == third || second == third {
		t.Fatalf("transfer IDs must be unique, got %q, %q, %q", first, second, third)
	}

	numGoroutines := 100
	idsPorGoroutine := 50
	totalEsperado := numGoroutines * idsPorGoroutine

	idsCha := make(chan string, totalEsperado)

	var wg sync.WaitGroup

	for range numGoroutines {
		wg.Go(func() {
			for range idsPorGoroutine {
				idsCha <- c.newTransferID("a")
			}
		})
	}

	wg.Wait()
	close(idsCha)

	unicos := make(map[string]bool, totalEsperado)

	for i := range idsCha {
		if unicos[i] {
			t.Fatalf("¡ID duplicado detectado! El ID %s ya había sido generado.", i)
		}

		unicos[i] = true
	}

	if len(unicos) != totalEsperado {
		t.Errorf("Se esperaban %d IDs únicos, pero se obtuvieron %d", totalEsperado, len(unicos))
	}

}
