package go_pdk

import (
	"google.golang.org/protobuf/proto"
	"log"
)

type Data struct {
	Subject string
	Payload proto.Message
}

type channel struct {
	Subject   string
	executors map[string]func(m proto.Message)
}

var channelStreams map[string]*channel = make(map[string]*channel)

type ChannelHandler[R proto.Message] func(data R)

func RegisterChannelSubject[R proto.Message](subject string, handler ChannelHandler[R]) {
	channelStream := createOrGetChannelStream(subject)

	channelStream.executors[subject] = func(m proto.Message) {
		if data, ok := m.(R); ok {
			handler(data)
		} else {
			log.Printf("Received message of unexpected type: %T", m)
		}
	}
}

func createOrGetChannelStream(subject string) *channel {
	if stream, ok := channelStreams[subject]; ok {
		return stream
	}

	stream := &channel{
		Subject:   subject,
		executors: make(map[string]func(m proto.Message)),
	}

	channelStreams[subject] = stream
	return stream
}

func (pdk *PDK) start(c channel) {
	go func() {
		for {
			select {
			case data := <-pdk.dataChan[c.Subject]:
				if executor, ok := c.executors[data.Subject]; ok {
					executor(data.Payload)
				}
			}
		}
	}()
}

func startChannelStream(pdk *PDK) {
	for _, c := range channelStreams {
		pdk.start(*c)
	}
}
