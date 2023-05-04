package go_pdk

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"log"
)

type ServiceHandler[R proto.Message] func(ctx *Service, request R)

func RegisterService[R proto.Message](conn *nats.Conn, url string, handler ServiceHandler[R]) {
	log.Println("Register Service: ", url)
	var request R

	ctx := Service{

		Context:    Context{},
		Connection: conn,
	}

	_, err := conn.QueueSubscribe(SubscriberURL(url), "API", func(m *nats.Msg) {
		if err := proto.Unmarshal(m.Data, &ctx.Request); err != nil {
			log.Print("Register unmarshal error response data:", err.Error())
			return
		}

		ctx.Reply = m.Reply
		ref := request.ProtoReflect().New()
		request = ref.Interface().(R)

		if ctx.Request.JSON {
			if err := json.Unmarshal(ctx.Request.Body, request); err != nil {
				log.Print("Bad Request: " + err.Error())
				ctx.Error(&Error{Code: 2, Message: "Bad Request"})
			} else {
				handler(&ctx, request)
			}

		} else {
			if err := proto.Unmarshal(ctx.Request.Body, request); err != nil {
				log.Print("Bad Request: " + err.Error())
				ctx.Error(&Error{Code: 2, Message: "Bad Request"})
			} else {
				handler(&ctx, request)
			}
		}
	})

	if err != nil {
		log.Fatal("Can not register service:", url)
	}
}

var jsonMarshaller = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

type Service struct {
	Context
	Request    Request
	Reply      string
	Connection *nats.Conn
}

func (ctx *Service) Error(e *Error) {
	response := Response{Code: 400}

	var err error
	if ctx.Request.JSON {
		response.Body, err = json.Marshal(e)
	}

	if err != nil {
		response.Code = int32(500)
		response.Body = []byte(err.Error())
	}
	ctx.flush(&response)
}

func (ctx *Service) Done(r proto.Message) {
	response := Response{Code: 200}

	var err error
	if ctx.Request.JSON {
		response.Body, err = jsonMarshaller.Marshal(r)
	} else {
		response.Body, err = proto.Marshal(r)
	}
	if err != nil {
		response.Code = int32(500)
		response.Body = []byte(err.Error())
	}
	ctx.flush(&response)
}

func (ctx *Service) flush(response *Response) {
	bytes, err := proto.Marshal(response)
	if err != nil {
		log.Print("Register marshal error response data:", err.Error())
		return
	}
	err = ctx.Connection.Publish(ctx.Reply, bytes)
	if err != nil {
		log.Println(fmt.Sprintf("Nats publish to [%s] error: %s", ctx.Reply, err.Error()))
	}
}
