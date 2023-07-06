package old

import (
	"github.com/reijiokito/go-pdk"
	"google.golang.org/protobuf/proto"
)

type IPostEvent interface {
	PostEvent(channel string, data proto.Message)
}

type IRegisterEvent interface {
	RegisterEvent(subject string, handler go_pdk.SubjectHandler[proto.Message])
}
