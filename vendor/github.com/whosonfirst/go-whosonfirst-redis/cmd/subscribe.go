package main

import (
	"bufio"
	"flag"
	"fmt"
	"gopkg.in/redis.v1"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {

	var redis_host = flag.String("redis-host", "localhost", "The Redis host to connect to.")
	var redis_port = flag.Int("redis-port", 6379, "The Redis port to connect to.")
	var redis_channel = flag.String("redis-channel", "", "The Redis channel to publish to.")

	var stdout = flag.Bool("stdout", false, "Output messages to STDOUT. If no other output options are defined this is enabled by default.")

	flag.Parse()

	if *redis_channel == "" {
		log.Fatal("Missing channel")
	}

	writers := make([]io.Writer, 0)

	if *stdout {
		writers = append(writers, os.Stdout)
	}

	if len(writers) == 0 {
		writers = append(writers, os.Stdout)
	}

	multi := io.MultiWriter(writers...)
	writer := bufio.NewWriter(multi)

	redis_endpoint := fmt.Sprintf("%s:%d", *redis_host, *redis_port)

	redis_client := redis.NewTCPClient(&redis.Options{
		Addr: redis_endpoint,
	})

	defer redis_client.Close()

	_, err := redis_client.Ping().Result()

	if err != nil {
		log.Fatal("Failed to ping Redis server ", err)
	}

	pubsub_client := redis_client.PubSub()
	defer pubsub_client.Close()

	err = pubsub_client.Subscribe(*redis_channel)

	if err != nil {
		msg := fmt.Sprintf("Failed to subscribe to channel %s, because %s", *redis_channel, err)
		log.Fatal(msg)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs

		err = pubsub_client.Unsubscribe(*redis_channel)

		if err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	for {

		i, _ := pubsub_client.Receive()

		if msg, _ := i.(*redis.Message); msg != nil {
			writer.WriteString(msg.Payload + "\n")
			writer.Flush()
		}
	}

	os.Exit(0)
}
