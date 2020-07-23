// Package outputs with managing of values of discovered outputs
package outputs

import (
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
)

// DomainOutputValues for managing values of discovered outputs
type DomainOutputValues struct {
	// c             lib.DomainCollection //
	raw           map[string]string
	latest        map[string]*types.OutputLatestMessage
	history       map[string]*types.OutputHistoryMessage
	event         map[string]*types.OutputEventMessage
	messageSigner *messaging.MessageSigner // subscription to output discovery messages
	updateMutex   *sync.Mutex              // mutex for async updating of outputs
}

// GetRaw returns the latest raw value of an output
func (dov *DomainOutputValues) GetRaw(rawAddress string) (value string, found bool) {
	dov.updateMutex.Lock()
	defer dov.updateMutex.Unlock()
	value, found = dov.raw[rawAddress]
	return value, found
}

// GetLatest returns the 'latest' value message of an output
func (dov *DomainOutputValues) GetLatest(latestAddress string) (value *types.OutputLatestMessage, found bool) {
	dov.updateMutex.Lock()
	defer dov.updateMutex.Unlock()
	value, found = dov.latest[latestAddress]
	return value, found
}

// UpdateEvent replaces the node event value
func (dov *DomainOutputValues) UpdateEvent(value *types.OutputEventMessage) {
	dov.updateMutex.Lock()
	defer dov.updateMutex.Unlock()
	dov.event[value.Address] = value
}

// UpdateHistory replaces the output history value
func (dov *DomainOutputValues) UpdateHistory(value *types.OutputHistoryMessage) {
	dov.updateMutex.Lock()
	defer dov.updateMutex.Unlock()
	dov.history[value.Address] = value
}

// UpdateLatest replaces the latest output value by output address
func (dov *DomainOutputValues) UpdateLatest(value *types.OutputLatestMessage) {
	dov.updateMutex.Lock()
	defer dov.updateMutex.Unlock()
	dov.latest[value.Address] = value
}

// UpdateRaw replaces the output raw value
func (dov *DomainOutputValues) UpdateRaw(address string, value string) {
	dov.updateMutex.Lock()
	defer dov.updateMutex.Unlock()
	dov.raw[address] = value
}

// NewDomainOutputValues creates a new instance for handling of discovered output values
func NewDomainOutputValues(messageSigner *messaging.MessageSigner) *DomainOutputValues {
	return &DomainOutputValues{
		// c:             lib.NewDomainCollection(messageSigner, reflect.TypeOf(&types.OutputLatestMessage{})),
		messageSigner: messageSigner,
		updateMutex:   &sync.Mutex{},
		raw:           make(map[string]string, 0),
		latest:        make(map[string]*types.OutputLatestMessage, 0),
		history:       make(map[string]*types.OutputHistoryMessage, 0),
		event:         make(map[string]*types.OutputEventMessage, 0),
	}
}
