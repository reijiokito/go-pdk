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

func (pdk *PDK) SendChannelData(subject string, payload proto.Message) error {
	// create a new Data object and send it to the channel
	if _, ok := pdk.dataChan[subject]; !ok {
		dataChannel := make(chan Data)
		pdk.dataChan[subject] = dataChannel
	}

	data := Data{
		Subject: subject,
		Payload: payload,
	}
	pdk.dataChan[subject] <- data
	return nil
}
