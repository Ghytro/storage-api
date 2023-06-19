package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"
	"os"
	"os/signal"
	"storageapi/internal/config"
)

func main() {
	server := serve()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.ListenerPort))
	if err != nil {
		log.Fatal(err)
	}

	sigIntC := make(chan os.Signal, 1)
	signal.Notify(sigIntC, os.Interrupt)
	go func() {
		for {
			select {
			case <-sigIntC:
				listener.Close()
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Print(err)
					continue
				}
				codec := jsonrpc.NewServerCodec(conn)
				go server.ServeCodec(codec)
			}
		}
	}()
}
