package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/vandanrohatgi/radish-go/app/config"
	"github.com/vandanrohatgi/radish-go/app/network"
)

func init() {
	var dir, dbFileName string
	flag.StringVar(&dir, "dir", "/tmp", "directory for your rdb file")
	flag.StringVar(&dbFileName, "dbFileName", "radish.rdb", "name for your rdb file")
	flag.Parse()
	config.ConfigMap["dir"] = dir
	config.ConfigMap["dbFileName"] = dbFileName
}

func main() {
	fmt.Println("Starting Radish-Go server!")
	fmt.Printf("Loaded configs: %v", config.ConfigMap)

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
		go network.ClientHandler(conn)
	}
}
