package main

import (
	"encoding/csv"
	"flag"
	_ "fmt"
	"github.com/whosonfirst/go-whosonfirst-redis/pubsub"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
	"github.com/whosonfirst/go-whosonfirst-updated-v2/processor"
	"io"
	"log"
	"os"
	"strings"
)

func main() {

	var redis_host = flag.String("redis-host", "localhost", "Redis host")
	var redis_port = flag.Int("redis-port", 6379, "Redis port")
	var redis_channel = flag.String("redis-channel", "updated", "Redis channel")
	var pubsub_daemon = flag.Bool("pubsubd", false, "")

	flag.Parse()

	ps_messages := make(chan string)
	up_messages := make(chan updated.Task)

	if *pubsub_daemon {

		server, err := pubsub.NewServer(*redis_host, *redis_port)

		if err != nil {
			log.Fatal(err)
		}

		ready := make(chan bool)

		go func() {

			err := server.ListenAndServeWithReadySignal(ready)

			if err != nil {
				log.Fatal(err)
			}
		}()

		sig := <-ready

		if !sig {
			log.Fatal("Received negative ready signal from PubSub server")
		}
	}

	// sudo make a generic receiver interface for things other than pubsub...
	// (20171227/thisisaaronland)

	sub, err := pubsub.NewSubscriber(*redis_host, *redis_port)

	if err != nil {
		log.Fatal(err)
	}

	defer sub.Close()

	go sub.Subscribe(*redis_channel, ps_messages)

	go func() {

		log.Println("Ready to process (updated) PubSub messages")

		for {

			// we are assuming this:
			// https://github.com/whosonfirst/go-webhookd/blob/master/transformations/github.commits.go

			msg := <-ps_messages
			
			msg = strings.Trim(msg, " ")

			// log.Printf("GOT MESSAGE '%s'\n", msg)
			
			if msg == "" {
				continue
			}	
			
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

				if strings.HasPrefix(path, "data/") {
					path = strings.Replace(path, "data/", "", -1)
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

	processors := make([]processor.Processor, 0)

	cp, err := processor.NewCopyProcessor()

	if err != nil {
		log.Fatal(err)
	}

	processors = append(processors, cp)

	for {

		select {

		case t := <-up_messages:

			for _, pr := range processors {
				pr.ProcessTask(t)
			}

		default:
			// pass
		}
	}

	os.Exit(0)
}
