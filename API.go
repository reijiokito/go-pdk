package go_pdk

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
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
