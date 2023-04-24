package go_pdk

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"log"
)

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
