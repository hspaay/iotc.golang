// Package outputs with updating and publishing of output forecasts
package outputs

import (
	"sync"

	"github.com/iotdomain/iotdomain-go/types"
)

// OutputForecast with forecasted values
type OutputForecast []types.OutputValue

// RegisteredForecastValues with registered forecasts for outputs
// A forecast is a list of timestamps with future projected values similar to history
type RegisteredForecastValues struct {
	domain           string // domain this forcast list belongs to
	publisherID      string // publisher of the forcasts
	forecastMap      map[string]OutputForecast
	updateMutex      *sync.Mutex
	updatedForecasts map[string]string // address of updated forecasts
}

// GetForecast returns the output's forecast by output address
// outputAddress is the discovery address of the output
// Returns nil if the type or instance is unknown or no forecast is available
func (regForecasts *RegisteredForecastValues) GetForecast(
	nodeID string, outputType types.OutputType, instance string) OutputForecast {
	regForecasts.updateMutex.Lock()
	defer regForecasts.updateMutex.Unlock()

	outputAddress := MakeOutputDiscoveryAddress(regForecasts.domain, regForecasts.publisherID, nodeID, outputType, instance)
	var forecast = regForecasts.forecastMap[outputAddress]
	return forecast
}

// GetUpdatedForecasts returns a list of output addresses that have updated forecasts
// clearUpdates clears the update list on return
func (regForecasts *RegisteredForecastValues) GetUpdatedForecasts(clearUpdates bool) []string {
	var addrList []string = make([]string, 0)

	regForecasts.updateMutex.Lock()
	defer regForecasts.updateMutex.Unlock()

	if regForecasts.updatedForecasts != nil {
		for _, addr := range regForecasts.updatedForecasts {
			addrList = append(addrList, addr)
		}
		if clearUpdates {
			regForecasts.updatedForecasts = nil
		}
	}
	return addrList
}

// UpdateForecast updates the output forecast list of values
func (regForecasts *RegisteredForecastValues) UpdateForecast(
	nodeID string, outputType types.OutputType, instance string, forecast OutputForecast) {

	outputAddress := MakeOutputDiscoveryAddress(regForecasts.domain, regForecasts.publisherID, nodeID, outputType, instance)

	regForecasts.updateMutex.Lock()
	defer regForecasts.updateMutex.Unlock()

	regForecasts.forecastMap[outputAddress] = forecast
	// output := publisher.Outputs.GetOutputByAddress(addr)
	// aliasAddress := publisher.getOutputAliasAddress(addr)

	if regForecasts.updatedForecasts == nil {
		regForecasts.updatedForecasts = make(map[string]string)
	}
	regForecasts.updatedForecasts[outputAddress] = outputAddress

	// publisher.publishForecast(aliasAddress, output)
}

// NewRegisteredForecastValues creates a new instance for storing output forecasts
func NewRegisteredForecastValues(domain string, publisherID string) *RegisteredForecastValues {
	rfv := RegisteredForecastValues{
		domain:      domain,
		publisherID: publisherID,
		forecastMap: make(map[string]OutputForecast),
		updateMutex: &sync.Mutex{},
	}
	return &rfv
}
