package amqp

import "sync"

// Confirms resequences and notifies one or multiple publisher confirmation listeners
type Confirms struct {
	m         sync.Mutex
	listeners []chan Confirmation
	sequencer map[uint64]Confirmation
	published uint64
	expecting uint64
}

// newConfirms allocates a confirms
func newConfirms() *Confirms {
	return &Confirms{
		sequencer: map[uint64]Confirmation{},
		published: 0,
		expecting: 1,
	}
}

// Published returns sequential number of published messages
func (c *Confirms) Published() uint64 {
	c.m.Lock()
	defer c.m.Unlock()

	return c.published
}

// Listen is used to listen on incoming confirmations
// of publishes
func (c *Confirms) Listen(l chan Confirmation) {
	c.m.Lock()
	defer c.m.Unlock()

	c.listeners = append(c.listeners, l)
}

// Publish increments the publishing counter
func (c *Confirms) Publish() uint64 {
	c.m.Lock()
	defer c.m.Unlock()

	c.published++
	return c.published
}

// confirm confirms one publishing, increments the expecting delivery tag, and
// removes bookkeeping for that delivery tag.
func (c *Confirms) confirm(confirmation Confirmation) {
	delete(c.sequencer, c.expecting)
	c.expecting++
	for _, l := range c.listeners {
		l <- confirmation
	}
}

// resequence confirms any out of order delivered confirmations
func (c *Confirms) resequence() {
	for c.expecting <= c.published {
		sequenced, found := c.sequencer[c.expecting]
		if !found {
			return
		}
		c.confirm(sequenced)
	}
}

// One confirms one publishing and all following in the publishing sequence
func (c *Confirms) One(confirmed Confirmation) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.expecting == confirmed.DeliveryTag {
		c.confirm(confirmed)
	} else {
		c.sequencer[confirmed.DeliveryTag] = confirmed
	}
	c.resequence()
}

// Multiple confirms all publishings up until the delivery tag
func (c *Confirms) Multiple(confirmed Confirmation) {
	c.m.Lock()
	defer c.m.Unlock()

	for c.expecting <= confirmed.DeliveryTag {
		c.confirm(Confirmation{c.expecting, confirmed.Ack})
	}
	c.resequence()
}

// Close closes all listeners, discarding any out of sequence confirmations
func (c *Confirms) Close() error {
	c.m.Lock()
	defer c.m.Unlock()

	for _, l := range c.listeners {
		close(l)
	}
	c.listeners = nil
	return nil
}
