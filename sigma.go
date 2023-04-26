package go_pdk

type PDK struct {
	module string
	LOG    *Logger
	Nats   Nats
	Chan   Chan
}

func (pdk *PDK) Release() {
	pdk.Nats.Connection.Close()
}

func (pdk *PDK) Start() {
	startEventStream(pdk.Nats.JetStream)
	startChannelStream(&pdk.Chan)
	//sig := make(chan os.Signal, 1)
	//signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	//<-sig
	select {}
}
