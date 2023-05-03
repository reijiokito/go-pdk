package go_pdk

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"log"
	"reflect"
	"strings"
)

type Nats struct {
	Connection *nats.Conn
	JetStream  nats.JetStreamContext
}

type eventStream struct {
	sender    string
	receiver  string
	executors map[string]func(m *nats.Msg)
}

var eventStreams map[string]*eventStream = make(map[string]*eventStream)

func RegisterNats[R proto.Message](subject string, handler SubjectHandler[R]) {
	parts := strings.Split(subject, ".")
	stream := createOrGetEventStream(parts[0])
	log.Println(fmt.Sprintf("Events: subject = %s, receiver = %s", subject, stream.receiver))
	var event R
	ref := reflect.New(reflect.TypeOf(event).Elem())
	event = ref.Interface().(R)

	stream.executors[subject] = func(m *nats.Msg) {
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
			log.Print("Error in parsing data nats:", err)
		}
	}
}

func (n *Nats) PostEvent(subject string, data proto.Message) error { // account_created
	msg := Event{}
	if data, err := proto.Marshal(data); err == nil {
		msg.Body = data
	} else {
		return err
	}

	if data, err := proto.Marshal(&msg); err == nil {
		n.JetStream.Publish(subject, data)
	} else {
		return err
	}
	return nil
}

func (es *eventStream) start(JetStream nats.JetStreamContext) {
	sub, err := JetStream.PullSubscribe("", es.receiver, nats.BindStream(es.sender))

	if err != nil {
		log.Println("Error in start event stream - sender ", es.sender, "- receiver ", es.receiver, " : ", err.Error())
	}

	go func() {
		for {
			if messages, err := sub.Fetch(1); err == nil {
				if len(messages) == 1 {
					m := messages[0]
					if executor, ok := es.executors[m.Subject]; ok {
						executor(m)
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
		receiver:  Module,
		executors: make(map[string]func(m *nats.Msg)),
	}

	eventStreams[sender] = stream
	return stream
}

func startEventStream(js nats.JetStreamContext) {
	for _, e := range eventStreams {
		e.start(js)
	}
}
