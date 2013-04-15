package amqp

import (
	"sync"
)

// Synchronized channel id to channel map with sequence generation
type channelRegistry struct {
	sync.RWMutex
	channels map[uint16]*Channel
	sequence uint16
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
// Collisions with existing keys are not considered.
func (m *channelRegistry) next() uint16 {
	m.Lock()
	defer m.Unlock()
	m.sequence++
	return m.sequence
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
