package contracts

type PubSub interface {
	Subscribe(key string) chan string
	Unsubscribe(key string, ch chan string)
	Publish(key, message string)
	ListSubscribers() []string
	UnsubscribeAllForKey(key string)
	Close()
}
