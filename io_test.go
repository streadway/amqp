package amqp_test

import (
	"io"
	"testing"
)

// Combines a reader and writer into a pipe
type pipe struct {
	io.Reader
	io.WriteCloser
}

// Returns a pipe pair that can be treated as a writer to the client and reader from the client
// and a client that is a writer to the server and reader from the server
func interPipes(t *testing.T) (server io.ReadWriteCloser, client io.ReadWriteCloser) {
	r1, w1 := io.Pipe()
	r2, w2 := io.Pipe()

	return &logIO{t, "server", pipe{r1, w2}}, &logIO{t, "client", pipe{r2, w1}}
}

type logIO struct {
	t      *testing.T
	prefix string
	proxy  io.ReadWriteCloser
}

func (me *logIO) Read(p []byte) (n int, err error) {
	me.t.Logf("%s reading %d\n", me.prefix, len(p))
	n, err = me.proxy.Read(p)
	if err != nil {
		me.t.Logf("%s read %x: %v\n", me.prefix, p[0:n], err)
	} else {
		me.t.Logf("%s read %x\n", me.prefix, p[0:n])
	}
	return
}

func (me *logIO) Write(p []byte) (n int, err error) {
	n, err = me.proxy.Write(p)
	if err != nil {
		me.t.Logf("%s write %d, %x: %v\n", me.prefix, len(p), p[0:n], err)
	} else {
		me.t.Logf("%s write %d, %x\n", me.prefix, len(p), p[0:n])
	}
	return
}

func (me *logIO) Close() (err error) {
	err = me.proxy.Close()
	if err != nil {
		me.t.Logf("%s close : %v\n", me.prefix, err)
	} else {
		me.t.Logf("%s close\n", me.prefix, err)
	}
	return
}

func (me *logIO) Test() {
	me.t.Logf("test: %v\n", me)
}
