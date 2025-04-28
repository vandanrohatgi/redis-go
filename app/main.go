package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	fmt.Println("Starting Radish-Go server!")

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
		conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	var conReader = bufio.NewReader(conn)
	for {
		// read the first '*' byte
		start, err := conReader.ReadByte()
		if err != nil || start != '*' {
			// skip to the starting of a new command if past command errored out
			// also skip if there is nothing to read
			continue
		}

		// var commandLength int
		args, err := conReader.ReadString('\n')
		if err != nil {
			log.Println("error reading args length: ", err)
			continue
		}
		argsLength, err := strconv.Atoi(strings.Trim(args, "\r\n"))
		var command []string
		for range argsLength {
			// skip the length data
			conReader.ReadString('\n')
			args, err := conReader.ReadString('\n')
			if err != nil {
				log.Println("error reading args: ", err)
				break
			}
			command = append(command, strings.Trim(args, "\r\n"))
		}
		fmt.Println(command)
		err = handleCommand(command, conn)
		if err != nil {
			fmt.Println("Error handling command: ", err)
			continue
		}
	}

}

func handleCommand(cmd []string, conn net.Conn) error {
	switch strings.ToUpper(cmd[0]) {
	case "PING":
		_, err := conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			return fmt.Errorf("Error writing response to connection: %w", err)
		}
	case "ECHO":
		payload := "$" + strconv.Itoa(len(cmd[1])) + "\r\n" + cmd[1] + "\r\n"
		_, err := conn.Write([]byte(payload))
		if err != nil {
			return fmt.Errorf("Error writing response to connection: %w", err)
		}
	default:
		return errors.New("This command is not supported yet!")
	}
	return nil
}
