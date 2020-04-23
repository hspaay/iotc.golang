// Package nodes with handling of node output values
package nodes

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/hspaay/iotconnect.golang/messaging"
)

// OutputHistory with output history value management
type OutputHistory struct {
	historyLists   map[string]messaging.OutputHistoryList // history lists by output address
	updatedOutputs map[string]string                      // addresses of updated outputs
	updateMutex    *sync.Mutex                            // mutex for async updating of outputs
}

// GetHistory returns the history list
// Returns nil if the type or instance is unknown
func (outputValues *OutputHistory) GetHistory(address string) messaging.OutputHistoryList {
	outputValues.updateMutex.Lock()
	var historyList = outputValues.historyLists[address]
	outputValues.updateMutex.Unlock()
	return historyList
}

// GetOutputValueByAddress returns the most recent output value by output discovery address
// This returns a HistoryValue object with the latest value and timestamp it was updated
func (outputValues *OutputHistory) GetOutputValueByAddress(address string) *messaging.OutputValue {
	var latest *messaging.OutputValue

	outputValues.updateMutex.Lock()
	history := outputValues.historyLists[address]
	outputValues.updateMutex.Unlock()

	if history == nil || len(history) == 0 {
		return nil
	}
	latest = &history[0]
	return latest
}

// GetOutputValueByType returns the current output value by output type and instance
func (outputValues *OutputHistory) GetOutputValueByType(node *Node, outputType string, instance string) *messaging.OutputValue {
	addr := MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, instance)
	return outputValues.GetOutputValueByAddress(addr)
}

// GetUpdatedOutputs returns a list of output addresses that have updated values
// clear the update outputs list on return
func (outputValues *OutputHistory) GetUpdatedOutputs(clearUpdates bool) []string {
	var addrList []string = make([]string, 0)

	outputValues.updateMutex.Lock()
	if outputValues.updatedOutputs != nil {
		for _, addr := range outputValues.updatedOutputs {
			addrList = append(addrList, addr)
		}
		if clearUpdates {
			outputValues.updatedOutputs = nil
		}
	}
	outputValues.updateMutex.Unlock()
	return addrList
}

// UpdateOutputFloatList adds a list of floats as the output value in the format: "[value1, value2, ...]"
func (outputValues *OutputHistory) UpdateOutputFloatList(node *Node, outputType string, outputInstance string, values []float32) bool {
	valuesAsString, _ := json.Marshal(values)
	return outputValues.UpdateOutputValue(node, outputType, outputInstance, string(valuesAsString))
}

// UpdateOutputIntList adds a list of integers as the output value in the format: "[value1, value2, ...]"
func (outputValues *OutputHistory) UpdateOutputIntList(node *Node, outputType string, outputInstance string, values []int) bool {
	valuesAsString, _ := json.Marshal(values)
	return outputValues.UpdateOutputValue(node, outputType, outputInstance, string(valuesAsString))
}

// UpdateOutputStringList adds a list of strings as the output value in the format: "[value1, value2, ...]"
func (outputValues *OutputHistory) UpdateOutputStringList(node *Node, outputType string, outputInstance string, values []string) bool {
	valuesAsString, _ := json.Marshal(values)
	return outputValues.UpdateOutputValue(node, outputType, outputInstance, string(valuesAsString))
}

// UpdateOutputValue adds the new node output value to the front of the history
// If the node has a repeatDelay configured, then the value is only added if
//  it has changed or the previous update was older than the repeatDelay.
// The history retains a max of 24 hours
// returns true if history is updated, false if history has not been updated
func (outputValues *OutputHistory) UpdateOutputValue(node *Node, outputType string, instance string, newValue string) bool {
	var previous *messaging.OutputValue
	var repeatDelay = 3600 // default repeat delay
	var ageSeconds = -1
	var hasUpdated = false

	addr := MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, instance)

	outputValues.updateMutex.Lock()
	// auto create the output if it hasn't been discovered yet
	// output := outputvalue.Outputs.GetOutputByAddress(addr)
	history := outputValues.historyLists[addr]

	// only update output if value changes or delay has passed
	if node.RepeatDelay != 0 {
		repeatDelay = node.RepeatDelay
	}
	if len(history) > 0 {
		previous = &history[0]
		prevTime := time.Unix(previous.EpochTime, 0)
		age := time.Now().Sub(prevTime)
		ageSeconds = int(age.Seconds())
	}

	doUpdate := ageSeconds < 0 || ageSeconds > repeatDelay || newValue != previous.Value
	if doUpdate {
		newHistory := updateHistory(history, newValue, node.HistorySize)

		outputValues.historyLists[addr] = newHistory
		hasUpdated = true

		if outputValues.updatedOutputs == nil {
			outputValues.updatedOutputs = make(map[string]string)
		}
		outputValues.updatedOutputs[addr] = addr

	}
	outputValues.updateMutex.Unlock()
	return hasUpdated
}

// updateHistory inserts a new value at the front of the history
// The resulting list contains a max of historySize entries limited to 24 hours
// This function is not thread-safe and should only be used from within a locked section
// history is optional and used to insert the value in the front. If nil then a new history is returned
// newValue contains the value to include in the history along with the current timestamp
// maxHistorySize is optional and limits the size in addition to the 24 hour limit
// returns the history list with the new value at the front of the list
func updateHistory(history messaging.OutputHistoryList, newValue string, maxHistorySize int) messaging.OutputHistoryList {

	timeStamp := time.Now()
	timeStampStr := timeStamp.Format(messaging.TimeFormat)

	latest := messaging.OutputValue{
		Timestamp: timeStampStr,
		EpochTime: timeStamp.Unix(),
		Value:     newValue,
	}
	if history == nil {
		history = make(messaging.OutputHistoryList, 1)
	} else {
		// make room at the front of the slice
		history = append(history, latest)
		copy(history[1:], history[0:])
	}
	history[0] = latest

	// remove old entries, determine the max
	if maxHistorySize == 0 || len(history) < maxHistorySize {
		maxHistorySize = len(history)
	}
	// cap at 24 hours
	for ; maxHistorySize > 1; maxHistorySize-- {
		entry := history[maxHistorySize-1]
		entrytime := time.Unix(entry.EpochTime, 0)
		if timeStamp.Sub(entrytime) <= time.Hour*24 {
			break
		}
	}
	history = history[0:maxHistorySize]
	return history
}

// NewOutputValue creates a new instance for output value and history management
func NewOutputValue() *OutputHistory {
	outputs := OutputHistory{
		historyLists: make(map[string]messaging.OutputHistoryList),
		updateMutex:  &sync.Mutex{},
	}
	return &outputs
}
