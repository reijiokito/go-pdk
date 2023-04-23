package go_pdk

import "google.golang.org/protobuf/proto"

type Context struct {
	Logger
}

func PostEvent(channel string, data proto.Message) { // account_created
	subject := module + "." + channel
	msg := Event{}
	if data, err := proto.Marshal(data); err == nil {
		msg.Body = data
	}

	if data, err := proto.Marshal(&msg); err == nil {
		JetStream.Publish(subject, data)
	}
}
