package chat

import "testing"

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
