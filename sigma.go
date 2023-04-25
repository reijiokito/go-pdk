package go_pdk

import (
	"github.com/nats-io/nats.go"
)

type PDK struct {
	Connection *nats.Conn
	module     string
	LOG        *Logger
	JetStream  nats.JetStreamContext
}

func (pdk *PDK) Release() {
	pdk.Connection.Close()
}

func (pdk *PDK) Start() {
	startEventStream(pdk.JetStream)
	select {}
}
