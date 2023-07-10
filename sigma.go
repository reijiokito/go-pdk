package go_pdk

import (
	"github.com/reijiokito/go-pdk/log"
	"github.com/reijiokito/go-pdk/old"
	"google.golang.org/protobuf/proto"
)

var Module string
var Server *PluginServer

type Context struct {
	log.Log
}

type PDK struct {
	LOG log.Log
}

//func (pdk *PDK) Release() {
//
//}
//
//func (pdk *PDK) Start() {
//	StartChannelStream(&pdk.Chan)
//}
//
//func (pdk *PDK) PostEvent(subject string, data proto.Message) { // account_create
//	go func() {
//		err := pdk.Chan.PostEvent(subject, data)
//		if err != nil {
//			return
//		}
//	}()
//	return
//}

type SubjectHandler[R proto.Message] func(ctx *Context, data R)

func RegisterSubject[R proto.Message](subject string, handler SubjectHandler[R]) {
	old.RegisterChan(subject, handler)
}
