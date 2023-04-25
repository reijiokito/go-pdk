package go_pdk

import "fmt"

type Data struct {
	FromPlugin string
	Subject    string
	Payload    interface{}
}

func (pdk *PDK) SendData(fromPlugin string, subject string, payload interface{}) error {
	// create a new Data object and send it to the channel
	data := Data{
		FromPlugin: fromPlugin,
		Subject:    subject,
		Payload:    payload,
	}
	pdk.dataChan <- data
	return nil
}

func (pdk *PDK) StartChan() error {
	// start a goroutine to handle incoming data
	go func() {
		for {
			select {
			case data := <-pdk.dataChan:
				// handle the incoming data
				if data.Subject == "pluginA" {
					// handle data from pluginA
					str, ok := data.Payload.(string)
					if !ok {
						fmt.Println("invalid data type, expected string")
					} else {
						fmt.Printf("PluginA received data: %s\n", str)
					}
				} else if data.Subject == "pluginB" {
					// handle data from pluginB
					num, ok := data.Payload.(int)
					if !ok {
						fmt.Println("invalid data type, expected int")
					} else {
						fmt.Printf("PluginB received data: %d\n", num)
					}
				} else {
					fmt.Println("unknown subject:", data.Subject)
				}
			}
		}
	}()

	return nil
}
