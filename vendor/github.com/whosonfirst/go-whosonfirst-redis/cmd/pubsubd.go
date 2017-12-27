package main

import (
	"flag"
	"github.com/whosonfirst/go-whosonfirst-redis/pubsub"
	"log"
	"os"
)

func main() {

	var host = flag.String("host", "localhost", "The hostname to listen on.")
	var port = flag.Int("port", 6379, "The port number to listen on.")
	var debug = flag.Bool("debug", false, "Print all RESP commands to STDOUT.")

	flag.Parse()

	server, err := pubsub.NewServer(*host, *port)

	if err != nil {
		log.Fatal(err)
	}

	server.Debug = *debug

	err = server.ListenAndServe()

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
