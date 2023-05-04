package go_pdk

import (
	"github.com/reijiokito/go-pdk/channel"
	"github.com/reijiokito/go-pdk/nats"
	"github.com/reijiokito/go-pdk/server"
	"google.golang.org/protobuf/proto"
)

var Module string
var Server *server.PluginServer

type Context struct {
	Logger
}

type PDK struct {
	LOG  *Logger
	Nats nats.Nats
	Chan channel.Chan
}

type Scope struct {
	Local bool
	Nats  bool
}

func (pdk *PDK) Release() {
	pdk.Nats.Connection.Close()
}

func (pdk *PDK) Start() {
	nats.StartEventStream(pdk.Nats.JetStream)
	channel.StartChannelStream(&pdk.Chan)
	//sig := make(chan os.Signal, 1)
	//signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	//<-sig
}

func (pdk *PDK) PostEvent(subject string, data proto.Message, sc Scope) { // account_created
	if sc.Nats {
		pdk.Nats.PostEvent(subject, data)
		return
	}
	if sc.Local {
		go pdk.Chan.PostEvent(subject, data)
		return
	}
}

type SubjectHandler[R proto.Message] func(ctx *Context, data R)

func RegisterSubject[R proto.Message](subject string, handler SubjectHandler[R]) {
	//register Nats
	nats.RegisterNats(subject, handler)

	//register chan
	channel.RegisterChan(subject, handler)
}
