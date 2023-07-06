package go_pdk

import (
	"google.golang.org/protobuf/proto"
)

var Module string
var Server *PluginServer

type Context struct {
	Logger
}

type PDK struct {
	LOG  *Logger
	Chan Chan
}

func (pdk *PDK) Release() {

}

func (pdk *PDK) Start() {
	StartChannelStream(&pdk.Chan)
	//sig := make(chan os.Signal, 1)
	//signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	//<-sig
}

func (pdk *PDK) PostEvent(subject string, data proto.Message) { // account_create
	go func() {
		err := pdk.Chan.PostEvent(subject, data)
		if err != nil {
			return
		}
	}()
	return
}

type SubjectHandler[R proto.Message] func(ctx *Context, data R)

func RegisterSubject[R proto.Message](subject string, handler SubjectHandler[R]) {
	RegisterChan(subject, handler)
}
