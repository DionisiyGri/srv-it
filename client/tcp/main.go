package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func main() {
	// Start with the server already running on localhost:8080.

	// 1. Normal request <= 100: should return immediately.
	con1 := mustCreateConn("localhost:8080")
	defer con1.Close()

	sendAndPrint(con1, "PAYMENT|100")

	// 2. Normal delayed request: 1000 should take ~1 second.
	con2 := mustCreateConn("localhost:8080")
	defer con2.Close()

	sendAndPrint(con2, "PAYMENT|1000")

	// 3. Multiple requests over the same connection.
	con3 := mustCreateConn("localhost:8080")
	defer con3.Close()

	sendAndPrint(con3, "PAYMENT|100")
	sendAndPrint(con3, "PAYMENT|15000")
	sendAndPrint(con3, "PAYMENT|300")

	// 4. Invalid request format.
	con4 := mustCreateConn("localhost:8080")
	defer con4.Close()

	sendAndPrint(con4, "INVALID|100")

	// 5. Invalid amount.
	con5 := mustCreateConn("localhost:8080")
	defer con5.Close()

	sendAndPrint(con5, "PAYMENT|abc")

	// 6. Zero / negative amounts.
	con6 := mustCreateConn("localhost:8080")
	defer con6.Close()

	sendAndPrint(con6, "PAYMENT|0")
	sendAndPrint(con6, "PAYMENT|-10")

	// 7. Multiple connections concurrently.
	runConcurrentRequests()

	// 8. Idle connection:
	// Keep this connection open without sending anything.
	// Start server shutdown manually and verify that it exits
	// after the grace period instead of hanging forever.
	idleConn := mustCreateConn("localhost:8080")
	defer idleConn.Close()

	log.Println("Idle connection created. Start server shutdown now.")
	time.Sleep(10 * time.Second)
}

func mustCreateConn(address string) net.Conn {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("failed to connect to %s: %v", address, err)
	}

	log.Printf("connected to %s", address)
	return conn
}

func sendAndPrint(conn net.Conn, request string) {
	log.Printf("sending: %q", request)

	if _, err := fmt.Fprint(conn, request+"\n"); err != nil {
		log.Printf("failed to send %q: %v", request, err)
		return
	}

	reader := bufio.NewReader(conn)

	response, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("failed to read response for %q: %v", request, err)
		return
	}

	log.Printf("response: %q", response)
}

func runConcurrentRequests() {
	requests := []string{
		"PAYMENT|200",
		"PAYMENT|500",
		"PAYMENT|1000",
		"PAYMENT|100",
		"PAYMENT|50",
	}

	var wg sync.WaitGroup

	for _, request := range requests {
		request := request

		wg.Add(1)

		go func() {
			defer wg.Done()

			conn := mustCreateConn("localhost:8080")
			defer conn.Close()

			sendAndPrint(conn, request)
		}()
	}

	wg.Wait()
}
