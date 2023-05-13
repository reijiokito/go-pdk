package go_pdk

import (
	"fmt"
	"time"
)

type StartEventData struct {
	InstanceId int    // Instance ID to start the event
	EventName  string // event name (not handler method name)
	// ....
}

type EventData struct {
	Id       int              // event id
	Instance *InstanceData    // plugin instance
	Ipc      chan interface{} // communication channel (TODO: use decoded structs)
	Pdk      *PDK             // go-pdk instance
}

func (s *PluginServer) HandleEvent(in StartEventData) (*StepData, error) {
	s.lock.RLock()
	instance, ok := s.Instances[in.InstanceId]
	s.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no plugin instance %d", in.InstanceId)
	}

	h, ok := instance.Handlers[in.EventName]
	if !ok {
		return nil, fmt.Errorf("undefined method %s on plugin %s",
			in.EventName, instance.Plugin.Name)
	}

	ipc := make(chan interface{})

	event := EventData{
		Instance: instance,
		Ipc:      ipc,
		//Pdk:      dk.Init(ipc),
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

		func() {
			defer func() { recover() }()
			ipc <- "ret"
		}()

		s.lock.Lock()
		defer s.lock.Unlock()
		event.Instance.lastEvent = time.Now()
		delete(s.Events, event.Id)
	}()

	ipc <- "run" // kickstart the handler

	out := &StepData{EventId: event.Id, Data: <-ipc}
	return out, nil
}

// A callback's response/request.
type StepData struct {
	EventId int         // event cycle to which this belongs
	Data    interface{} // carried data
}

// Step carries a callback's answer back from Sigma to the plugin,
// the return value is either a new callback request or a finish signal.
//
// RPC exported method
func (s *PluginServer) Step(in StepData) (*StepData, error) {
	s.lock.RLock()
	event, ok := s.Events[in.EventId]
	s.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("No running event %d", in.EventId)
	}

	event.Ipc <- in.Data
	outStr := <-event.Ipc
	out := &StepData{EventId: in.EventId, Data: outStr}

	return out, nil
}
