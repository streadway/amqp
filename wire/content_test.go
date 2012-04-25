package wire

import (
	"testing"
)

var (
	empty = []byte{0, 10, 0, 0, 0, 0, 0, 0, 0, 0, 0, 20, 0, 0}
)

func TestEmptyContentHeader(t *testing.T) {
	m := newBuffer(empty).NextContentHeader()

	if m.Class != 10 {
		t.Error("bad class")
	}

	if m.Size != 20 {
		t.Error("bad size")
	}
}
