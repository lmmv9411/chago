package client

import (
	"strings"
	"testing"
)

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
}
