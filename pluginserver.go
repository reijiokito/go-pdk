package go_pdk

import (
	"fmt"
	"os"
	"path"
	"plugin"
	"reflect"
	"strings"
	"sync"
	"time"
)

// --- PluginServer --- //

type PluginServer struct {
	lock           sync.RWMutex
	PluginsDir     string
	Plugins        map[string]*PluginData
	Instances      map[int]*InstanceData
	Events         map[int]*EventData
	nextInstanceId int
	nextEventId    int
}

// Create a new server context.
func NewServer() *PluginServer {
	s := PluginServer{
		Plugins:   map[string]*PluginData{},
		Instances: map[int]*InstanceData{},
	}

	return &s
}

// SetPluginDir tells the server where to find the Plugins.

func (s *PluginServer) SetPluginDir(dir string) string {
	s.lock.Lock()
	s.PluginsDir = dir
	s.lock.Unlock()
	return "ok"
}

// --- status --- //

type ServerStatusData struct {
	Pid     int
	Plugins map[string]PluginStatusData
}

func (s *PluginServer) GetStatus(n int) (*ServerStatusData, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	reply := &ServerStatusData{
		Pid:     os.Getpid(),
		Plugins: make(map[string]PluginStatusData),
	}

	var err error
	for pluginname := range s.Plugins {
		reply.Plugins[pluginname], err = s.getPluginStatus(pluginname)
		if err != nil {
			return nil, err
		}
	}

	return reply, nil
}

// --- PluginData  --- //

type PluginData struct {
	lock              sync.Mutex
	Name              string
	Code              *plugin.Plugin
	Modtime           time.Time
	Loadtime          time.Time
	Constructor       func() interface{}
	Config            interface{}
	LastStartInstance time.Time
	LastCloseInstance time.Time
	Services          map[string]func(...interface{})
	Callers           map[string]func(...interface{}) interface{}
}

func getModTime(fname string) (modtime time.Time, err error) {
	finfo, err := os.Stat(fname)
	if err != nil {
		return
	}

	modtime = finfo.ModTime()
	return
}

func (s *PluginServer) loadPlugin(name string) (plug *PluginData, err error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	plug, ok := s.Plugins[name]
	if ok {
		return
	}

	plugFName := path.Join(s.PluginsDir, name+".so")
	plugModTime, err := getModTime(plugFName)
	if err != nil {
		return
	}

	code, err := plugin.Open(plugFName)
	if err != nil {
		err = fmt.Errorf("failed to open server %s: %w", name, err)
		return
	}

	constructorSymbol, err := code.Lookup("New")
	if err != nil {
		err = fmt.Errorf("No constructor function on server %s: %w", name, err)
		return
	}

	constructor, ok := constructorSymbol.(func() interface{})
	if !ok {
		err = fmt.Errorf("Wrong constructor signature on server %s: %w", name, err)
		return
	}

	getServicesSymbol, err := code.Lookup("GetServices")
	if err != nil {
		err = fmt.Errorf("No constructor function on server %s: %w", name, err)
		return
	}

	getServices, ok := getServicesSymbol.(func() map[string]func(...interface{}))
	if !ok {
		err = fmt.Errorf("Wrong constructor signature on server %s: %w", name, err)
		return
	}

	getCallersSymbol, err := code.Lookup("GetCallers")
	if err != nil {
		err = fmt.Errorf("No constructor function on server %s: %w", name, err)
		return
	}

	getCallers, ok := getCallersSymbol.(func() map[string]func(...interface{}) interface{})
	if !ok {
		err = fmt.Errorf("Wrong constructor signature on server %s: %w", name, err)
		return
	}

	plug = &PluginData{
		Name:        name,
		Code:        code,
		Modtime:     plugModTime,
		Loadtime:    time.Now(),
		Constructor: constructor,
		Config:      constructor(),
		Services:    getServices(),
		Callers:     getCallers(),
	}

	s.Plugins[name] = plug

	return
}

type schemaDict map[string]interface{}

