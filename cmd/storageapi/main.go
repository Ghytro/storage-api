package main

import (
	"fmt"
	"log"
	"net"
	"storageapi/internal/config"
)

func main() {
	server := serve()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.ListenerPort))
	if err != nil {
		log.Fatal(err)
	}

	go func() {

	}()
}
