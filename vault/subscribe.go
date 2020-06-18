package vault

import "sync"

// Event from vault.
type Event interface{}

// UnlockEvent when keyring is unlocked.
type UnlockEvent struct {
	Provision *Provision
}

// LockEvent when keyring is locked.
type LockEvent struct{}

// SetEvent when item is saved.
type SetEvent struct {
	ID string
}

type subscribers struct {
	sync.Mutex
	subs map[string]chan Event
}

func newSubscribers() *subscribers {
	return &subscribers{
		subs: map[string]chan Event{},
	}
}

// TODO: Fix subscribers

// // Subscribe to topic.
// func (v *Vault) Subscribe(topic string) chan Event {
// 	return v.subs.Subscribe(topic)
// }

// // Unsubscribe from topic.
// func (v *Vault) Unsubscribe(topic string) {
// 	v.subs.Unsubscribe(topic)
// }

// Subscribe to events.
func (s *subscribers) Subscribe(topic string) chan Event {
	s.Lock()
	defer s.Unlock()
	c := make(chan Event, 2)
	s.subs[topic] = c
	return c
}

// Unsubscribe from events.
func (s *subscribers) Unsubscribe(topic string) {
	s.Lock()
	defer s.Unlock()

	delete(s.subs, topic)
}

// func (s *subscribers) notify(event Event) {
// 	s.Lock()
// 	defer s.Unlock()

// 	for _, c := range s.subs {
// 		// TODO: This will block if buffer is met
// 		c <- event
// 	}
// }
