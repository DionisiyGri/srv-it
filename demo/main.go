package main

import (
	"context"
	"log"
	"os/signal"
	"srv-it/tcp"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	tcpServer := tcp.New(
		tcp.WithHost("0.0.0.0"),
		tcp.WithGraceDuration(4*time.Second),
		tcp.WithHandler(Handle),
	)
	if err := tcpServer.Start(ctx); err != nil {
		log.Printf("server start error: %v", err)
	}
}
