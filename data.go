package go_pdk

import (
	"google.golang.org/protobuf/proto"
	"reflect"
)

type Data struct {
	Subject string
	Payload proto.Message
}

type ChannelHandler[R proto.Message] func(data R)

type channel struct {
	Subject   string
	executors map[string]func(m proto.Message)
}

var channelStreams map[string]*channel = make(map[string]*channel)

func RegisterChannelSubject[R proto.Message](subject string, handler ChannelHandler[R]) {
	channelStream := createOrGetChannelStream(subject)

	var event R
	ref := reflect.New(reflect.TypeOf(event).Elem())
	event = ref.Interface().(R)

	channelStream.executors[subject] = func(m proto.Message) {
		handler(event)
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
