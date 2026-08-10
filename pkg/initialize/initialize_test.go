package initialize_test

import (
	"io"
	"time"
)

// Reference: https://www.alanwood.net/demos/ansi.html
const (
	// defaultSubmit is appended to all responses to move to the next one. These represent \r\n.
	defaultSubmit = "\x0D\x0A"

	// selectSubmit is a special case where the defaultSubmit messes up the input in select statements.
	selectSubmit = "\x0D"

	// selectOption is used in a select and multiselect to mark or unmark an item.
	_ = "\x20"

	// arrowDown is used in a select and multiselect to move downwards.
	arrowDown = "\x1b[B"

	// arrowRight is used in a confirm to move between yes and no.
	_ = "\x1b[C"
)

// PacedReader yields input one byte at a time with a small delay between reads.
//
// An unpaced strings.Reader races huh's async field-switch
// and leaks characters between fields.
//
// By feeding bytes one at a time with a delay avoids it.
type PacedReader struct {
	data  []byte
	delay time.Duration
}

var _ io.Reader = (*PacedReader)(nil) // ensure interface is implemented

// NewPacedReader creates a PacedReader over s, sleeping delay before each byte read.
func NewPacedReader(s string, delay time.Duration) *PacedReader {
	return &PacedReader{data: []byte(s), delay: delay}
}

// Read implements io.Reader.
func (r *PacedReader) Read(p []byte) (int, error) {
	if len(r.data) == 0 {
		return 0, io.EOF
	}
	time.Sleep(r.delay)
	n := copy(p, r.data[:1])
	r.data = r.data[1:]
	return n, nil
}
