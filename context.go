package go_pdk

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"log"
	"reflect"
	"strings"
)

type Context struct {
	Logger
}

func RegisterSubject[R proto.Message](subject string, handler SubjectHandler[R]) {
	//register Nats
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
			log.Print("Error in parsing data:", err)
		}
	}

	//register chan
	channelStream := createOrGetChannelStream(subject)

	channelStream.executors[subject] = func(m proto.Message) {
		context := Context{
			Logger{ID: 1},
		}

		if data, ok := m.(R); ok {
			handler(&context, data)
		} else {
			log.Printf("Received message of unexpected type: %T", m)
		}
	}
}
