// Package lib with generic collection of discovered domain objects such as inputs, outputs and nodes
package lib

import (
	"crypto/ecdsa"
	"reflect"
	"strings"
	"sync"

	"github.com/iotdomain/iotdomain-go/messaging"
)

// DomainCollection for managing discovered nodes,inputs and outputs
// Hopefully this can be replaced with generics soon
type DomainCollection struct {
	DiscoMap map[string]interface{} // discovered by addres
	// MessageSigner *messaging.MessageSigner // subscription to discovery messages
	GetPublicKey func(string) *ecdsa.PublicKey // get the public key for signature verification
	UpdateMutex  *sync.Mutex                   // mutex for async updating
	ItemPtr      reflect.Type                  // pointer type of item in map
}

// Add adds or replaces the discovered object
// objectPtr must be a pointer to the instance
func (dc *DomainCollection) Add(address string, objectPtr interface{}) {
	dc.UpdateMutex.Lock()
	defer dc.UpdateMutex.Unlock()
	base := MakeBaseAddress(address)
	dc.DiscoMap[base] = objectPtr
}

// Get returns an item by node address and optionally ioType and instance
// This constructs the address from the nodeAddress, ioType and instance. If the ioType and
// instance are empty, the lookup is done using the base of the node address (without the $node suffix)
func (dc *DomainCollection) Get(nodeAddress string, ioType string, instance string) interface{} {
	base := MakeBaseAddress(nodeAddress)
	if ioType != "" {
		base = base + "/" + ioType + "/" + instance
	}

	dc.UpdateMutex.Lock()
	defer dc.UpdateMutex.Unlock()
	item := dc.DiscoMap[base]
	return item
}

// GetByAddress returns an item by its domain address.
// This ignores any trailing messageType and uses only the base addres of node, input or output
func (dc *DomainCollection) GetByAddress(address string) interface{} {
	base := MakeBaseAddress(address)

	dc.UpdateMutex.Lock()
	defer dc.UpdateMutex.Unlock()
	item := dc.DiscoMap[base]
	return item
}

// GetByAddressPrefix fills the given slice (pointer) with all items that start with the given address
// The message type is removed from addressPrefix, so a node discover address can be used to
//  find corresponding inputs and outputs.
// The result is stored in resultSlicePtr which must be a pointer to a slice
//   that contains pointers to items, eg: []*Item
func (dc *DomainCollection) GetByAddressPrefix(addressPrefix string, resultSlicePtr interface{}) {

	// todo: check that resultSlicePtr is of the right type

	dc.UpdateMutex.Lock()
	defer dc.UpdateMutex.Unlock()
	base := MakeBaseAddress(addressPrefix)
	itemListVal := reflect.ValueOf(resultSlicePtr).Elem()

	for addr, item := range dc.DiscoMap {
		if strings.HasPrefix(addr, base) {
			objectValue := reflect.ValueOf(item)
			itemListVal.Set(reflect.Append(itemListVal, objectValue))

		}
	}
}

// GetAll populates the given list with the objects in this collection
// resultSlicePtr is a pointer to a slice to store the result. This reduces the magic
// somewhat.
// This magic goo is brewed by a genius named Martin Tournoij:
//  https://stackoverflow.com/questions/37939388/how-can-i-add-elements-to-slice-reflection
func (dc *DomainCollection) GetAll(resultSlicePtr interface{}) {
	dc.UpdateMutex.Lock()
	defer dc.UpdateMutex.Unlock()

	// Oh the magic! create a slice instance of the item type
	// models := reflect.New(reflect.SliceOf(dc.ItemPtr)).Interface()
	// all := reflect.ValueOf(models).Elem()

	// update - to prevent a nil result when the map is empty, pass in the slice instead
	all := reflect.ValueOf(resultSlicePtr).Elem()

	for _, object := range dc.DiscoMap {
		// Append the object to the slice with ... magic!
		objectValue := reflect.ValueOf(object)
		all.Set(reflect.Append(all, objectValue))
	}
}

// HandleDiscovery updates the collection with a discovered item
// This verifies that the discovery message is properly signed by its publisher,
// unmarshals the message into newItem and adds it to the collection. newItem must
// be a pointer to an object of the proper type.
//  For convenience this also set the PublisherID, NodeID in the target object. If the
// discovery is of an input/output then the OutputType/Instance is also set. These
// are derived from the address as they are not separate parameters in the standard.
func (dc *DomainCollection) HandleDiscovery(
	address string, rawMessage string, newItem interface{}) error {

	// verify the message signature and get the payload
	_, err := messaging.VerifySenderJWSSignature(rawMessage, newItem,
		dc.GetPublicKey)

	if err != nil {
		return MakeErrorf("HandleDiscovery: Failed verifying signature on address %s: %s", address, err)
	}
	segments := strings.Split(address, "/")
	if len(segments) > 2 {
		setObjectField(newItem, "PublisherID", segments[1])
		setObjectField(newItem, "NodeID", segments[2])
	}
	if len(segments) > 4 {
		setObjectField(newItem, "OutputType", segments[3])
		setObjectField(newItem, "Instance", segments[4])
	}

	dc.Add(address, newItem)
	return nil
}

// Remove removes an object using its address.
// If the object doesn't exist, this is ignored.
func (dc *DomainCollection) Remove(address string) {
	base := MakeBaseAddress(address)
	dc.UpdateMutex.Lock()
	defer dc.UpdateMutex.Unlock()
	delete(dc.DiscoMap, base)
}

// MakeBaseAddress returns the base address without messagetype suffix
func MakeBaseAddress(address string) string {
	segments := strings.Split(address, "/")
	if len(segments) < 2 {
		return address
	}
	// remove the last segment if it is a message type (starts with $)
	lastSegment := segments[len(segments)-1]
	if strings.HasPrefix(lastSegment, "$") {
		segments = segments[:len(segments)-1]
	}
	baseAddr := strings.Join(segments, "/")
	return baseAddr
}

func setObjectField(object interface{}, fieldName string, value string) {
	valueType := reflect.ValueOf(object).Elem()
	field := valueType.FieldByName(fieldName)
	if !field.CanSet() {
		return
	}
	field.SetString(value)
}

// NewDomainCollection creates an instance for generic handling of discovered inputs, outputs and nodes
// itemPtr is a pointer to a dummy instance of the item
func NewDomainCollection(itemPtr reflect.Type, getPublicKey func(string) *ecdsa.PublicKey) DomainCollection {

	domainCollection := DomainCollection{
		DiscoMap:     make(map[string]interface{}),
		GetPublicKey: getPublicKey,
		ItemPtr:      itemPtr,
		UpdateMutex:  &sync.Mutex{},
	}
	return domainCollection
}
