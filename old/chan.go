package old

import (
	"github.com/reijiokito/go-pdk"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Chan struct {
	DataChan map[string]chan Data
}

type Data struct {
	Subject string
	Payload proto.Message
}

type handler struct {
	Subject   string
	executors map[string]func(m proto.Message)
}

var channelStreams map[string]*handler = make(map[string]*handler)

func RegisterChan[R proto.Message](subject string, handler go_pdk.SubjectHandler[R]) {
	channelStream := createOrGetChannelStream(subject)

	channelStream.executors[subject] = func(m proto.Message) {
		context := go_pdk.Context{
			go_pdk.Logger{ID: 1},
		}

		if data, ok := m.(R); ok {
			handler(&context, data)
		} else {
			log.Print("Error in parsing data chan:", m)
		}
	}
}

func (ch *Chan) PostEvent(subject string, payload proto.Message) error {
	if _, ok := ch.DataChan[subject]; !ok {
		dataChannel := make(chan Data)
		ch.DataChan[subject] = dataChannel
	}

	data := Data{
		Subject: subject,
		Payload: payload,
	}
	ch.DataChan[subject] <- data
	return nil
}

func createOrGetChannelStream(subject string) *handler {
	if stream, ok := channelStreams[subject]; ok {
		return stream
	}

	stream := &handler{
		Subject:   subject,
		executors: make(map[string]func(m proto.Message)),
	}

	channelStreams[subject] = stream
	return stream
}

func (ch *Chan) start(c handler) {
	go func() {
		for {
			select {
			case data := <-ch.DataChan[c.Subject]:
				if executor, ok := c.executors[data.Subject]; ok {
					executor(data.Payload)
				}
			}
		}
	}()
}

func StartChannelStream(ch *Chan) {
	for _, c := range channelStreams {
		ch.start(*c)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			select {
			case <-sig:
				for _, d := range ch.DataChan {
					close(d)
				}
			}
		}
	}()

}
