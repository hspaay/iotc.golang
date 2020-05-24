// Package nodes with handling of node output values
package nodes

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/hspaay/iotc.golang/iotc"
)

// OutputHistory with history values
type OutputHistory []iotc.OutputValue

// OutputValueList with output values for all outputs
type OutputValueList struct {
	historyMap     map[string]OutputHistory // history lists by output address
	updateMutex    *sync.Mutex              // mutex for async updating of outputs
	updatedOutputs map[string]string        // addresses of updated outputs
}

// GetHistory returns the history list
// Returns nil if the type or instance is unknown
func (outputValues *OutputValueList) GetHistory(address string) OutputHistory {
	outputValues.updateMutex.Lock()
	var historyList = outputValues.historyMap[address]
	outputValues.updateMutex.Unlock()
	return historyList
}

// GetOutputValueByAddress returns the most recent output value by output discovery address
// This returns a HistoryValue object with the latest value and timestamp it was updated
func (outputValues *OutputValueList) GetOutputValueByAddress(address string) *iotc.OutputValue {
	var latest *iotc.OutputValue

	outputValues.updateMutex.Lock()
	defer outputValues.updateMutex.Unlock()

	history := outputValues.historyMap[address]

	if history == nil || len(history) == 0 {
		return nil
	}
	latest = &history[0]
	return latest
}

// GetOutputValueByType returns the current output value by output type and instance
func (outputValues *OutputValueList) GetOutputValueByType(node *iotc.NodeDiscoveryMessage, outputType string, instance string) *iotc.OutputValue {
	addr := MakeOutputDiscoveryAddress(node.Address, outputType, instance)
	return outputValues.GetOutputValueByAddress(addr)
}

// GetUpdatedOutputs returns a list of output discovery addresses that have updated values
// clear the update outputs list on return
func (outputValues *OutputValueList) GetUpdatedOutputs(clearUpdates bool) []string {
	var addrList []string = make([]string, 0)

	outputValues.updateMutex.Lock()
	defer outputValues.updateMutex.Unlock()

	if outputValues.updatedOutputs != nil {
		for _, addr := range outputValues.updatedOutputs {
			addrList = append(addrList, addr)
		}
		if clearUpdates {
			outputValues.updatedOutputs = nil
		}
	}
	return addrList
}

// UpdateOutputFloatList adds a list of floats as the output value in the format: "[value1, value2, ...]"
func (outputValues *OutputValueList) UpdateOutputFloatList(address string, values []float32) bool {
	valuesAsString, _ := json.Marshal(values)
	return outputValues.UpdateOutputValue(address, string(valuesAsString))
}

// UpdateOutputIntList adds a list of integers as the output value in the format: "[value1, value2, ...]"
func (outputValues *OutputValueList) UpdateOutputIntList(address string, values []int) bool {
	valuesAsString, _ := json.Marshal(values)
	return outputValues.UpdateOutputValue(address, string(valuesAsString))
}

// UpdateOutputStringList adds a list of strings as the output value in the format: "[value1, value2, ...]"
func (outputValues *OutputValueList) UpdateOutputStringList(address string, values []string) bool {
	valuesAsString, _ := json.Marshal(values)
	return outputValues.UpdateOutputValue(address, string(valuesAsString))
}

// UpdateOutputValue adds the new node output value to the front of the history
// If the node has a repeatDelay configured, then the value is only added if
//  it has changed, or if the previous update was older than the repeatDelay.
// The history retains a max of 24 hours
// returns true if history is updated, false if history has not been updated
func (outputValues *OutputValueList) UpdateOutputValue(address string, newValue string) bool {
	var previous *iotc.OutputValue
	var repeatDelay = 3600 // default repeat delay is 1 hour
	var ageSeconds = -1
	var hasUpdated = false

	outputValues.updateMutex.Lock()
	defer outputValues.updateMutex.Unlock()

	// auto create the output if it hasn't been discovered yet
	// output := outputvalue.Outputs.GetOutputByAddress(addr)
	history := outputValues.historyMap[address]

	// only update output if value changes or delay has passed
	// for now use 1 hour repeat delay. Need to get the config from somewhere
	// if history.RepeatDelay != 0 {
	// 	repeatDelay = history.RepeatDelay
	// }
	if len(history) > 0 {
		previous = &history[0]
		prevTime := time.Unix(previous.EpochTime, 0)
		age := time.Now().Sub(prevTime)
		ageSeconds = int(age.Seconds())
	}
	doUpdate := ageSeconds < 0 || ageSeconds > repeatDelay || newValue != previous.Value
	if doUpdate {
		// 24 hour history
		newHistory := updateHistory(history, newValue, 0)

		outputValues.historyMap[address] = newHistory
		hasUpdated = true

		if outputValues.updatedOutputs == nil {
			outputValues.updatedOutputs = make(map[string]string)
		}
		outputValues.updatedOutputs[address] = address

	}
	return hasUpdated
}

// updateHistory inserts a new value at the front of the history
// The resulting list contains a max of historySize entries limited to 24 hours
// This function is not thread-safe and should only be used from within a locked section
// history is optional and used to insert the value in the front. If nil then a new history is returned
// newValue contains the value to include in the history along with the current timestamp
// maxHistorySize is optional and limits the size in addition to the 24 hour limit
// returns the history list with the new value at the front of the list
func updateHistory(history OutputHistory, newValue string, maxHistorySize int) OutputHistory {

	timeStamp := time.Now()
	timeStampStr := timeStamp.Format(iotc.TimeFormat)

	latest := iotc.OutputValue{
		Timestamp: timeStampStr,
		EpochTime: timeStamp.Unix(),
		Value:     newValue,
	}
	if history == nil {
		history = make(OutputHistory, 1)
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

// NewOutputValueList creates a new instance for output value and history management
func NewOutputValueList() *OutputValueList {
	outputs := OutputValueList{
		historyMap:  make(map[string]OutputHistory),
		updateMutex: &sync.Mutex{},
	}
	return &outputs
}
