package amqp

import (
	"strings"
	"testing"
)

func TestGoFuzzCrashers(t *testing.T) {
	testData := []string{
		"\b000000",
		"\x02\x16\x10�[��\t\xbdui�" + "\x10\x01\x00\xff\xbf\xef\xbfｻn\x99\x00\x10r",
		"\x0300\x00\x00\x00\x040000",
		"\x020000000000000000000" + "0\x00\x00\x000!00000000000000" + "0000000000000000000x" + "\x800000000000000000000" + "00000000000000000000" + "00000000000000000000" + "00",
	}

	for idx, testStr := range testData {
		r := reader{strings.NewReader(testStr)}
		frame, err := r.ReadFrame()
		if err != nil && frame != nil {
			t.Errorf("%d. frame is not nil: %#v err = %v", idx, frame, err)
		}
	}
}
