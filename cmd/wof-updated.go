package main

import (
       "encoding/csv"
	"flag"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"	
	// "github.com/whosonfirst/go-whosonfirst-updated-v2/processor"
	"gopkg.in/redis.v1"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	var redis_host = flag.String("redis-host", "localhost", "Redis host")
	var redis_port = flag.Int("redis-port", 6379, "Redis port")
	var redis_channel = flag.String("redis-channel", "updated", "Redis channel")

	ps_messages := make(chan string)
	up_messages := make(chan updated.Task)

	// sudo make a generic receiver interface for things other than pubsub...
	// (20171227/thisisaaronland)
	
	go func() {

		redis_endpoint := fmt.Sprintf("%s:%d", *redis_host, *redis_port)

		redis_client := redis.NewTCPClient(&redis.Options{
			Addr: redis_endpoint,
		})

		defer redis_client.Close()

		pubsub_client := redis_client.PubSub()
		defer pubsub_client.Close()

		err := pubsub_client.Subscribe(*redis_channel)

		if err != nil {
			log.Fatal(err)
		}

		log.Println("Ready to receive (updated) PubSub messages")

		for {

			i, _ := pubsub_client.Receive()

			if msg, _ := i.(*redis.Message); msg != nil {
				ps_messages <- msg.Payload
			}
		}

	}()

	go func() {

		log.Println("Ready to process (updated) PubSub messages")

		for {

			// we are assuming this:
			// https://github.com/whosonfirst/go-webhookd/blob/master/transformations/github.commits.go

			msg := <-ps_messages

			rdr := csv.NewReader(strings.NewReader(msg))

			tasks := make(map[string]map[string][]string)

			for {
				row, err := rdr.Read()

				if err == io.EOF {
					break
				}

				if err != nil {
					log.Println("Failed to read data", err)
					break
				}

				if len(row) != 3 {
					log.Println("No idea how to process row", row)
					continue
				}

				hash := row[0]
				repo := row[1]
				path := row[2]

				_, ok := tasks[repo]

				if !ok {
					tasks[repo] = make(map[string][]string)
				}

				commits, ok := tasks[repo][hash]

				if !ok {
					commits = make([]string, 0)
				}

				commits = append(commits, path)
				tasks[repo][hash] = commits
			}

			for repo, details := range tasks {

				for hash, commits := range details {

					t := updated.Task{
						Hash:    hash,
						Repo:    repo,
						Commits: commits,
					}

					up_messages <- t
				}
			}
		}
	}()

	for {

		select {

		       case t := <- up_messages:
		       	    log.Println(t)
		       default:
			    // pass
	        }
	}
	
	os.Exit(0)
}
