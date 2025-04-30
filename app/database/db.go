package database

import (
	"log"
	"os"
)

func PrepDB(dir string, dbFileName string) {
	dbFile, err := os.Create(dir + "/" + dbFileName)
	if err != nil {
		log.Fatal("Error creating RDB file: ", err)
	}
	defer dbFile.Close()
	var initialHeaders = []byte("REDIS0011\n")
	initialHeaders = append(initialHeaders, 0xFA)
	initialHeaders = append(initialHeaders, []byte("\nredis-ver\n6.0.16\n")...)
	initialHeaders = append(initialHeaders, []byte{0xFE, '\n', 0x00, 0xFB, 0x03, 0x02}...)
	_, err = dbFile.Write(initialHeaders)
	if err != nil {
		log.Fatal("Error writing to rdb file", err)
	}
}
