package pubsub

import (
	"fmt"
	"gopkg.in/redis.v1"
)

type Subscriber struct {
	redis_client  *redis.Client
	pubsub_client *redis.PubSub
}

func NewSubscriber(host string, port int) (*Subscriber, error) {

	redis_endpoint := fmt.Sprintf("%s:%d", host, port)

	redis_client := redis.NewTCPClient(&redis.Options{
		Addr: redis_endpoint,
	})

	pubsub_client := redis_client.PubSub()

	s := Subscriber{
		redis_client:  redis_client,
		pubsub_client: pubsub_client,
	}

	return &s, nil
}

func (s *Subscriber) Subscribe(channel string, message_ch chan string) error {

	err := s.pubsub_client.Subscribe(channel)

	if err != nil {
		return err
	}

	for {

		i, _ := s.pubsub_client.Receive()

		if msg, _ := i.(*redis.Message); msg != nil {
			message_ch <- msg.Payload
		}
	}

	return nil
}

func (s *Subscriber) Close() error {

	var err error

	err = s.redis_client.Close()

	if err != nil {
		return err
	}

	err = s.pubsub_client.Close()

	if err != nil {
		return err
	}

	return nil
}
