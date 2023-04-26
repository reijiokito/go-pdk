package go_pdk

import (
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"log"
)

type Nats struct {
	Connection *nats.Conn
	JetStream  nats.JetStreamContext
}

type SubjectHandler[R proto.Message] func(ctx *Context, data R)

type eventStream struct {
	sender    string
	receiver  string
	executors map[string]func(m *nats.Msg)
}

var eventStreams map[string]*eventStream = make(map[string]*eventStream)

func (nats *Nats) PostEvent(subject string, data proto.Message) { // account_created
	msg := Event{}
	if data, err := proto.Marshal(data); err == nil {
		msg.Body = data
	}

	if data, err := proto.Marshal(&msg); err == nil {
		nats.JetStream.Publish(subject, data)
	}
}

func (es *eventStream) start(JetStream nats.JetStreamContext) {
	sub, err := JetStream.PullSubscribe("", es.receiver, nats.BindStream(es.sender))

	if err != nil {
		log.Fatal("Error in start event stream - sender ", es.sender, "- receiver ", es.receiver, " : ", err.Error())
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
		receiver:  "manager",
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
