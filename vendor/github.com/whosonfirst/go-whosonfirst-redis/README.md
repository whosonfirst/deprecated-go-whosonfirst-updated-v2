# go-whosonfirst-redis


## Install

You will need to have both `Go` and the `make` programs installed on your computer. Assuming you do just type:

```
make bin
```

All of this package's dependencies are bundled with the code in the `vendor` directory.

## Tools

### pubsubd

```
./bin/pubsubd -h
Usage of ./bin/pubsubd:
  -debug
    	Print all RESP commands to STDOUT.
  -host string
    	The hostname to listen on. (default "localhost")
  -port int
    	The port number to listen on. (default 6379)
```

This will launch a daemon to support most (but not all) of the [Redis Publish/Subscribe protocol](https://redis.io/topics/pubsub). It has not been tested for load or scale but it works. The following commands are supported: `PING, SUBSCRIBE, UNSUBSCRIBE, PUBLISH`

### publish

```
./bin/publish  -h
Usage of ./bin/publish:
  -debug
    	Print all RESP commands to STDOUT (only really useful if you have invoked the -pubsubd flag).
  -pubsubd
    	Invoke a local pubsubd daemon that publish and subscribe clients (or at least the publish client) will
	connect to. This may be useful when you don't have a local copy of Redis around.
  -redis-channel string
    	The Redis channel to publish to.
  -redis-host string
    	The Redis host to connect to. (default "localhost")
  -redis-port int
    	The Redis port to connect to. (default 6379)
```

Publish a message to PubSub channel. If the message is `-` then the client will read and publish all subsequent input from STDIN. For example:

```
./bin/publish -redis-channel debug -pubsubd -debug -
*1
$4
PING
+PONG
*1
$4
PING
+PONG
*2
$9
SUBSCRIBE
$5
debug
*3
$9
subscribe
$5
debug
:1
hello world
*3
$7
PUBLISH
$5
debug
$11
hello world
$-1
*2
$11
UNSUBSCRIBE
$5
debug
```

### subscribe

```
./bin/subscribe -h
Usage of ./bin/subscribe:
  -redis-channel string
    	The Redis channel to publish to.
  -redis-host string
    	The Redis host to connect to. (default "localhost")
  -redis-port int
    	The Redis port to connect to. (default 6379)
  -stdout
    	Output messages to STDOUT. If no other output options are defined this is enabled by default.
```

Subscribe to a PubSub channel and print the result to `STDOUT`. For example:

```
./bin/subscribe -redis-channel debug
hello world
^C
```

#### Output options

If no other output options are defined then all PubSub messages are written to STDOUT.

* _stdout_ - Write output to STDOUT. If you have chosen another output option and still want to write messages to STDOUT you will need to pass this flag.

For example:

```
./bin/subscribe -redis-channel debug -tts -tts-engine osx
[ imagine your computer saying "hello world" here ]
```

## See also

* https://github.com/whosonfirst/redis-tools
* https://redis.io/topics/pubsub
