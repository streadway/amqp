package amqp

import (
	"crypto/rand"
	"encoding/ascii85"
	mrand "math/rand"
)

func randomTag() string {
	in := make([]byte, 32)
	out := make([]byte, ascii85.MaxEncodedLen(len(in)))

	if _, err := rand.Read(in); err != nil {
		for i, n := range mrand.Perm(256)[:len(in)] {
			in[i] = byte(n)
		}
	}

	ascii85.Encode(in, out)
	return string(out)
}
