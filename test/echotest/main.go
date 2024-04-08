package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"sync"
)

func startServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server started on port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading from connection:", err)
			return
		}
		fmt.Print("Received message: ", string(message))
		conn.Write([]byte(message))
	}
}

func startClient(serverIP, port string, numGoroutines int) {
	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			conn, err := net.Dial("tcp", serverIP+":"+port)
			if err != nil {
				fmt.Printf("Error connecting to server from goroutine %d: %s\n", id, err)
				return
			}
			defer conn.Close()
			fmt.Printf("Goroutine %d connected to server at %s:%s\n", id, serverIP, port)

			message := "HELLOTEST\n"
			for {
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

	flag.Parse()

	if *serverMode {
		startServer(*port)
	} else if *clientMode {
		startClient(*serverIP, *port, *numGoroutines)
	} else {
		fmt.Println("Please specify -server or -client mode.")
	}
}