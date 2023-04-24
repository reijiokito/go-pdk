package go_pdk

import "google.golang.org/protobuf/proto"

type Context struct {
	Logger
}

func (pdk *PDK) PostEvent(channel string, data proto.Message) { // account_created
	subject := pdk.module + "." + channel
	msg := Event{}
	if data, err := proto.Marshal(data); err == nil {
		msg.Body = data
	}

	if data, err := proto.Marshal(&msg); err == nil {
		pdk.JetStream.Publish(subject, data)
	}
}
