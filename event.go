package go_pdk

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"log"
	"reflect"
	"strings"
)

type EventHandler[R proto.Message] func(ctx *Context, data R)

type eventStream struct {
	sender    string
	receiver  string
	executors map[string]func(m *nats.Msg, id int64)
}

var eventStreams map[string]*eventStream = make(map[string]*eventStream)

func RegisterEvent[R proto.Message](subject string, handler EventHandler[R]) {
	parts := strings.Split(subject, ".")
	stream := createOrGetEventStream(parts[0])
	log.Println(fmt.Sprintf("Events: subject = %s, receiver = %s", subject, stream.receiver))
	var event R
	ref := reflect.New(reflect.TypeOf(event).Elem())
	event = ref.Interface().(R)

	stream.executors[subject] = func(m *nats.Msg, eventID int64) {
		var msg Event
		if err := proto.Unmarshal(m.Data, &msg); err != nil {
			log.Print("Register unmarshal error response data:", err.Error())
			return
		}
		context := Context{
			Logger{ID: 1},
		}

		if err := proto.Unmarshal(msg.Body, event); err == nil {
			handler(&context, event)
		} else {
			log.Print("Error in parsing data:", err)
		}
	}
}

func (es *eventStream) start() {
	sub, err := JetStream.PullSubscribe("", es.receiver, nats.BindStream(es.sender))

	if err != nil {
		log.Fatal("Error in start event stream - sender ", es.sender, "- receiver ", es.receiver, " : ", err.Error())
	}

	go func() {
		for {
			if messages, err := sub.Fetch(1); err == nil {
				if len(messages) == 1 {
					m := messages[0]
					if _, ok := es.executors[m.Subject]; ok {
						log.Println("Execute successfully")
					}
					m.Ack()
				}
			}
		}
	}()
}

func createOrGetEventStream(sender string) *eventStream {
	if stream, ok := eventStreams[sender]; ok {
		return stream
	}

	stream := &eventStream{
		sender:    sender,
		receiver:  module,
		executors: make(map[string]func(m *nats.Msg, id int64)),
	}

	eventStreams[sender] = stream
	return stream
}

func startEventStream() {
	for _, e := range eventStreams {
		e.start()
	}
}
