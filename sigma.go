package go_pdk

import (
	"github.com/nats-io/nats.go"
)

type PDK struct {
	Connection *nats.Conn
	module     string
	LOG        *Logger
	JetStream  nats.JetStreamContext
	dataChan   chan Data
}

func (pdk *PDK) Release() {
	pdk.Connection.Close()
}

func (pdk *PDK) Start() {
	startEventStream(pdk.JetStream)
	//sig := make(chan os.Signal, 1)
	//signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	//<-sig
	select {}
}
