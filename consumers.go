// Copyright (c) 2012, Sean Treadway, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/streadway/amqp

package amqp

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
)

var consumerSeq uint64

func uniqueConsumerTag() string {
	return fmt.Sprintf("ctag-%s-%d", os.Args[0], atomic.AddUint64(&consumerSeq, 1))
}

type consumerChannels map[string]chan Delivery

// Concurrent type that manages the consumerTag ->
// consumerChannel mapping
type consumers struct {
	sync.Mutex
	chans consumerChannels
}

func makeConsumers() consumers {
	return consumers{chans: make(consumerChannels)}
}

// On key conflict, close the previous channel.
func (me *consumers) add(tag string, ch chan Delivery) {
	me.Lock()
	defer me.Unlock()

	if prev, found := me.chans[tag]; found {
		close(prev)
	}

	me.chans[tag] = ch
}

func (me *consumers) close(tag string) (found bool) {
	me.Lock()
	defer me.Unlock()

	ch, found := me.chans[tag]

	if found {
		delete(me.chans, tag)
		close(ch)
	}

	return found
}

func (me *consumers) closeAll() {
	me.Lock()
	defer me.Unlock()

	for _, ch := range me.chans {
		close(ch)
	}

	me.chans = make(consumerChannels)
}

// Sends a delivery to a the consumer identified by `tag`.
// If unbuffered channels are used for Consume this method
// could block all deliveries until the consumer
// receives on the other end of the channel.
func (me *consumers) send(tag string, msg *Delivery) bool {
	me.Lock()
	defer me.Unlock()

	ch, found := me.chans[tag]
	if found {
		ch <- *msg
	}

	return found
}
