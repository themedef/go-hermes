package pubsub

import (
	"sync"
)

type Config struct {
	BufferSize int
}

type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan string]struct{}
	bufferSize  int
}

func NewPubSub(config Config) *PubSub {
	bs := 10000
	if config.BufferSize > 0 {
		bs = config.BufferSize
	}
	return &PubSub{
		subscribers: make(map[string]map[chan string]struct{}),
		bufferSize:  bs,
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

	ch := make(chan string, ps.bufferSize)
	if ps.subscribers[key] == nil {
		ps.subscribers[key] = make(map[chan string]struct{})
	}
	ps.subscribers[key][ch] = struct{}{}
	return ch
}

func (ps *PubSub) Unsubscribe(key string, ch chan string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if subscribers, exists := ps.subscribers[key]; exists {
		if _, found := subscribers[ch]; found {
			delete(subscribers, ch)
			if len(subscribers) == 0 {
				delete(ps.subscribers, key)
			}
		}
	}
	close(ch)
}

func (ps *PubSub) Publish(key, message string) {
	ps.mu.RLock()
	subscribers, exists := ps.subscribers[key]
	ps.mu.RUnlock()
	if !exists {
		return
	}
	for ch := range subscribers {
		select {
		case ch <- message:
		default:
		}
	}
}

func (ps *PubSub) UnsubscribeAllForKey(key string) {
	ps.mu.Lock()
	subscribers, exists := ps.subscribers[key]
	if exists {
		delete(ps.subscribers, key)
	}
	ps.mu.Unlock()

	if !exists {
		return
	}
	for ch := range subscribers {
		close(ch)
	}
}

func (ps *PubSub) Close() {
	ps.mu.Lock()
	subsCopy := ps.subscribers
	ps.subscribers = make(map[string]map[chan string]struct{})
	ps.mu.Unlock()

	for _, subs := range subsCopy {
		for ch := range subs {
			close(ch)
		}
	}
}
