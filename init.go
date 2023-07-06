package go_pdk

type AuthenticationConfig struct {
	ManagerUrl string
	Name       string
	Secret     string
}

func Authenticate(config AuthenticationConfig) {
	//We can authenticate server here with a manager
}

func Init(pluginDir string) *PDK {
	Server = NewServer()
	Server.SetPluginDir(pluginDir)

	var LOG *Logger

	Pdk = &PDK{
		LOG: LOG,
		Chan: Chan{
			DataChan: make(map[string]chan Data),
		},
	}

	return Pdk
}
