// Package nodes with updating and publishing of output forecasts
package nodes

import (
	"sync"

	"github.com/hspaay/iotc.golang/iotc"
)

// OutputForecast with forecasted values
type OutputForecast []iotc.OutputValue

// OutputForecastList with management of forecasts for outputs
// A forecast is a list of timestamps with projected values
type OutputForecastList struct {
	forecastMap    map[string]OutputForecast
	updateMutex    *sync.Mutex
	updatedOutputs map[string]string
}

// GetForecast returns the output's forecast by output address
// outputAddress is the discovery address of the output
// Returns nil if the type or instance is unknown or no forecast is available
func (forecastList *OutputForecastList) GetForecast(outputAddress string) OutputForecast {
	forecastList.updateMutex.Lock()
	defer forecastList.updateMutex.Unlock()

	var forecast = forecastList.forecastMap[outputAddress]
	return forecast
}

// GetUpdatedForecasts returns a list of output addresses that have updated forecasts
// clearUpdates clears the update list on return
func (forecastList *OutputForecastList) GetUpdatedForecasts(clearUpdates bool) []string {
	var addrList []string = make([]string, 0)

	forecastList.updateMutex.Lock()
	defer forecastList.updateMutex.Unlock()

	if forecastList.updatedOutputs != nil {
		for _, addr := range forecastList.updatedOutputs {
			addrList = append(addrList, addr)
		}
		if clearUpdates {
			forecastList.updatedOutputs = nil
		}
	}
	return addrList
}

// UpdateForecast publishes the output forecast list of values
// outputAddress is the discovery address of the output
func (forecastList *OutputForecastList) UpdateForecast(
	nodeAddress string, outputType iotc.OutputType, instance string, forecast OutputForecast) {

	outputAddress := MakeOutputDiscoveryAddress(nodeAddress, outputType, instance)

	forecastList.updateMutex.Lock()
	defer forecastList.updateMutex.Unlock()

	forecastList.forecastMap[outputAddress] = forecast
	// output := publisher.Outputs.GetOutputByAddress(addr)
	// aliasAddress := publisher.getOutputAliasAddress(addr)

	if forecastList.updatedOutputs == nil {
		forecastList.updatedOutputs = make(map[string]string)
	}
	forecastList.updatedOutputs[outputAddress] = outputAddress

	// publisher.publishForecast(aliasAddress, output)
}

// NewOutputForecastList creates a new instance of a forecast list
func NewOutputForecastList() *OutputForecastList {
	fcl := OutputForecastList{
		forecastMap: make(map[string]OutputForecast),
		updateMutex: &sync.Mutex{},
	}
	return &fcl
}
