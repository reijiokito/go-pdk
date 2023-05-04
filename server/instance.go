package server

import (
	"encoding/json"
	"fmt"
	"github.com/reijiokito/go-pdk"
	"log"
	"time"
)

// --- instanceData --- //
type instanceData struct {
	id          int
	plugin      *pluginData
	startTime   time.Time
	initialized bool
	config      interface{}
	handlers    map[string]func(pdk *go_pdk.PDK)
	lastEvent   time.Time
}

type (
	accesser interface{ Access(*go_pdk.PDK) }
)

func getHandlers(config interface{}) map[string]func(kong *go_pdk.PDK) {
	handlers := map[string]func(kong *go_pdk.PDK){}

	if h, ok := config.(accesser); ok {
		handlers["access"] = h.Access
	}

	return handlers
}

func (s *PluginServer) expireInstances() error {
	const instanceTimeout = 60
	expirationCutoff := time.Now().Add(time.Second * -instanceTimeout)

	oldinstances := map[int]bool{}
	for id, inst := range s.instances {
		if inst.startTime.Before(expirationCutoff) && inst.lastEvent.Before(expirationCutoff) {
			oldinstances[id] = true
		}
	}

	for id := range oldinstances {
		inst := s.instances[id]
		log.Printf("closing instance %#v:%v", inst.plugin.name, inst.id)
		delete(s.instances, id)
	}

	return nil
}

// Configuration data for a new server instance.
type PluginConfig struct {
	Name   string // server name
	Config []byte // configuration data, as a JSON string
}

// Current state of a server instance.  TODO: add some statistics
type InstanceStatus struct {
	Name      string      // server name
	Id        int         // instance id
	Config    interface{} // configuration data, decoded
	StartTime int64
}

// StartInstance starts a server instance, as requred by configuration data.  More than
// one instance can be started for a single server.  If the configuration changes,
// a new instance should be started and the old one closed.
func (s *PluginServer) StartInstance(config PluginConfig) (*InstanceStatus, error) {
	plug, err := s.loadPlugin(config.Name)
	if err != nil {
		return nil, err
	}

	plug.lock.Lock()
	defer plug.lock.Unlock()

	instanceConfig := plug.constructor()

	if err := json.Unmarshal(config.Config, instanceConfig); err != nil {
		return nil, fmt.Errorf("Decoding config: %w", err)
	}

	instance := instanceData{
		plugin:    plug,
		startTime: time.Now(),
		config:    instanceConfig,
		handlers:  getHandlers(instanceConfig),
	}

	s.lock.Lock()
	instance.id = s.nextInstanceId
	s.nextInstanceId++
	s.instances[instance.id] = &instance

	plug.lastStartInstance = instance.startTime

	s.lock.Unlock()

	status := &InstanceStatus{
		Name:      config.Name,
		Id:        instance.id,
		Config:    instance.config,
		StartTime: instance.startTime.Unix(),
	}

	log.Printf("Started instance %#v:%v", config.Name, instance.id)

	return status, nil
}

// InstanceStatus returns a given resource's status (the same given when started)
func (s *PluginServer) InstanceStatus(id int) (*InstanceStatus, error) {
	s.lock.RLock()
	instance, ok := s.instances[id]
	s.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("No server instance %d", id)
	}

	status := &InstanceStatus{
		Name:   instance.plugin.name,
		Id:     instance.id,
		Config: instance.config,
	}

	return status, nil
}

// CloseInstance is used when an instance shouldn't be used anymore.
// Doesn't kill any running event but the instance is no longer accesible,
// so it's not possible to start a new event with it and will be garbage
// collected after the last reference event finishes.
// Returns the status just before closing.
func (s *PluginServer) CloseInstance(id int) (*InstanceStatus, error) {
	s.lock.RLock()
	instance, ok := s.instances[id]
	s.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("No server instance %d", id)
	}

	status := &InstanceStatus{
		Name:   instance.plugin.name,
		Id:     instance.id,
		Config: instance.config,
	}

	// kill?

	log.Printf("closed instance %#v:%v", instance.plugin.name, instance.id)

	s.lock.Lock()
	instance.plugin.lastCloseInstance = time.Now()
	delete(s.instances, id)
	s.expireInstances()
	s.lock.Unlock()

	return status, nil
}
