package protocolfile

import (
	"bufio"
	"strings"
	"testing"
)

func TestReadHeadersRejectsOversizedLine(t *testing.T) {
	input := "key:" + strings.Repeat("a", MaxHeaderLineSize) + "\n"

	if _, err := ReadHeaders(bufio.NewReader(strings.NewReader(input))); err == nil {
		t.Fatal("ReadHeaders accepted an oversized header line")
	}
}

func TestReadHeadersAcceptsLineAtLimit(t *testing.T) {
	valueSize := MaxHeaderLineSize - len("key:\n")
	input := "key:" + strings.Repeat("a", valueSize) + "\n\n"

	if _, err := ReadHeaders(bufio.NewReader(strings.NewReader(input))); err != nil {
		t.Fatalf("ReadHeaders rejected a header line at the limit: %v", err)
	}
}
