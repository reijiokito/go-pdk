package go_pdk

import (
	"fmt"
	"log"
	"time"
)

type StartEventData struct {
	InstanceId int    // Instance ID to start the event
	EventName  string // event name (not handler method name)
	// ....
}

type EventData struct {
	Id       int           // event id
	Instance *InstanceData // plugin instance
	Ipc      chan any      // communication channel (TODO: use decoded structs)
	Pdk      *PDK          // go-pdk instance
}

func (s *PluginServer) HandleEvent(in StartEventData) error {
	s.lock.RLock()
	instance, ok := s.Instances[in.InstanceId]
	s.lock.RUnlock()
	if !ok {
		return fmt.Errorf("no plugin instance %d", in.InstanceId)
	}

	h, ok := instance.Handlers[in.EventName]
	if !ok {
		return fmt.Errorf("undefined method %s on plugin %s",
			in.EventName, instance.Plugin.Name)
	}

	ipc := make(chan any)

	event := EventData{
		Instance: instance,
		Ipc:      ipc,
		Pdk:      Init(ipc),
	}

	s.lock.Lock()
	event.Id = s.nextEventId
	s.nextEventId++
	s.Events[event.Id] = &event
	s.lock.Unlock()

	//log.Printf("Will launch goroutine for key %d / operation %s\n", key, op)
	go func() {
		_ = <-ipc
		h(event.Pdk)
		log.Println("Done run")

		s.lock.Lock()
		defer s.lock.Unlock()
		event.Instance.lastEvent = time.Now()
		log.Println("Done Update Event")
		log.Println("LENTH: ", len(s.Events))
		delete(s.Events, event.Id)
		log.Println("LENTH: ", len(s.Events))
		log.Println("Done Delete")
	}()

	ipc <- "run" // kickstart the handler

	return nil
}
