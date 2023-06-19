package config

import (
	"log"
	"os"
	"strconv"
)

var DatabaseURL = os.Getenv("DATABASE_URL")
var ListenerPort int

func init() {
	var err error
	if ListenerPort, err = strconv.Atoi(os.Getenv("LISTENER_PORT")); err != nil {
		log.Fatal(err)
	}
}
