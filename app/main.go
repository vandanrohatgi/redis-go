package main

import (
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		fmt.Println("Failed to bind to port 6379")
		os.Exit(1)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		fmt.Println("New client found!")
		// conn.SetReadDeadline(time.Time(5 * time.Second))
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	var dataFromConn []byte
	var err error
	for {
		_, err = conn.Read(dataFromConn)
		if err != nil {
			log.Fatal("unable to read")
		}
		fmt.Println(dataFromConn)

		_, err = conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			log.Println("Error writing response to connection", err)
			return
		}
	}

}
