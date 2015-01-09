package amqp

import (
	"sync"
)

// Synchronized channel id to channel map with sequence generation
type channelRegistry struct {
	sync.RWMutex
	channels map[uint16]*Channel
	sequence uint16
	max      uint16
}

// Overwrites the channel at the given id
func (m *channelRegistry) add(id uint16, c *Channel) *Channel {
	m.Lock()
	defer m.Unlock()
	m.channels[id] = c
	return c
}

// Returns nil if not found
func (m *channelRegistry) get(id uint16) *Channel {
	m.RLock()
	defer m.RUnlock()
	return m.channels[id]
}

// Deletes this id from the channel map, noop if the channel id doesn't exist
func (m *channelRegistry) remove(id uint16) {
	m.Lock()
	defer m.Unlock()
	delete(m.channels, id)
}

// Returns the next sequence to be used for a channel id, starting at 1.
func (m *channelRegistry) next() (uint16, error) {
	m.Lock()
	defer m.Unlock()
	id := (m.sequence + 1) % m.max
	for ; m.channels[id] != nil; id = (id + 1) % m.max {
		if id == m.sequence { // when wrapped around
			return 0, ErrChannelMax
		}
	}
	m.sequence = id
	return m.sequence, nil
}

// Removes all channels, returning the channels removed
func (m *channelRegistry) removeAll() []*Channel {
	m.Lock()
	defer m.Unlock()

	removed := make([]*Channel, 0, len(m.channels))
	for id, c := range m.channels {
		removed = append(removed, c)
		delete(m.channels, id)
	}

	return removed
}
