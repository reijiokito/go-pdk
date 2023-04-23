package go_pdk

import (
	"github.com/nats-io/nats.go"
	"os"
	"os/signal"
	"syscall"
)

var Connection *nats.Conn
var JetStream nats.JetStreamContext
var module string
var LOG *Logger

func Release() {
	Connection.Close()
}

func Start() {
	startEventStream()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}
