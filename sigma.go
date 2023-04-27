package go_pdk

import "google.golang.org/protobuf/proto"

var Module string

type PDK struct {
	LOG  *Logger
	Nats Nats
	Chan Chan
}

type Scope struct {
	Local bool
	Nats  bool
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

func (pdk *PDK) PostEvent(subject string, data proto.Message, sc Scope) { // account_created
	if sc.Nats {
		pdk.Nats.PostEvent(subject, data)
		return
	}
	if sc.Local {
		pdk.Chan.PostEvent(subject, data)
		return
	}
}

type SubjectHandler[R proto.Message] func(ctx *Context, data R)

func (pdk *PDK) RegisterSubject(subject string, handler SubjectHandler[proto.Message]) {
	//register Nats
	pdk.Nats.RegisterEvent(subject, handler)

	//register chan
	pdk.Chan.RegisterEvent(subject, handler)

}
