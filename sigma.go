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

func (pdk *PDK) SendData(fromPlugin string, subject string, payload interface{}) error {
	// create a new Data object and send it to the channel
	data := Data{
		FromPlugin: fromPlugin,
		Subject:    subject,
		Payload:    payload,
	}
	pdk.dataChan <- data
	return nil
}
