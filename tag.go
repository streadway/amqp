// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

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
