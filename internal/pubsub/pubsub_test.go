package pubsub

import (
	"sync"
	"testing"
	"time"
)

func TestPubSubBasic(t *testing.T) {
	ps := NewPubSub()

	ch := ps.Subscribe("test")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		msg := <-ch
		if msg != "Hello, World!" {
			t.Errorf("Expected 'Hello, World!', got '%s'", msg)
		}
	}()

	ps.Publish("test", "Hello, World!")

	wg.Wait()
}

func TestUnsubscribe(t *testing.T) {
	ps := NewPubSub()
	ch := ps.Subscribe("test")

	ps.Unsubscribe("test", ch)

	ps.Publish("test", "Message after unsubscribe")

	_, open := <-ch
	if open {
		t.Errorf("Channel should be closed after unsubscribe")
	}
}

func TestPubSubMultipleSubscribers(t *testing.T) {
	ps := NewPubSub()

	ch1 := ps.Subscribe("test")
	ch2 := ps.Subscribe("test")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		msg := <-ch1
		if msg != "Broadcast" {
			t.Errorf("Subscriber 1: expected 'Broadcast', got '%s'", msg)
		}
	}()

	go func() {
		defer wg.Done()
		msg := <-ch2
		if msg != "Broadcast" {
			t.Errorf("Subscriber 2: expected 'Broadcast', got '%s'", msg)
		}
	}()

	ps.Publish("test", "Broadcast")

	wg.Wait()
}

func TestClose(t *testing.T) {
	ps := NewPubSub()

	ch := ps.Subscribe("test")
	ps.Close()

	_, open := <-ch
	if open {
		t.Errorf("Channel should be closed after PubSub Close()")
	}
}

func TestPubSubBufferOverflow(t *testing.T) {
	ps := NewPubSub()
	ch := ps.Subscribe("test")

	for i := 0; i < 15; i++ {
		ps.Publish("test", "Spam")
	}

	time.Sleep(50 * time.Millisecond)

	select {
	case <-ch:
	default:
		t.Errorf("Expected message in buffer, but got nothing")
	}
}
