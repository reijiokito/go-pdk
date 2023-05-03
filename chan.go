package go_pdk

import (
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type Chan struct {
	dataChan map[string]chan Data
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

func RegisterChan[R proto.Message](subject string, handler SubjectHandler[R]) {
	channelStream := createOrGetChannelStream(subject)

	channelStream.executors[subject] = func(m proto.Message) {
		context := Context{
			Logger{ID: 1},
		}

		if data, ok := m.(R); ok {
			handler(&context, data)
		} else {
			log.Print("Error in parsing data chan:", m)
		}
	}
}

func (ch *Chan) PostEvent(subject string, payload proto.Message) error {
	if _, ok := ch.dataChan[subject]; !ok {
		dataChannel := make(chan Data)
		ch.dataChan[subject] = dataChannel
	}

	data := Data{
		Subject: subject,
		Payload: payload,
	}
	ch.dataChan[subject] <- data
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
			case data := <-ch.dataChan[c.Subject]:
				if executor, ok := c.executors[data.Subject]; ok {
					executor(data.Payload)
				}
			}
		}
	}()
}

func startChannelStream(ch *Chan) {
	for _, c := range channelStreams {
		ch.start(*c)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case <-sig:
			for _, d := range ch.dataChan {
				close(d)
			}
		}
	}

}
