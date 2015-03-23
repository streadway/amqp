package amqp

import "testing"

func TestConfirmOneResequences(t *testing.T) {
	var (
		fixtures = []Confirmation{
			{1, true},
			{2, false},
			{3, true},
		}
		c = newConfirms()
		l = make(chan Confirmation, len(fixtures))
	)

	c.listen(l)

	for i, _ := range fixtures {
		if want, got := uint64(i+1), c.publish(); want != got {
			t.Fatalf("expected publish to return the 1 based delivery tag published, want: %d, got: %d", want, got)
		}
	}

	c.one(fixtures[1])
	c.one(fixtures[2])

	select {
	case confirm := <-l:
		t.Fatalf("expected to wait in order to properly resequence results, got: %+v", confirm)
	default:
	}

	c.one(fixtures[0])

	for i, fix := range fixtures {
		if want, got := fix, <-l; want != got {
			t.Fatalf("expected to return confirmations in sequence for %d, want: %+v, got: %+v", i, want, got)
		}
	}
}

func TestConfirmMultipleResequences(t *testing.T) {
	var (
		fixtures = []Confirmation{
			{1, true},
			{2, true},
			{3, true},
			{4, true},
		}
		c = newConfirms()
		l = make(chan Confirmation, len(fixtures))
	)
	c.listen(l)

	for _, _ = range fixtures {
		c.publish()
	}

	c.multiple(fixtures[len(fixtures)-1])

	for i, fix := range fixtures {
		if want, got := fix, <-l; want != got {
			t.Fatalf("expected to confirm multiple in sequence for %d, want: %+v, got: %+v", i, want, got)
		}
	}
}

func BenchmarkSequentialBufferedConfirms(t *testing.B) {
	var (
		c = newConfirms()
		l = make(chan Confirmation, 10)
	)

	c.listen(l)

	for i := 0; i < t.N; i++ {
		if i > cap(l)-1 {
			<-l
		}
		c.one(Confirmation{c.publish(), true})
	}
}
