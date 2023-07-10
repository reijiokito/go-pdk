package go_pdk

import "github.com/reijiokito/go-pdk/log"

type AuthenticationConfig struct {
	ManagerUrl string
	Name       string
	Secret     string
}

func Authenticate(config AuthenticationConfig) {
	//We can authenticate server here with a manager
}

func InitServer(pluginDir string) {
	Server = NewServer()
	Server.SetPluginDir(pluginDir)
}

func Init(ch chan any) *PDK {
	return &PDK{
		LOG: log.New(ch),
	}
}
