package tcp

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type server struct {
	wg            sync.WaitGroup
	host          string
	port          int
	graceDuration time.Duration
	handlerFunc   HandlerFunc
}

type Option func(*server)

type HandlerFunc func(ctx context.Context, request string) string

func New(opts ...Option) *server {
	//default values for server configuration
	s := &server{
		host:          "localhost",
		port:          8080,
		graceDuration: 3 * time.Second,
		wg:            sync.WaitGroup{},
	}

	for _, opt := range opts {
		opt(s)
	}
	return s
}

func WithHost(host string) Option {
	return func(s *server) {
		s.host = host
	}
}

func WithPort(port int) Option {
	return func(s *server) {
		s.port = port
	}
}

func WithGraceDuration(dur time.Duration) Option {
	return func(s *server) {
		s.graceDuration = dur
	}
}

func WithHandler(handler HandlerFunc) Option {
	return func(s *server) {
		s.handlerFunc = handler
	}
}

func (s *server) Start(ctx context.Context) error {
	shutdownCtx, shCancel := context.WithCancel(context.Background())
	defer shCancel()

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return err
	}
	defer listener.Close()

	// register gracefull shutdown flow
	go s.gracefulShutdown(ctx, listener, shCancel)

	log.Printf("Starting on port: [%d]", s.port)

	// processing connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Printf("accept conn: %v", err)
			continue
		}

		s.wg.Go(func() { s.handleConn(shutdownCtx, conn) })
	}
	s.wg.Wait()

	return nil
}

func (s *server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	go func() {
		<-ctx.Done()
		if err := conn.SetReadDeadline(time.Now()); err != nil {
			log.Printf("failed to set read deadline: %v", err)
		}
	}()

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		response := s.handlerFunc(ctx, scanner.Text())

		if _, err := fmt.Fprintf(conn, "%s\n", response); err != nil {
			return
		}

		if ctx.Err() != nil {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("scanner error: %v", err)
	}
}

func (s *server) gracefulShutdown(ctx context.Context, listener net.Listener, cancel context.CancelFunc) {
	<-ctx.Done()

	log.Print("Shutting down... stop accepting new connections")
	_ = listener.Close()

	timer := time.NewTimer(s.graceDuration)
	defer timer.Stop()

	<-timer.C

	log.Print("Grace period expired. Cancelling active requests")
	cancel()
}
