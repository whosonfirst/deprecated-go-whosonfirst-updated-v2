package pubsub

// https://redis.io/topics/protocol
// https://redis.io/topics/pubsub

import (
	"errors"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-redis/resp"
	"io"
	_ "log"
	"net"
	"strings"
	"sync"
)

// s.channels[ channel ][ remote_addr] = bool
type Channels map[string]bool

// s.subscriptions[ remote_addr ][ channel ] = bool
type Subscriptions map[string]bool

type Server struct {
	endpoint      string
	host          string
	port          int
	channels      map[string]Channels
	subscriptions map[string]Subscriptions
	conns         map[string]net.Conn
	mu            *sync.Mutex
	Debug         bool
}

func NewServer(host string, port int) (*Server, error) {

	conns := make(map[string]net.Conn)
	channels := make(map[string]Channels)
	subs := make(map[string]Subscriptions)

	endpoint := fmt.Sprintf("%s:%d", host, port)

	mu := new(sync.Mutex)

	s := Server{
		endpoint:      endpoint,
		host:          host,
		port:          port,
		conns:         conns,
		channels:      channels,
		subscriptions: subs,
		mu:            mu,
		Debug:         false,
	}

	return &s, nil
}

func (s *Server) ListenAndServe() error {

	daemon, err := net.Listen("tcp", s.endpoint)

	if err != nil {
		return err
	}

	defer daemon.Close()

	return s.listen(daemon)
}

func (s *Server) ListenAndServeWithReadySignal(signal chan bool) error {

	daemon, err := net.Listen("tcp", s.endpoint)

	if err != nil {
		signal <- false
		return err
	}

	defer daemon.Close()

	signal <- true

	return s.listen(daemon)
}

func (s *Server) listen(daemon net.Listener) error {

	for {

		conn, err := daemon.Accept()

		if err != nil {
			return err
		}

		go s.receive(conn)
	}

	return nil
}

func (s *Server) receive(conn net.Conn) {

	client := s.whoami(conn)
	// log.Printf("%s CONNECT", client)

	reader := resp.NewRESPReader(conn)
	writer := resp.NewRESPWriter(conn)

	if s.Debug {
		reader = resp.NewRESPDebugReader(conn)
		writer = resp.NewRESPDebugWriter(conn)
	}

	for {
		raw, err := reader.ReadObject()

		if err != nil {

			if err != io.EOF {
				// log.Printf("Failed to read from client (%s) because %s (%T)", client, err, err)
			}

			break
		}

		str_raw := strings.Trim(string(raw), " ")
		body := strings.Split(str_raw, "\r\n")

		if len(body) == 0 {
			continue
		}

		cmd := body[2]

		if cmd == "SUBSCRIBE" {

			channels := make([]string, 0)

			for _, ch := range body[3:] {

				if strings.HasPrefix(ch, "$") {
					continue
				}

				ch = strings.Trim(ch, " ")

				if ch == "" {
					continue
				}

				channels = append(channels, ch)
			}

			rsp, err := s.subscribe(conn, channels)

			if err != nil {
				writer.WriteErrorMessage(err)
				break
			}

			writer.WriteSubscribeMessage(rsp)

		} else if cmd == "UNSUBSCRIBE" {

			channels := make([]string, 0)

			for _, ch := range body[3:] {

				if strings.HasPrefix(ch, "$") {
					continue
				}

				channels = append(channels, ch)
			}

			rsp, err := s.unsubscribe(conn, channels)

			if err != nil {
				writer.WriteErrorMessage(err)
				break
			}

			writer.WriteUnsubscribeMessage(rsp)
			conn.Close()

		} else if cmd == "PUBLISH" {

			channel := body[4]

			msg := make([]string, 0)

			for _, str := range body[5:] {

				if strings.HasPrefix(str, "$") {
					continue
				}

				msg = append(msg, str)
			}

			str_msg := strings.Join(msg, " ")

			err := s.publish(channel, str_msg)

			if err != nil {
				writer.WriteErrorMessage(err)
				break
			}

			writer.WriteNullMessage()

		} else if cmd == "PING" {

			writer.WriteStringMessage("PONG")

		} else {

			msg := fmt.Sprintf("unknown command '%s'", cmd)
			err := errors.New(msg)

			writer.WriteErrorMessage(err)
			break
		}

	}

	conn.Close()

	go s.prune_client(client)
}

func (s *Server) subscribe(conn net.Conn, channels []string) ([]string, error) {

	client := s.whoami(conn)
	rsp := make([]string, 0)

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.conns[client]

	if !ok {
		s.conns[client] = conn
	}

	for _, ch := range channels {

		subs, ok := s.subscriptions[client]

		if !ok {
			subs = make(map[string]bool)
			s.subscriptions[ch] = subs
		}

		s.subscriptions[ch][client] = true

		chs, ok := s.channels[ch]

		if !ok {

			chs = make(map[string]bool)
			s.channels[ch] = chs
		}

		s.channels[ch][client] = true
		rsp = append(rsp, ch)
	}

	return rsp, nil
}

func (s *Server) unsubscribe(conn net.Conn, channels []string) ([]string, error) {

	client := s.whoami(conn)
	rsp := make([]string, 0)

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.conns[client]

	if !ok {
		msg := fmt.Sprintf("Can not find connection thingy for %s", client)
		err := errors.New(msg)
		return rsp, err
	}

	for _, ch := range channels {

		var ok bool

		_, ok = s.subscriptions[client]

		if !ok {
			continue
		}

		_, ok = s.subscriptions[client][ch]

		if !ok {
			continue
		}

		delete(s.subscriptions[client], ch)

		_, ok = s.channels[ch]

		if !ok {
			continue
		}

		_, ok = s.channels[ch][client]

		if !ok {
			continue
		}

		delete(s.channels[ch], client)

		if len(s.channels[ch]) == 0 {
			delete(s.channels, ch)
		}
	}

	return rsp, nil
}

func (s *Server) publish(channel string, message string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	clients, ok := s.channels[channel]

	if !ok {
		return nil
	}

	for client, _ := range clients {

		conn, ok := s.conns[client]

		if !ok {
			continue
		}

		go func(c net.Conn, ch string, m string) {

			writer := resp.NewRESPWriter(c)
			writer.WritePublishMessage(ch, message)

		}(conn, channel, message)

	}

	return nil
}

func (s *Server) whoami(conn net.Conn) string {

	return conn.RemoteAddr().String()
}

func (s *Server) prune_client(client string) {

	s.mu.Lock()
	defer s.mu.Unlock()

	var ok bool

	_, ok = s.conns[client]

	if ok {
		delete(s.conns, client)
	}

	_, ok = s.subscriptions[client]

	if ok {
		delete(s.subscriptions, client)
	}

	for ch, _ := range s.channels {

		_, ok = s.channels[ch][client]

		if ok {
			delete(s.channels[ch], client)
		}

		if len(s.channels[ch]) == 0 {
			delete(s.channels, ch)
		}
	}

}
