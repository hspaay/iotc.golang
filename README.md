# iotzone.golang
iotzone.golang is the core library in the Golang language *for developers of iotzone publishers*. It serves as a reference implementation in the golang language. 

iotzone is an implementation of the IotConnect standard for publishing and subscribing to IoT information on a message bus. This library is part of the reference implementation. The standard can be found at: https://github.com/hspaay/iotzone.standard

## This Library Provides
* systemd launcher for linux 
* Management of nodes, inputs and outputs (see IotConnect standard for further explanation)
* Auto publish discovery when nodes and configuration are updated 
* Auto publish updates to output values 
* Signing of published messages
* Provide hooks to handle control input messages
* Provide hooks to handle configuration updates
* Define the iotzone data types in Golang

## Getting Started (Linux)

This section describes how to get started building your own iotzone publisher in golang using this library. The first part describes the project setup to start building. The second part shows how to create a simple weather forecast publisher. The instructions are for Linux but building for MacOS or Windows should not be too different. 

This example uses Go modules as this lets you control versioning and choose your own project folder location.

## Prerequisites
This guide assumes that you are familiar with programming in golang, and golang 1.13 or newer is installed on your development PC. If you are new to golang, check out their website https://golang.org/doc/install for more information. 

A working MQTT broker (server) is needed to test and run a publisher. Mosquitto is a lightweight MQTT broker that runs nicely on Linux (including Raspberry-pi), Mac, and Windows. More info can be found here: https://mosquitto.org/. Only a single broker is needed for all you publishers. For a home automation application you will do fine with running Mosquitto on a Raspberry-pi 2, 3 or 4 connected to a small UPS and park it somewhere out of sight.

For commercial, industrial, or government applications the bus requirements are likely more demanding. Various commercial 

Access to git and github.com is needed to retrieve the development libraries. 

A code editor or IDE is needed to edit source code. Visual Studio Code is a free IDE that is quite popular and has good support for golang. See https://code.visualstudio.com/ for more information and downloads.

## Installing

Other than the prerequisites above no other software needs to be installed to start developing a iotzone publisher in golang. 


## Developing Tests

It is highly recommended to use the golang testing facilities. Included in the example are a few basic test cases. To run the tests, the mosquitto broker must be running and reachable on the local network. 

## Deployment (Linux)

The folder structure for deployment as a normal user is:
* ~/bin/iotzone/bin      location of the publisher binaries
* ~/bin/iotzone/config   location of the configuration files, including iotzone.conf
* ~/bin/iotzone/logs     logging output

When deploying as an application, create these folders and update /etc/iotzone.conf
* /opt/iotzone/             location of the publisher binaries
* /etc/iotzone/conf         location of iotzone.conf main configuration file
* /var/lib/iotzone/         location of the persistence files
* /var/log/iotzone/         location of iotzone log files

Starting a publisher using systemd 
1. Copy the iotzone@.service and iotzone.target files to /etc/systemd/system/
2. Edit the paths in iotzone.service to make sure the folder and user IDs are correct
3. Start manually using systemd:
   $ sudo service iotzone@myweather start
4. To start on bootup:
   $ sudo systemctl enable iotzone@myweather

## Contributing
Contribution to the iotzone project is welcome. There are many areas where help is needed, especially with building publishers for IoT and other devices.
See [CONTRIBUTING](docs/CONTRIBUTING.md) for guidelines.

## Questions
For issues and questions, please open a ticket.
Common questions will be captured in the [Q&A](docs/FAQ.md).


# Creating A Publisher 

This example creates a publisher for a weather forecast that updates the forecast every hour. The project folder is *~/Projects/iotzone/myweather*

## Step 1: Create A New Project

This uses golang module so you can use the folder of your choice. More info here: https://blog.golang.org/using-go-modules


```bash
$ mkdir -p ~/Projects/iotzone/myweather
$ cd ~/Projects/iotzone/myweather
$ go mod init myweather
   > go: creating new go.mod: module myweather
```
Create a file named 'myweather.go' that looks like:
```golang
package main
import "github.com/hspaay/iotzone.golang"
import "fmt"
func main() {
  fmt.Printf("hello, myweather\n")
}
```
Build and run:
```bash
$ go build
$ ./myweather   
  > Hello, myweather
```

A file go.mod contains the module info include dependencies and versions of the dependencies. 
Make sure go.mod and go.sum are added to your version control.

## Step 2. Implement The Publisher
A basic publisher is implemented through only a few functions as shown below.

Change myweather.go to look like this (pseudocode):
```golang
import "github.com/hspaay/iotzone.golang"

const PublisherId = "myweather"
const NodeId = "amsterdam"  
const ConfigLatitude = "latitude"
const ConfigLongitude = "longitude"
const ConfigWeatherServiceUrl = "url"
const OutputTypeForecast = "forecast"
const DefaultWeatherServiceUrl = "http://weatherservice.com/forecast?latitude=${latitude}&longitude=${longitude}"

struct iotzoneConfig {
    Zone: string,
    Publisher: string,
    DiscoveryInterval: int,
    
}
function Initialize(iotzone *iotzone) {

  iotzone.log.warn("Starting myweather")
  // update discovery once a day
  iotzone.SetDefaultDiscoveryInterval(3600*24, this.Discover) 
  // update the forecast once an hour
  iotzone.SetDefaultOutputInterval(3600, this.PollOutputs)
  // The default config handler is good for basic use
  // iotzone.SetConfigHandler(this.ConfigHandler)
  // The node has no control inputs to handle
  // iotzone.SetInputHandler(this.InputHandler)
}

function Terminate(iotzone *iotzone) {
  iotzone.log.warn("Stopping myweather")
}

function Discover(iotzone *iotzone) {
  // Discovery of node. Discovery can be updated any time.
  node = iotzone.UpdateNode(NodeId) // add/update node with ID forecast
  iotzone.SetNodeDefaultConfig(node, ConfigLatitude, DataTypeFloat, "55")
  iotzone.SetNodeDefaultConfig(node, ConfigLongitude, DataTypeFloat, "123")
  iotzone.SetNodeDefaultConfig(node, ConfigWeatherServiceUrl, DataTypeString, DefaultWeatherServiceUrl)
}

function Poll(iotzone *iotzone) {
  node = iotzone.GetNode(NodeId)
  configValues = iotzone.GetNodeConfigValues(node)
  url = configValues[ConfigWeatherServiceUrl] % configValues
  forecastRaw = httplib.get(url)
  if (forecastRaw) {
      forecastObject = JSON.parse(forecastRaw)
      // iotzone implements the convention; Only the output value needs to be set
      // This publishes the forecast on iotzone/myweather/forecast/amsterdam/0/value|latest|history (see convention)
      iotzone.UpdateOutput(node, OutputTypeForecast, "0", forecastObject.value)
  } else {
    iotzone.UpdateOutputError(node, OutputTypeForecast, "0", "Server provided no forecast")
  }
}

func main() {
  myWeather = iotzone.New(DefaultZone, PublisherId)
  myWeather.Start(Initialize, Terminate, Discover, Poll)
}
```

# Step 3. Add Nodes
Create a weather forecase node with configuration for latitude, longitude and location name.
The iotzone library will automatically publish the node with any updates including the configuration.

# Step 4. Add Output Values
Obtain the forcast and add it as an output value to the node.
The iotzone library will automatically publish the output discovery and values.





... work in progress ...
