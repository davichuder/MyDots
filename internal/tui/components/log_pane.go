package components

const LogPaneMaxLines = 200

// LogPane stores the bounded installation feedback shown by TUI screens.
type LogPane struct {
	Lines []string
}

func NewLogPane() LogPane {
	return LogPane{}
}

func (pane *LogPane) Append(line string) {
	if line == "" {
		return
	}

	pane.Lines = append(pane.Lines, line)
	if len(pane.Lines) > LogPaneMaxLines {
		pane.Lines = pane.Lines[len(pane.Lines)-LogPaneMaxLines:]
	}
}
