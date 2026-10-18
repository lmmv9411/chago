package serverfiles

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/lmmv9411/chago/internal/protocolfile"
)

func TestValidFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{name: "simple", filename: "document.pdf", want: true},
		{name: "empty"},
		{name: "dot", filename: "."},
		{name: "parent", filename: ".."},
		{name: "unix traversal", filename: "../secret"},
		{name: "windows traversal", filename: `..\secret`},
		{name: "colon", filename: "C:secret"},
		{name: "line break", filename: "file\nstatus:0"},
		{name: "terminal escape", filename: "file\x1b[31m"},
		{name: "null", filename: "file\x00name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validFilename(tt.filename); got != tt.want {
				t.Fatalf("validFilename(%q) = %t, want %t", tt.filename, got, tt.want)
			}
		})
	}
}

func TestInterruptedUploadDoesNotReplaceExistingFile(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	uploadDir := filepath.Join(dir, "uploads")
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(uploadDir, "document.txt")
	if err := os.WriteFile(target, []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}

	serverConn, clientConn := net.Pipe()
	server := &Connection{
		Conn:   serverConn,
		Reader: bufio.NewReader(serverConn),
		Writer: bufio.NewWriter(serverConn),
	}
	done := make(chan struct{})
	go func() {
		handleConnection(server)
		close(done)
	}()

	headers := protocolfile.BuildHeader(map[string]string{
		"filename": "document.txt",
		"size":     "10",
		"method":   "upload",
	})
	if _, err := clientConn.Write([]byte(headers)); err != nil {
		t.Fatal(err)
	}
	if _, err := protocolfile.ReadHeaders(bufio.NewReader(clientConn)); err != nil {
		t.Fatalf("read upload acknowledgment: %v", err)
	}
	if _, err := clientConn.Write([]byte("short")); err != nil {
		t.Fatal(err)
	}
	if err := clientConn.Close(); err != nil {
		t.Fatal(err)
	}
	<-done

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "original" {
		t.Fatalf("interrupted upload replaced the existing file with %q", got)
	}
	temporaryFiles, err := filepath.Glob(filepath.Join(uploadDir, ".chago-upload-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("temporary upload files were not cleaned up: %v", temporaryFiles)
	}
}

func TestUploadPublishesFileAfterCompleteTransfer(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	serverConn, clientConn := net.Pipe()
	server := &Connection{
		Conn:   serverConn,
		Reader: bufio.NewReader(serverConn),
		Writer: bufio.NewWriter(serverConn),
	}
	done := make(chan struct{})
	go func() {
		handleConnection(server)
		close(done)
	}()

	headers := protocolfile.BuildHeader(map[string]string{
		"filename": "document.txt",
		"size":     "5",
		"method":   "upload",
	})
	if _, err := clientConn.Write([]byte(headers)); err != nil {
		t.Fatal(err)
	}
	reader := bufio.NewReader(clientConn)
	if _, err := protocolfile.ReadHeaders(reader); err != nil {
		t.Fatalf("read upload acknowledgment: %v", err)
	}
	if _, err := fmt.Fprint(clientConn, "hello"); err != nil {
		t.Fatal(err)
	}
	if _, err := protocolfile.ReadHeaders(reader); err != nil {
		t.Fatalf("read upload completion: %v", err)
	}
	if err := clientConn.Close(); err != nil {
		t.Fatal(err)
	}
	<-done

	got, err := os.ReadFile(filepath.Join(dir, "uploads", "document.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("uploaded file = %q, want %q", got, "hello")
	}
}
