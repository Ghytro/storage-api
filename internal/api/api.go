package api

import "time"

type ApiConf struct {
	RequestHandleTimeout time.Duration
}

type Empty struct {
}
