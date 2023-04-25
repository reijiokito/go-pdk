package go_pdk

import (
	"google.golang.org/protobuf/proto"
	"log"
	"reflect"
)

type Data struct {
	Subject string
	Payload []byte
}

type ChannelHandler[R proto.Message] func(data R)

type channel struct {
	Subject   string
	executors map[string]func(m []byte)
}

var channelStreams map[string]*channel = make(map[string]*channel)

func (pdk *PDK) SendChannelData(subject string, payload []byte) error {
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

func RegisterChannelSubject[R proto.Message](subject string, handler ChannelHandler[R]) {
	channelStream := createOrGetChannelStream(subject)

	var event R
	ref := reflect.New(reflect.TypeOf(event).Elem())
	event = ref.Interface().(R)

	channelStream.executors[subject] = func(m []byte) {
		if err := proto.Unmarshal(m, event); err == nil {
			handler(event)
		} else {
			log.Print("Error in parsing data:", err)
		}
	}
}

func createOrGetChannelStream(subject string) *channel {
	if stream, ok := channelStreams[subject]; ok {
		return stream
	}

	stream := &channel{
		Subject:   subject,
		executors: make(map[string]func(m []byte)),
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
