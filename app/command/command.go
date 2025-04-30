package command

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/vandanrohatgi/radish-go/app/config"
)

func CommandHandler(cmd []string, conn net.Conn) error {
	switch strings.ToUpper(cmd[0]) {
	case "PING":
		_, err := conn.Write([]byte("+PONG\r\n"))
		if err != nil {
			return fmt.Errorf("error writing response to connection: %w", err)
		}
	case "ECHO":
		if len(cmd) < 2 {
			return fmt.Errorf("ECHO takes 1 arguement, 0 given")
		}
		payload := "$" + strconv.Itoa(len(cmd[1])) + "\r\n" + cmd[1] + "\r\n"
		_, err := conn.Write([]byte(payload))
		if err != nil {
			return fmt.Errorf("error writing response to connection: %w", err)
		}
	case "SET":
		if len(cmd) < 3 {
			return fmt.Errorf("SET requires at least 3 arguements, %d given", len(cmd))
		}
		config.Storage[cmd[1]] = cmd[2]
		if len(cmd) == 5 && strings.ToUpper(cmd[3]) == "PX" {
			expire, err := strconv.Atoi(cmd[4])
			if err != nil {
				return fmt.Errorf("invalid expiration time provided")
			}
			go expireKey(cmd[1], expire)
		}
		_, err := conn.Write([]byte("+OK\r\n"))
		if err != nil {
			return fmt.Errorf("error writing response to connection: %w", err)
		}

	case "GET":
		if len(cmd) < 2 {
			return fmt.Errorf("GET takes 1 arguement, 0 given")
		}
		val, exists := config.Storage[cmd[1]]
		if !exists {
			_, err := conn.Write([]byte("$-1\r\n"))
			if err != nil {
				return fmt.Errorf("error writing response to connection: %w", err)
			}
			break
		}
		payload := "$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
		_, err := conn.Write([]byte(payload))
		if err != nil {
			return fmt.Errorf("error writing response to connection: %w", err)
		}
	case "CONFIG":
		if strings.ToUpper(cmd[1]) == "GET" {
			if len(cmd) < 2 {
				return fmt.Errorf("CONFIG GET takes 1 arguement, 0 given")
			}
			val, exists := config.ConfigMap[cmd[2]]
			if !exists {
				_, err := conn.Write([]byte("$-1\r\n"))
				if err != nil {
					return fmt.Errorf("error writing response to connection: %w", err)
				}
				break
			}
			payload := "*2\r\n" + "$" + strconv.Itoa(len(cmd[2])) + "\r\n" + cmd[2] + "\r\n$" + strconv.Itoa(len(val)) + "\r\n" + val + "\r\n"
			_, err := conn.Write([]byte(payload))
			if err != nil {
				return fmt.Errorf("error writing response to connection: %w", err)
			}
		}

	default:
		return errors.New("this command is not supported yet")
	}
	return nil
}

func expireKey(key string, t int) {
	time.Sleep(time.Duration(t) * time.Millisecond)
	delete(config.Storage, key)
}
