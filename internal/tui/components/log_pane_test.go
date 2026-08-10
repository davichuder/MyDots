package components

import (
	"fmt"
	"testing"
)

func TestLogPaneAppend(t *testing.T) {
	tests := []struct {
		name   string
		append func(*LogPane)
		want   []string
	}{
		{
			name: "ignores empty lines",
			append: func(pane *LogPane) {
				pane.Append("first")
				pane.Append("")
				pane.Append("second")
			},
			want: []string{"first", "second"},
		},
		{
			name: "evicts the oldest line when capacity is exceeded",
			append: func(pane *LogPane) {
				for line := range LogPaneMaxLines + 1 {
					pane.Append(fmt.Sprintf("line-%03d", line))
				}
			},
			want: func() []string {
				lines := make([]string, LogPaneMaxLines)
				for index := range lines {
					lines[index] = fmt.Sprintf("line-%03d", index+1)
				}
				return lines
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pane := NewLogPane()

			tt.append(&pane)

			if len(pane.Lines) != len(tt.want) {
				t.Fatalf("line count = %d, want %d", len(pane.Lines), len(tt.want))
			}
			for index, want := range tt.want {
				if got := pane.Lines[index]; got != want {
					t.Errorf("line %d = %q, want %q", index, got, want)
				}
			}
		})
	}
}
