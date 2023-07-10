package log

import (
	"github.com/reijiokito/go-pdk/bridge"
)

// Holds this module's functions.  Accessible as `sigma.Log`
type Log struct {
	bridge.PdkBridge
}

func New(ch chan interface{}) Log {
	return Log{bridge.New(ch)}
}

func (r Log) Err(args ...interface{}) error {
	_, err := r.Ask(`sigma.log.err`, args...)
	return err
}

func (r Log) Warn(args ...interface{}) error {
	_, err := r.Ask(`sigma.log.warn`, args...)
	return err
}

func (r Log) Info(args ...interface{}) error {
	_, err := r.Ask(`sigma.log.info`, args...)
	return err
}
