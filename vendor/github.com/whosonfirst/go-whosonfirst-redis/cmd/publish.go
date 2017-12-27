package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-redis/pubsub"
	"gopkg.in/redis.v1"
	"log"
	"os"
	"strings"
)

func main() {

	var redis_host = flag.String("redis-host", "localhost", "The Redis host to connect to.")
	var redis_port = flag.Int("redis-port", 6379, "The Redis port to connect to.")
	var redis_channel = flag.String("redis-channel", "", "The Redis channel to publish to.")
	var pubsubd = flag.Bool("pubsubd", false, "Invoke a local pubsubd daemon that publish and subscribe clients (or at least the publish client) will connect to. This may be useful when you don't have a local copy of Redis around.")
	var debug = flag.Bool("debug", false, "Print all RESP commands to STDOUT (only really useful if you have invoked the -pubsubd flag).")

	flag.Parse()

	if *redis_channel == "" {
		log.Fatal("Missing channel")
	}

	redis_endpoint := fmt.Sprintf("%s:%d", *redis_host, *redis_port)

	if *pubsubd {

		server, err := pubsub.NewServer(*redis_host, *redis_port)
		server.Debug = *debug

		if err != nil {
			log.Fatal(err)
		}

		go func() {

			err := server.ListenAndServe()

			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	redis_client := redis.NewTCPClient(&redis.Options{
		Addr: redis_endpoint,
	})

	defer redis_client.Close()

	_, err := redis_client.Ping().Result()

	if err != nil {
		log.Fatal("Failed to ping Redis server ", err)
	}

	args := flag.Args()

	if len(args) == 1 && args[0] == "-" {

		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			msg := scanner.Text()
			redis_client.Publish(*redis_channel, msg)
		}

	} else {

		msg := strings.Join(args, " ")
		redis_client.Publish(*redis_channel, msg)
	}

	os.Exit(0)
}
