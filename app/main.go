package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

var storage = make(map[string]string)
var configMap = make(map[string]string)
var dir, dbFileName *string

func init() {
	dir = flag.String("dir", "/tmp", "directory for your rdb file")
	dbFileName = flag.String("dbFileName", "radish.rdb", "name for your rdb file")
	flag.Parse()
	configMap["dir"] = *dir
	configMap["dbFileName"] = *dbFileName
}

func main() {
	fmt.Println("Starting Radish-Go server!")
	fmt.Printf("Loaded configs: %v", configMap)

	// Prep rdb file
	dbFile, err := os.Create(*dir + "/" + *dbFileName)
	if err != nil {
		log.Fatal("Error creating RDB file: ", err)
	}
	defer dbFile.Close()
	var initialHeaders = []byte("REDIS0011\n")
	initialHeaders = append(initialHeaders, 0xFA)
	initialHeaders = append(initialHeaders, []byte("\nredis-ver\n6.0.16\n")...)
	initialHeaders = append(initialHeaders, 0xFE)
	initialHeaders = append(initialHeaders, '\n')
	initialHeaders = append(initialHeaders, 0x00)
	initialHeaders = append(initialHeaders, 0xFB)
	initialHeaders = append(initialHeaders, 0x03)
	initialHeaders = append(initialHeaders, 0x02)
	_, err = dbFile.Write(initialHeaders)
	if err != nil {
		log.Fatal("Error writing to rdb file", err)
	}

	// start networking
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
		if len(cmd) < 2 {
			return fmt.Errorf("ECHO takes 1 arguement, 0 given")
		}
		payload := "$" + strconv.Itoa(len(cmd[1])) + "\r\n" + cmd[1] + "\r\n"
		_, err := conn.Write([]byte(payload))
		if err != nil {
			return fmt.Errorf("Error writing response to connection: %w", err)
		}
	case "SET":
		if len(cmd) < 3 {
			return fmt.Errorf("SET requires at least 3 arguements, %d given", len(cmd))
		}
		storage[cmd[1]] = cmd[2]
		if len(cmd) == 5 && strings.ToUpper(cmd[3]) == "PX" {
			expire, err := strconv.Atoi(cmd[4])
			if err != nil {
				return fmt.Errorf("invalid expiration time provided")
			}
			go expireKey(cmd[1], expire)
		}
		_, err := conn.Write([]byte("+OK\r\n"))
		if err != nil {
			return fmt.Errorf("Error writing response to connection: %w", err)
		}

	case "GET":
		if len(cmd) < 2 {
			return fmt.Errorf("GET takes 1 arguement, 0 given")
		}
		val, exists := storage[cmd[1]]
		if !exists {
			_, err := conn.Write([]byte("$-1\r\n"))
			if err != nil {
				return fmt.Errorf("Error writing response to connection: %w", err)
			}
			break
		}
		payload := "$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
		_, err := conn.Write([]byte(payload))
		if err != nil {
			return fmt.Errorf("Error writing response to connection: %w", err)
		}
	case "CONFIG":
		if strings.ToUpper(cmd[1]) == "GET" {
			if len(cmd) < 2 {
				return fmt.Errorf("CONFIG GET takes 1 arguement, 0 given")
			}
			val, exists := configMap[cmd[2]]
			if !exists {
				_, err := conn.Write([]byte("$-1\r\n"))
				if err != nil {
					return fmt.Errorf("Error writing response to connection: %w", err)
				}
				break
			}
			payload := "*2\r\n" + "$" + strconv.Itoa(len(cmd[2])) + "\r\n" + cmd[2] + "\r\n$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
			_, err := conn.Write([]byte(payload))
			if err != nil {
				return fmt.Errorf("Error writing response to connection: %w", err)
			}
		}

	default:
		return errors.New("This command is not supported yet!")
	}
	return nil
}

func expireKey(key string, t int) {
	time.Sleep(time.Duration(t) * time.Millisecond)
	delete(storage, key)
}
