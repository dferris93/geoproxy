package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"net"
	"sync"
	"time"
)

func startServer(ctx context.Context, port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server started on port", port)

	tcpListener, _ := listener.(*net.TCPListener)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		if tcpListener != nil {
			_ = tcpListener.SetDeadline(time.Now().Add(1 * time.Second))
		}
		conn, err := listener.Accept()
		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				continue
			}
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(ctx, conn)
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}
		fmt.Print("Received message: ", string(message))
		conn.Write([]byte(message))
	}
}

func startClient(ctx context.Context, serverIP, port string, numGoroutines int, numMessages int) {
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			dialer := net.Dialer{}
			conn, err := dialer.DialContext(ctx, "tcp", serverIP+":"+port)
			if err != nil {
				fmt.Printf("Error connecting to server from goroutine %d: %s\n", id, err)
				return
			}
			defer conn.Close()
			fmt.Printf("Goroutine %d connected to server at %s:%s\n", id, serverIP, port)

			message := "HELLOTEST\n"
			for j := 0; j < numMessages; j++ {
				select {
				case <-ctx.Done():
					return
				default:
				}
				_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
				_, err = conn.Write([]byte(fmt.Sprintf("Goroutine %d: %s", id, message)))
				if err != nil {
					fmt.Printf("Error sending message from goroutine %d: %s\n", id, err)
					break
				}

				response, err := bufio.NewReader(conn).ReadString('\n')
				if err != nil {
					fmt.Printf("Error receiving response in goroutine %d: %s\n", id, err)
					break
				}
				fmt.Printf("Goroutine %d received from server: %s", id, response)
			}
		}(i)
	}

	wg.Wait()
}

func main() {
	serverMode := flag.Bool("server", false, "Start in server mode")
	clientMode := flag.Bool("client", false, "Start in client mode")
	port := flag.String("port", "8080", "Port to use")
	serverIP := flag.String("serverIP", "localhost", "Server IP to connect to")
	numGoroutines := flag.Int("n", 1, "Number of goroutines for the client")
	numMessages := flag.Int("messages", 1, "Messages per goroutine")
	timeout := flag.Duration("timeout", 30*time.Second, "Timeout for server/client")

	flag.Parse()

	if *serverMode {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		startServer(ctx, *port)
	} else if *clientMode {
		ctx, cancel := context.WithTimeout(context.Background(), *timeout)
		defer cancel()
		startClient(ctx, *serverIP, *port, *numGoroutines, *numMessages)
	} else {
		fmt.Println("Please specify -server or -client mode.")
	}
}