func getSchemaDict(t reflect.Type) schemaDict {
	switch t.Kind() {
	case reflect.String:
		return schemaDict{"type": "string"}

	case reflect.Bool:
		return schemaDict{"type": "boolean"}

	case reflect.Int, reflect.Int32:
		return schemaDict{"type": "integer"}

	case reflect.Uint, reflect.Uint32:
		return schemaDict{
			"type":    "integer",
			"between": []int{0, 2147483648},
		}

	case reflect.Float32, reflect.Float64:
		return schemaDict{"type": "number"}

	case reflect.Slice:
		elemType := getSchemaDict(t.Elem())
		if elemType == nil {
			break
		}
		return schemaDict{
			"type":     "array",
			"elements": elemType,
		}

	case reflect.Map:
		kType := getSchemaDict(t.Key())
		vType := getSchemaDict(t.Elem())
		if kType == nil || vType == nil {
			break
		}
		return schemaDict{
			"type":   "map",
			"keys":   kType,
			"values": vType,
		}

	case reflect.Struct:
		fieldsArray := []schemaDict{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			typeDecl := getSchemaDict(field.Type)
			if typeDecl == nil {
				// ignore unrepresentable types
				continue
			}
			name := field.Tag.Get("json")
			if name == "" {
				name = strings.ToLower(field.Name)
			}
			fieldsArray = append(fieldsArray, schemaDict{name: typeDecl})
		}
		return schemaDict{
			"type":   "record",
			"fields": fieldsArray,
		}
	}

	return nil
}

// Information obtained from a server's compiled Code.
type PluginInfo struct {
	Name     string     // server Name
	ModTime  time.Time  `codec:",omitempty"` // server file modification time
	LoadTime time.Time  `codec:",omitempty"` // server load time
	Phases   []string   // events it can handle
	Version  string     // version number
	Priority int        // priority info
	Schema   schemaDict // JSON representation of the Config schema
}

// GetPluginInfo loads and retrieves information from the compiled server.
// TODO: reload if the server Code has been updated.

func (s *PluginServer) GetPluginInfo(name string) (*PluginInfo, error) {
	plug, err := s.loadPlugin(name)
	if err != nil {
		return nil, err
	}

	info := &PluginInfo{Name: name, LoadTime: plug.Loadtime, ModTime: plug.Modtime}

	plug.lock.Lock()
	defer plug.lock.Unlock()
	handlers := getHandlers(plug.Config)

	info.Phases = make([]string, len(handlers))
	var i = 0
	for name := range handlers {
		info.Phases[i] = name
		i++
	}

	v, _ := plug.Code.Lookup("Version")
	if v != nil {
		info.Version = *v.(*string)
	}

	prio, _ := plug.Code.Lookup("Priority")
	if prio != nil {
		info.Priority = *prio.(*int)
	}

	// 	st, _ := getSchemaDict(reflect.TypeOf(plug.Config).Elem())
	info.Schema = schemaDict{
		"Name": name,
		"fields": []schemaDict{
			schemaDict{"Config": getSchemaDict(reflect.TypeOf(plug.Config).Elem())},
		},
	}

	return info, nil
}

type PluginStatusData struct {
	Name              string
	Modtime           int64
	LoadTime          int64
	Instances         []InstanceStatus
	LastStartInstance int64
	LastCloseInstance int64
}

func (s *PluginServer) getPluginStatus(name string) (status PluginStatusData, err error) {
	plug, ok := s.Plugins[name]
	if !ok {
		err = fmt.Errorf("server %#v not loaded", name)
		return
	}

	instances := []InstanceStatus{}
	for _, instance := range s.Instances {
		if instance.Plugin == plug {
			instances = append(instances, InstanceStatus{
				Name:      name,
				Id:        instance.Id,
				Config:    instance.Config,
				StartTime: instance.StartTime.Unix(),
			})
		}
	}

	status = PluginStatusData{
		Name:              name,
		Modtime:           plug.Modtime.Unix(),
		LoadTime:          plug.Loadtime.Unix(),
		Instances:         instances,
		LastStartInstance: plug.LastStartInstance.Unix(),
		LastCloseInstance: plug.LastCloseInstance.Unix(),
	}
	return
}
