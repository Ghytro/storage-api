package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

var DatabaseURL = os.Getenv("DATABASE_URL")
var ListenerPort int
var RequestHandleTimeout time.Duration

func init() {
	var err error
	if ListenerPort, err = strconv.Atoi(os.Getenv("LISTENER_PORT")); err != nil {
		log.Fatal(err)
	}
	reqHandleTimeoutMS, err := strconv.Atoi(os.Getenv("REQUEST_HANDLE_TIMEOUT_MS"))
	if err != nil {
		log.Fatal(err)
	}
	RequestHandleTimeout = time.Millisecond * time.Duration(reqHandleTimeoutMS)
}
