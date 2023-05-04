package go_pdk

import (
	"github.com/reijiokito/go-pdk/server"
)

type AuthenticationConfig struct {
	ManagerUrl string
	Name       string
	Secret     string
}

func Authenticate(config AuthenticationConfig) {
	//We can authenticate server here with a manager
}

func Init(pluginDir string) *PDK {
	//Module = name

	Server = server.NewServer()
	Server.SetPluginDir(pluginDir)

	//var Connection *nats.Conn
	var LOG *Logger
	//
	//var err error
	//var nats_ []string
	//for _, n := range strings.Split(c.NatsUrl, ",") {
	//	fmt.Printf("Nats configuration: nats://%s:4222\n", n)
	//	nats_ = append(nats_, fmt.Sprintf("nats://%s:4222", n))
	//}
	//
	//if c.NatsUsername != "" && c.NatsPassword != "" {
	//	Connection, err = nats.Connect(strings.Join(nats_, ","), nats.UserInfo(c.NatsUsername, c.NatsPassword))
	//} else {
	//	Connection, err = nats.Connect(strings.Join(nats_, ","))
	//}
	//
	//if err != nil {
	//	log.Println("Can not connect to NATS:", nats_)
	//}
	//
	///*init jetstream*/
	//JetStream, err := Connection.JetStream()
	//if err != nil {
	//	log.Println(err)
	//}

	return &PDK{
		LOG: LOG,
		//Nats: nats2.Nats{
		//	Connection: Connection,
		//	JetStream:  JetStream,
		//},
		Chan: Chan{
			DataChan: make(map[string]chan Data),
		},
	}
}
