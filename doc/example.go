/* Example: A Weather Publisher  [Linux]

This example creates a publisher for a weather forecast that updates the forecast every hour.
The publisher is called myweather, and each node is a city. More cities can be added with more nodes.
This example assumes you have a MQTT broker running locally.

## Step 1: Create A New Project

This uses golang modules so you can use the folder of your choice. The project folder used here is *~/Projects/iotc/myweather*. More info on go modules can be found here: https://blog.golang.org/using-go-modules

~~~bash
mkdir -p ~/Projects/iotc/myweather
cd ~/Projects/iotc/myweather
go mod init myweather
   > go: creating new go.mod: module myweather
~~~

Create a file named 'myweather.go' that looks like:
~~~golang
package main
import "github.com/hspaay/iotc.golang"
import "fmt"
func main() {
  fmt.Printf("hello, myweather\n")
}

~~~
Build and run:

~~~bash
$ go build
$ ./myweather
  > Hello, myweather
~~~

A file go.mod contains the module info include dependencies and versions of the dependencies.
Make sure go.mod and go.sum are added to your version control.

When all this works well it is time to turn it into a publisher.

## Step 2. Implement The Publisher

A simple publisher is implemented through only a few functions as shown below.

There are plenty of ways to enhance this with additional outputs and forecasts. You can even add an input to the publisher itself for adding and removing cities remotely. Please obtain a valid API key from openweathermap.
Enjoy!

Change myweather.go to look like this:
*/
package main

import (
	"encoding/json"
	"fmt"
	"github.com/hspaay/iotc.golang/iotc"
	"github.com/hspaay/iotc.golang/publisher"
	"io/ioutil"
	"net/http"
)

const appID = "myweather"
const domain = iotc.LocalDomainID
const mqttServerAddress = "localhost"
const outputTypeForecast = "forecast"

// APIKEY is a example apikey. Sign up to openweathermap.org to obtain a key"
const APIKEY = "a0e4c673de4517470b956be21abbf377"
const weatherCity = "amsterdam"
const weatherUnits = "metric"

// DefaultWeatherServiceURL contains the openweathermap URL for querying current weather
// See https://openweathermap.org/current#data for more information
const DefaultWeatherServiceURL = "https://api.openweathermap.org/data/2.5/weather?q=" +
	weatherCity + "&units=" + weatherUnits + "&appid=" + APIKEY

// AppConfig contains the application configuration and can be loaded from {AppID}.yaml
type AppConfig struct {
	PublisherID string `yaml:"publisherId"`
	WeatherURL  string `yaml:"weatherURL"`
}

// CurrentWeather is the struct to load the openweathermap result
type CurrentWeather struct {
	Main struct {
		Humidity    int     `json:"humidity"`
		Temperature float32 `json:"temp"` //
	} `json:"main"`
}

// Sign up to openweathermap.org to obtain a key
var appConfig = &AppConfig{
	PublisherID: appID,
	WeatherURL:  DefaultWeatherServiceURL,
}

// SetupNodes creates a node for each city with outputs for temperature and humidity
func SetupNodes(pub *publisher.Publisher, city string) {
	pub.NewNode(city, iotc.NodeTypeWeatherService)
	output := pub.NewOutput(city, iotc.OutputTypeTemperature, iotc.DefaultOutputInstance)
	output.Unit = iotc.UnitCelcius
	pub.Outputs.UpdateOutput(output)
	pub.NewOutput(city, iotc.OutputTypeHumidity, iotc.DefaultOutputInstance)
}

// UpdateWeather obtains the forecast and updates the output value.
// The iotc library will automatically publish the output discovery and values.
func UpdateWeather(pub *publisher.Publisher) {
	nodeID := weatherCity
	// allow custom weather URL from a node configuration, fall back to the default URL
	requestURL, _ := pub.GetNodeConfigValue(nodeID, "url", DefaultWeatherServiceURL)
	resp, err := http.Get(requestURL)

	if err == nil {
		var weather CurrentWeather
		forecastRaw, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(forecastRaw, &weather)

		// This publishes the forecast on local/myweather/amsterdam/$value/forecast/0
		temp := fmt.Sprintf("%0.1f", weather.Main.Temperature)
		hum := fmt.Sprintf("%d", weather.Main.Humidity)
		pub.UpdateOutputValue(nodeID, iotc.OutputTypeTemperature, iotc.DefaultOutputInstance, temp)
		pub.UpdateOutputValue(nodeID, iotc.OutputTypeHumidity, iotc.DefaultOutputInstance, hum)
		pub.SetNodeErrorStatus(nodeID, iotc.NodeRunStateReady, "Forecast loaded successfully")
	} else {
		// pub.SetOutputError(nodeID, OutputType, OutputTypeTemperature, "Forecast not available")
		pub.SetNodeErrorStatus(nodeID, iotc.NodeRunStateError, "Forecast not available")
	}
}

// Run the example
func main() {
	// this auto loads the messenger.yaml and myweather.yaml from ~/.config/iotc
	pub, _ := publisher.NewAppPublisher(appID, "", "", appConfig, false)

	SetupNodes(pub, weatherCity)
	// Update the forecast once an hour
	pub.SetPollInterval(3600, UpdateWeather)

	pub.Start()
	pub.WaitForSignal()
	pub.Stop()
}
