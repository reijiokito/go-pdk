package go_pdk

import (
	"fmt"
	"github.com/nats-io/nats.go"
	"log"
	"strings"
)

type AuthenticationConfig struct {
	ManagerUrl string
	Name       string
	Secret     string
}

func Authenticate(config AuthenticationConfig) {
	//We can authenticate plugin here with a manager
}

func Init(name string, c *Configuration) *PDK {
	Module = name

	var Connection *nats.Conn
	var LOG *Logger

	var err error
	var nats_ []string
	for _, n := range strings.Split(c.NatsUrl, ",") {
		fmt.Printf("Nats configuration: nats://%s:4222\n", n)
		nats_ = append(nats_, fmt.Sprintf("nats://%s:4222", n))
	}

	if c.NatsUsername != "" && c.NatsPassword != "" {
		Connection, err = nats.Connect(strings.Join(nats_, ","), nats.UserInfo(c.NatsUsername, c.NatsPassword))
	} else {
		Connection, err = nats.Connect(strings.Join(nats_, ","))
	}

	if err != nil {
		log.Println("Can not connect to NATS:", nats_)
	}

	/*init jetstream*/
	JetStream, err := Connection.JetStream()
	if err != nil {
		log.Println(err)
	}

	return &PDK{
		LOG: LOG,
		Nats: Nats{
			Connection: Connection,
			JetStream:  JetStream,
		},
		Chan: Chan{
			dataChan: make(map[string]chan Data),
		},
	}
}
