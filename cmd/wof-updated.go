package main

import (
	"encoding/csv"
	"flag"
	_ "fmt"
	"github.com/whosonfirst/go-webhookd/config"
	"github.com/whosonfirst/go-webhookd/daemon"
	_ "github.com/whosonfirst/go-whosonfirst-readwrite/reader"
	"github.com/whosonfirst/go-whosonfirst-readwrite/writer"
	"github.com/whosonfirst/go-whosonfirst-redis/pubsub"
	"github.com/whosonfirst/go-whosonfirst-updated-v2"
	"github.com/whosonfirst/go-whosonfirst-updated-v2/flags"
	"github.com/whosonfirst/go-whosonfirst-updated-v2/process"
	"github.com/whosonfirst/go-whosonfirst-updated-v2/process/copy"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	var pubsub_host = flag.String("pubsub-host", "localhost", "Redis host")
	var pubsub_port = flag.Int("pubsub-port", 6379, "Redis port")
	var pubsub_channel = flag.String("pubsub-channel", "updated", "Redis channel")

	var pubsub_daemon = flag.Bool("pubsubd", false, "")

	var webhook_daemon = flag.Bool("webhookd", false, "")
	var webhook_config = flag.String("webhookd-config", "", "")

	var copy_flags flags.CopyFlags
	flag.Var(&copy_flags, "copy", "")

	flag.Parse()

	ps_messages := make(chan string)
	up_messages := make(chan updated.Task)

	if *pubsub_daemon {

		server, err := pubsub.NewServer(*pubsub_host, *pubsub_port)

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

		log.Println("Ready to receive (updated) PubSub messages")
	}

	// we start this after -pubsubd so that we can ensure a pubsub daemon to connect to

	if *webhook_daemon {

		wh_config, err := config.NewConfigFromFile(*webhook_config)

		if err != nil {
			log.Fatal(err)
		}

		wh_daemon, err := daemon.NewWebhookDaemonFromConfig(wh_config)

		if err != nil {
			log.Fatal(err)
		}

		go func() {

			log.Println("Ready to receive (updated) Webhook messages")

			err = wh_daemon.Start()

			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	// sudo make a generic receiver interface for things other than pubsub...
	// (20171227/thisisaaronland)

	sub, err := pubsub.NewSubscriber(*pubsub_host, *pubsub_port)

	if err != nil {
		log.Fatal(err)
	}

	defer sub.Close()

	go sub.Subscribe(*pubsub_channel, ps_messages)

	go func() {

		log.Println("Ready to process (updated) PubSub messages")

		for {

			msg := <-ps_messages

			msg = strings.Trim(msg, " ")

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
					// log.Printf("ROW '%s'\n", row)
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

	processors := make([]process.Processor, 0)

	var writers []writer.Writer

	null_writer, err := writer.NewNullWriter()

	if err != nil {
		log.Fatal(err)
	}

	writers = append(writers, null_writer)

	for _, fl := range copy_flags.Flags {

		if strings.HasPrefix(fl, "s3#") {

			str_cfg := strings.Replace(fl, "s3#", "", -1)
			s3_cfg, err := writer.NewS3ConfigFromString(str_cfg)

			if err != nil {
				log.Fatal(err)
			}

			s3_writer, err := writer.NewS3Writer(s3_cfg)

			if err != nil {
				log.Fatal(err)
			}

			writers = append(writers, s3_writer)
		}
	}

	multi_writer, err := writer.NewMultiWriter(writers...)

	if err != nil {
		log.Fatal(err)
	}

	gh_copier, err := copy.NewGitHubCopier(multi_writer)

	if err != nil {
		log.Fatal(err)
	}

	processors = append(processors, gh_copier)

	for {

		select {

		case t := <-up_messages:

			t1 := time.Now()

			status_ch := make(chan string)
			error_ch := make(chan error)
			done_ch := make(chan bool)

			go func() {

				for {
					select {
					case s := <-status_ch:
						log.Println(s)
					case e := <-error_ch:
						log.Println("ERROR", e)
					case <-done_ch:
						return
					}
				}
			}()

			for _, pr := range processors {
				pr.ProcessTask(t, status_ch, error_ch, done_ch)
			}

			t2 := time.Since(t1)
			log.Printf("time to process task %s %v\n", t, t2)

		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	os.Exit(0)
}
