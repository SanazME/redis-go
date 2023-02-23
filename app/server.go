package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func exitWithError(err error) {
	fmt.Println(err.Error())
	os.Exit(1)
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		exitWithError(fmt.Errorf("failed to bind to port 6379: %v", err))
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			exitWithError(fmt.Errorf("Error accepting connection: %v", err))
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	for {
		b := make([]byte, 2056)
		_, err := conn.Read(b)

		if err == io.EOF {
			break
		}

		if err != nil {
			exitWithError(fmt.Errorf("Error reading from stream %v", err))
		}

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			exitWithError(fmt.Errorf("Error writing to conneciton %v", err))
		}
	}
}
