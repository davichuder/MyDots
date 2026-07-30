package modules

import (
	"errors"
	"fmt"
	"io"
)

func writeDiagnostics(w io.Writer, messages ...string) error {
	var errs []error
	for _, message := range messages {
		if _, err := fmt.Fprintln(w, message); err != nil {
			errs = append(errs, fmt.Errorf("write diagnostic: %w", err))
		}
	}
	return errors.Join(errs...)
}
