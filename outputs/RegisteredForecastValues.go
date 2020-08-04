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
	updatedForecasts map[string]string // map of output IDs with updated forecasts
}

// GetForecast returns the output's forecast by outputID
// Returns nil if the output has no forecast
func (regForecasts *RegisteredForecastValues) GetForecast(outputID string) OutputForecast {
	regForecasts.updateMutex.Lock()
	defer regForecasts.updateMutex.Unlock()

	var forecast = regForecasts.forecastMap[outputID]
	return forecast
}

// GetUpdatedForecasts returns a list of output IDs that have updated forecasts
// clearUpdates clears the update list on return
func (regForecasts *RegisteredForecastValues) GetUpdatedForecasts(clearUpdates bool) []string {
	var idList []string = make([]string, 0)

	regForecasts.updateMutex.Lock()
	defer regForecasts.updateMutex.Unlock()

	if regForecasts.updatedForecasts != nil {
		for _, fcID := range regForecasts.updatedForecasts {
			idList = append(idList, fcID)
		}
		if clearUpdates {
			regForecasts.updatedForecasts = nil
		}
	}
	return idList
}

// UpdateForecast updates the output forecast list of values
func (regForecasts *RegisteredForecastValues) UpdateForecast(
	outputID string, forecast OutputForecast) {

	regForecasts.updateMutex.Lock()
	defer regForecasts.updateMutex.Unlock()

	regForecasts.forecastMap[outputID] = forecast

	if regForecasts.updatedForecasts == nil {
		regForecasts.updatedForecasts = make(map[string]string)
	}
	regForecasts.updatedForecasts[outputID] = outputID

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
