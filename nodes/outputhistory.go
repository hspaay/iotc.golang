// Package nodes with handling of node output history values
package nodes

import (
	"sync"
	"time"

	"github.com/hspaay/iotconnect.golang/standard"
)

// OutputHistoryList with output history value management
type OutputHistoryList struct {
	outputHistory  map[string]standard.HistoryList // history lists by output address
	updatedOutputs map[string]string               // addresses of updated outputs
	updateMutex    *sync.Mutex                     // mutex for async updating of outputs
}

// GetHistory returns the history list
// Returns nil if the type or instance is unknown
func (historylist *OutputHistoryList) GetHistory(address string) standard.HistoryList {
	historylist.updateMutex.Lock()
	var historyList = historylist.outputHistory[address]
	historylist.updateMutex.Unlock()
	return historyList
}

// GetOutputValueByAddress returns the most recent output value by output discovery address
// This returns a HistoryValue object with the latest value and timestamp it was updated
func (historylist *OutputHistoryList) GetOutputValueByAddress(address string) *standard.HistoryValue {
	var latest *standard.HistoryValue

	historylist.updateMutex.Lock()
	history := historylist.outputHistory[address]
	historylist.updateMutex.Unlock()

	if history == nil || len(history) == 0 {
		return nil
	}
	latest = history[0]
	return latest
}

// GetOutputValueByType returns the current output value by output type and instance
func (historylist *OutputHistoryList) GetOutputValueByType(node *standard.Node, outputType string, instance string) *standard.HistoryValue {
	addr := standard.MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, instance)
	return historylist.GetOutputValueByAddress(addr)
}

// GetUpdatedOutputs returns a list of output addresses that have updated values
// clear the update outputs list on return
func (historylist *OutputHistoryList) GetUpdatedOutputs(clearUpdates bool) []string {
	var addrList []string = make([]string, 0)

	historylist.updateMutex.Lock()
	if historylist.updatedOutputs != nil {
		for _, addr := range historylist.updatedOutputs {
			addrList = append(addrList, addr)
		}
		if clearUpdates {
			historylist.updatedOutputs = nil
		}
	}
	historylist.updateMutex.Unlock()
	return addrList
}

// UpdateOutputValue adds the new node output value to the front of the history
// If the node has a repeatDelay configured, then the value is only added if
//  it has changed or the previous update was older than the repeatDelay.
// The history retains a max of 24 hours
// returns true if history is updated, false if history has not been updated
func (historylist *OutputHistoryList) UpdateOutputValue(node *standard.Node, outputType string, instance string, newValue string) bool {
	var previous *standard.HistoryValue
	var repeatDelay = 3600 // default repeat delay
	var ageSeconds = -1
	var hasUpdated = false

	addr := standard.MakeOutputDiscoveryAddress(node.Zone, node.PublisherID, node.ID, outputType, instance)

	historylist.updateMutex.Lock()
	// auto create the output if it hasn't been discovered yet
	// output := historylist.Outputs.GetOutputByAddress(addr)
	history := historylist.outputHistory[addr]

	// only update output if value changes or delay has passed
	if node.RepeatDelay != 0 {
		repeatDelay = node.RepeatDelay
	}
	if len(history) > 0 {
		previous = history[0]
		age := time.Now().Sub(previous.Timestamp)
		ageSeconds = int(age.Seconds())
	}

	doUpdate := ageSeconds < 0 || ageSeconds > repeatDelay || newValue != previous.Value
	if doUpdate {
		newHistory := updateHistory(history, newValue, node.HistorySize)

		historylist.outputHistory[addr] = newHistory
		hasUpdated = true

		if historylist.updatedOutputs == nil {
			historylist.updatedOutputs = make(map[string]string)
		}
		historylist.updatedOutputs[addr] = addr

	}
	historylist.updateMutex.Unlock()
	return hasUpdated
}

// updateHistory inserts a new value at the front of the history
// The resulting list contains a max of historySize entries limited to 24 hours
// This function is not thread-safe and should only be used from within a locked section
// history is optional and used to insert the value in the front. If nil then a new history is returned
// newValue contains the value to include in the history along with the current timestamp
// maxHistorySize is optional and limits the size in addition to the 24 hour limit
// returns the history list with the new value at the front of the list
func updateHistory(history standard.HistoryList, newValue string, maxHistorySize int) standard.HistoryList {

	timeStamp := time.Now()
	// timeStampStr := timeStamp.Format("2006-01-02T15:04:05.000-0700")

	latest := standard.HistoryValue{
		Timestamp: timeStamp,
		// TimeStamp: timeStampStr,
		Value: newValue,
	}
	if history == nil {
		history = make(standard.HistoryList, 1)
	} else {
		// make room at the front of the slice
		history = append(history, &latest)
		copy(history[1:], history[0:])
	}
	history[0] = &latest

	// remove old entries, determine the max
	if maxHistorySize == 0 || len(history) < maxHistorySize {
		maxHistorySize = len(history)
	}
	// cap at 24 hours
	for ; maxHistorySize > 1; maxHistorySize-- {
		entry := history[maxHistorySize-1]
		if timeStamp.Sub(entry.Timestamp) <= time.Hour*24 {
			break
		}
	}
	history = history[0:maxHistorySize]
	return history
}

// // Get the latest output historyvalue
// // This is not thread-safe. Use within a locked area
// func (historylist *HistoryList) getLatestOutputValue(inoutput *standard.InOutput) *standard.HistoryValue {
// 	history := historylist.outputHistory[inoutput.Address]
// 	if history == nil || len(history) == 0 {
// 		return nil
// 	}
// 	return history[0]
// }

// NewOutputHistoryList creates a new instance for output value history management
func NewOutputHistoryList() *OutputHistoryList {
	outputs := OutputHistoryList{
		outputHistory: make(map[string]standard.HistoryList),
		updateMutex:   &sync.Mutex{},
	}
	return &outputs
}
