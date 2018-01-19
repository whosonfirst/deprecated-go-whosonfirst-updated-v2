# go-whosonfirst-updated-v2

## Important

This is an experimental rewrite/refactoring of [go-whosonfirst-updated](https://github.com/whosonfirst/go-whosonfirst-updated-v2). It is too soon for you. Really...

## Example

_These examples are not meant to be comprehensive. Honestly if you're reading this these examples are as much to help me remember how things work and what we need to preserve between v1 and v2.

### Simple

```
bin/wof-updated -pubsubd -pubsub-host 0.0.0.0 -webhookd -webhookd-config docker/webhookd-config.json
2018/01/19 21:40:05 Ready to receive (updated) PubSub messages
2018/01/19 21:40:05 Ready to process (updated) PubSub messages
2018/01/19 21:40:05 Ready to receive (updated) Webhook messages
```

And then:

```
curl -s -v -X POST -d '{"foo":"bar"}' localhost:8080/insecure-test
*   Trying 127.0.0.1...
* TCP_NODELAY set
* Connected to localhost (127.0.0.1) port 8080 (#0)
> POST /insecure-test HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.57.0
> Accept: */*
> Content-Length: 13
> Content-Type: application/x-www-form-urlencoded
>
* upload completely sent off: 13 out of 13 bytes
< HTTP/1.1 200 OK
< Date: Fri, 19 Jan 2018 21:41:15 GMT
< Content-Length: 0
< Content-Type: text/plain; charset=utf-8
<
* Connection #0 to host localhost left intact
```

Which will output (to STDOUT):

```
2018/01/19 21:41:15 {"foo":"bar"}
```

If you call `curl -s -v -X POST -d '{"foo":"bar"}' localhost:8080/chicken-test` instead you'll end up with this:

```
2018/01/19 21:46:01 {"🐔":"🐔"}
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-updated
