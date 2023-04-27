package go_pdk

import "google.golang.org/protobuf/proto"

type IPostEvent interface {
	PostEvent(channel string, data proto.Message)
}

type IRegisterEvent interface {
	RegisterEvent(subject string, handler SubjectHandler[proto.Message])
}
