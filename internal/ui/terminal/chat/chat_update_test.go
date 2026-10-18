package chat

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestClientDisconnectWaitsForAcknowledgement(t *testing.T) {
	model := chatmodel{}

	updated, cmd := model.Update(ClientDisconnectedMsg{})
	if cmd != nil {
		t.Fatal("disconnect quit immediately instead of allowing the reason to be read")
	}
	model = updated.(chatmodel)
	if !model.disconnected {
		t.Fatal("chat model did not enter disconnected state")
	}

	_, cmd = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("pressing a key after disconnect did not return a quit command")
	}
	if msg := cmd(); msg != (tea.QuitMsg{}) {
		t.Fatalf("disconnect command returned %T, want tea.QuitMsg", msg)
	}
}

func TestParseFileCommand(t *testing.T) {
	tests := []struct {
		name      string
		message   string
		wantPath  string
		wantMatch bool
	}{
		{name: "path with spaces", message: "/file /tmp/my file.txt", wantPath: "/tmp/my file.txt", wantMatch: true},
		{name: "missing path", message: "/file", wantMatch: true},
		{name: "command prefix only", message: "/filex document.txt"},
		{name: "ordinary message", message: "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotMatch := parseFileCommand(tt.message)
			if gotPath != tt.wantPath || gotMatch != tt.wantMatch {
				t.Fatalf("parseFileCommand(%q) = (%q, %t), want (%q, %t)", tt.message, gotPath, gotMatch, tt.wantPath, tt.wantMatch)
			}
		})
	}
}
