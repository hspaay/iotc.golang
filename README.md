# iotc.golang

iotc.golang is an implementation of the IoTConnect (iotc) standard for publishing and subscribing to IoT information on a message bus. This library is part of the reference implementation. The standard can be found at: https://github.com/hspaay/iotc.standard

## Status

This readme is currently under construction... 
The golang library is functional but should be considered Alpha code.
The current focus is on improving this library by adding adapters that use it.

TODO:
* TLS connection to MQTT brokers
* ...

## Audience

This library is intended for software developers that wish to develop IoT applications in the golang language.
A similar library for the Python and Javascript/Typescript languages is planned for the future.

## This Library Provides

* systemd launcher of adapters for Linux 
* Messenger for MQTT brokers
* Management of nodes, inputs and outputs (see IoTConnect standard for further explanation)
* Publish discovery when nodes are updated 
* Publish updates to output values 
* Signing of published messages
* Hook to handle node input control messages
* Hook to handle node configuration updates
* Constants and Type Definitions of the IoTConnect standard

## Prerequisites

1. Golang
   
   This guide assumes that you are familiar with programming in golang, and have golang 1.13 or newer installed. If you are new to golang, check out their website https://golang.org/doc/install for more information. 

2. MQTT broker
  
   A working [MQTT broker](https://en.wikipedia.org/wiki/MQTT) (message bus) is needed to test and run a publisher. [Mosquitto](https://mosquitto.org/) is a lightweight MQTT broker that runs nicely on Linux (including Raspberry-pi), Mac, and Windows. Only a single broker is needed for all your publishers. For a home automation application you will do fine with running Mosquitto on a Raspberry-pi 2, 3 or 4 connected to a small UPS and park it somewhere out of sight.

   For industrial or government applications that require GDPR or SOC2 compliance, look at enterprise message brokers such as [HiveMQT](www.hivemq.com), [RabbitMQT](https://www.rabbitmq.com/), [Apache ActiveMQ](https://activemq.apache.org/). For hosted cloud versions look at [CloudMQTT](www.cloudmqtt.com) (uses managed Mosquitto instances under the hood). Amazon AWS and Google also support IoT message buses and are worth a look.

3. Access to git and github.com is needed to retrieve the development libraries. 

4. A code editor or IDE is needed to edit source code. Visual Studio Code is a free IDE that is quite popular and has good support for golang. See https://code.visualstudio.com/ for more information and downloads.

## How To Use This Library

under construction...

The first part describes the project setup to start building. The second part shows how to create a simple weather forecast publisher. The instructions are for Linux but building for MacOS or Windows should not be too different. 

This example uses Go modules as this lets you control versioning and choose your own project folder location.

## Installing Publishers

Recommended installation of your publisher on a linux platform. The iotc namespace is used for publisher applications:

The folder structure for deployment as a normal user:
* ~/bin/iotc/bin      location of the publisher binaries
* ~/bin/iotc/config   location of the configuration files, including iotc.conf
* ~/bin/iotc/logs     logging output

When deploying as an system application, create these folders and update /etc/iotc.conf
* /etc/iotc/conf         location of iotc.conf main configuration file
* /opt/iotc/             location of the publisher binaries
* /var/lib/iotc/         location of the persistence files
* /var/log/iotc/         location of iotc log files

Starting a publisher using systemd 
1. Edit the paths in iotc.service to make sure the folder and user IDs are correct
2. Copy the iotc@.service and iotc.target files to /etc/systemd/system/
3. Start manually using systemd using myweather as an example:
   $ sudo service iotc@myweather start
4. To enable autostart:
   $ sudo systemctl enable iotc@myweather
5. To disable autostart:
   $ sudo systemctl disable iotc@myweather

## Contributing

Contribution to the IoTConnect project is welcome. There are many areas where help is needed, especially with building publishers for IoT and other devices.
See [CONTRIBUTING](docs/CONTRIBUTING.md) for guidelines.

## Questions

For issues and questions, please open a ticket.
Common questions will be captured in the [Q&A](docs/FAQ.md).


# Example: A Weather Publisher 

... under construction ...

This example creates a publisher for a weather forecast that updates the forecast every hour. The project folder is *~/Projects/iotc/myweather*
The publisher is called myweather, and each node is a city. More cities can be added with more nodes. 
This example assumes you have a MQTT broker running locally. 

## Step 1: Create A New Project

This uses golang modules so you can use the folder of your choice. More info here: https://blog.golang.org/using-go-modules

~~~bash
$ mkdir -p ~/Projects/iotc/myweather
$ cd ~/Projects/iotc/myweather
$ go mod init myweather
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

## Step 2. Implement The Publisher

A basic publisher is implemented through only a few functions as shown below.

Change myweather.go to look like this:

~~~golang
package myweather

import "github.com/hspaay/iotc"

const ZoneID = standard.LocalZoneID
const MqttServerAddress = "localhost"
const PublisherId = "myweather"
const NodeId = "amsterdam"  
const OutputTypeForecast = "forecast"
// this is their example apikey. Sign up to openweathermap.org to obtain a key for your app"
const APIKEY = "b6907d289e10d714a6e88b30761fae22" 
const DefaultWeatherServiceUrl = "https://api.openweathermap.org/data/2.5/weather?q=amsterdam&appid="+APIKEY

var publisher publisher.Publisher

func main() {
  messenger = NewMessenger(MqttServerAddress, 0) // use default mqtt port
  
  publisher = publisher.NewPublisher(ZoneID, PublisherID, messenger)

  // Update the forecast once an hour
  iotc.SetDefaultOutputInterval(3600, this.PollOutputs)

  // See below for Discover and Poll functions
  publisher.Start(nil, nil, Discover, Poll)
}

// Discover creates a node for each city with configuration for latitude, longitude.
function Discover(publisher *Publisher) {
  node = publisher.DiscoverNode(NodeId) // add/update node with ID forecast
  publisher.SetNodeDefaultConfig(node, "url", DataTypeString, DefaultWeatherServiceUrl)
  publisher.DiscoverOutput(node, OutputTypeForecast, "temperature")
  publisher.DiscoverOutput(node, OutputTypeForecast, "humidity")
}

// Poll obtains the forcast and updates the output value.
// The iotc library will automatically publish the output discovery and values.
function Poll(publisher *Publisher) {
  node = publisher.GetNode(NodeId)
  configValues = publisher.GetNodeConfigValues(node)
  forecastRaw, err = httplib.get(configValues["url"]))
  if (err == nil) {
      forecastObject := JSON.parse(forecastRaw)
      mainForecast := forecastObject["main"]
      // This publishes the forecast on $local/myweather/amsterdam/$value/forecast/0
      publisher.UpdateOutput(node, OutputType, OutputTypeTemperature, main["temp"])
      publisher.UpdateOutput(node, OutputType, OutputTypeHumidity, main["humidity"])
  } else {
    publisher.UpdateOutputError(node, OutputType, OutputTypeTemperature, "Forecast not available")
    publisher.UpdateOutputError(node, OutputType, OutputTypeHumidity, "Forecast not available")
  }
}
~~~

There are plenty of ways to enhance this with additional outputs and forecasts. You can even add an input to the publisher itself for adding and removing cities remotely. Please obtain a valid API key from openweathermap.
Enjoy!


## Step 3. Install Mosquitto

[Mosquitto](https://mosquitto.org/) is a lightweight MQTT server. Installation for the different platforms[is described here](https://mosquitto.org/download/).

Mosquitto needs little configuration as [described on their documentation](https://mosquitto.org/man/mosquitto-conf-5.html).
On Linux the configuration takes place in the /etc/mosquitto/mosquitto.conf file.

### Basic Configuration

No configuration neccesary as the defaults are good to go. 

### Advanced Configuration

Unless you run locally on a trusted LAN, you're going to want to:
1. Use SSL/TLS for secure connections
2. Add authentication for your publishers 



### Automatically Start Mosquitto
On Linux:
$ sudo systemctl enable mosquitto
$ sudo systemctl start mosquitto


## Last, Run!

... work in progress ...
