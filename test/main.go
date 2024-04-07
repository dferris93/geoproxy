package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

func main() {

	user, _ := os.LookupEnv("USER")
	keyfile := flag.String("keyfile", "id_rsa", "Path to private key file")
	username := flag.String("username", user, "Username for SSH connection")
	serverAddr := flag.String("server", "", "Server address for SSH connection")
	port := flag.String("port", "22", "Port for SSH connection")
	numConnections := flag.Int("num", 1, "Number of SSH connections to establish")

	flag.Parse()

	// Load private key
	key, err := os.ReadFile(*keyfile)
	if err != nil {
		log.Fatalf("Failed to load private key: %s", err)
	}

	fmt.Print("Enter passphrase for private key: ")
	passPhrase, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatalf("Failed to read passphrase: %s", err)
	}
	signer, err := ssh.ParsePrivateKeyWithPassphrase(key, passPhrase)
	if err != nil {
		log.Fatalf("Failed to parse private key: %s", err)
	}

	// SSH configuration with key-based authentication
	config := &ssh.ClientConfig{
		User: *username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		Timeout: 5 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: Insecure, use for testing only
	}

	// SSH server address
	addr := *serverAddr
	sshPort := *port
	
	sshTuple := fmt.Sprintf("%s:%s", addr, sshPort)

	// Create SSH connections
	wg := sync.WaitGroup{}
	for i := 0; i < *numConnections; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Establish SSH connection
			conn, err := ssh.Dial("tcp", sshTuple, config)
			if err != nil {
				log.Fatalf("Failed to dial: %s", err)
			}
			defer conn.Close()

			// Create session
			session, err := conn.NewSession()
			if err != nil {
				log.Printf("Failed to create session: %s", err)
				return
			}
			defer session.Close()

			// Run command
			output, err := session.CombinedOutput("hostname")
			if err != nil {
				log.Printf("Failed to run command: %s", err)
				return
			}

			// Print output
			fmt.Printf("%s\n", string(output))
		}()
	}
	fmt.Printf("Established %d SSH connections\n", *numConnections)
	wg.Wait()
}

