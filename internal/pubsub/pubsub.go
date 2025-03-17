package pubsub

import (
	"sync"
)

type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan string]struct{}
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string]map[chan string]struct{}),
	}
}

func (ps *PubSub) ListSubscribers() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	keys := make([]string, 0, len(ps.subscribers))
	for key := range ps.subscribers {
		keys = append(keys, key)
	}

	return keys
}

func (ps *PubSub) Subscribe(key string) chan string {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan string, 100)
	if ps.subscribers[key] == nil {
		ps.subscribers[key] = make(map[chan string]struct{})
	}
	ps.subscribers[key][ch] = struct{}{}

	return ch
}

func (ps *PubSub) Unsubscribe(key string, ch chan string) {
	ps.mu.Lock()
	if subscribers, exists := ps.subscribers[key]; exists {
		if _, found := subscribers[ch]; found {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(ps.subscribers, key)
			}
		}
	}
	ps.mu.Unlock()

	close(ch)
}

func (ps *PubSub) Publish(key, message string) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if subscribers, exists := ps.subscribers[key]; exists {
		for ch := range subscribers {
			select {
			case ch <- message:
			default:
			}
		}
	}
}

func (ps *PubSub) UnsubscribeAllForKey(key string) {
	ps.mu.Lock()
	subscribers, exists := ps.subscribers[key]
	if !exists {
		ps.mu.Unlock()
		return
	}
	delete(ps.subscribers, key)
	ps.mu.Unlock()

	for ch := range subscribers {
		close(ch)
	}
}

func (ps *PubSub) Close() {
	ps.mu.Lock()
	subscribersCopy := make(map[string]map[chan string]struct{})
	for key, subs := range ps.subscribers {
		subscribersCopy[key] = subs
	}
	ps.subscribers = make(map[string]map[chan string]struct{})
	ps.mu.Unlock()

	for _, subs := range subscribersCopy {
		for ch := range subs {
			close(ch)
		}
	}
}
