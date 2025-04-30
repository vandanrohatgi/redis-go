package network

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	cmd "github.com/vandanrohatgi/radish-go/app/command"
)

func ClientHandler(conn net.Conn) {
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
		if err != nil {
			log.Println("error processing args length:", err)
			continue
		}
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
		err = cmd.CommandHandler(command, conn)
		if err != nil {
			fmt.Println("Error handling command: ", err)
			continue
		}
	}

}
