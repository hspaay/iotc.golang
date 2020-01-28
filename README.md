# myzone.golang
myzone.golang is the core library in the Golang language *for developers of MyZone publishers*. If you just want to run a publisher, then download a binary of that publisher and run it according its instructions. 

myzone is a convention for publishing information on a message bus. This library is part of the reference implementation.
The convention can be found at: https://github.com/hspaay/myzone.convention

## This provides
* systemd launcher for linux 
* Auto publish discovery when nodes and configuration are updated 
* Auto publish updates to output values 
* Management of nodes, inputs and outputs
* Functions for publishing output and discovery messages
* Functions for handling input and configure messages
* MyZone convention data types in Golang

## Getting Started (Linux)

Below to build your own publisher in golang as a go module. This lets you pick your own project folder.
Lets say you are building multiple publishers and have a dedicated myzone project folder in: ~/Projects/myzone.
Now lets build a new publisher that publishes the local weather forecast. Lets call the publisher 'myweather' and build it from scratch.

1. Install golang 
  If you are new to golang, check out their website: https://golang.org/doc/install
  This depends on your linux distribution. On Ubuntu:
  ```bash
  $ sudo apt install golang
  ```
  
2. Create a project folder using go modules.
  Go modules lets you use your own folder structure. More info here: https://blog.golang.org/using-go-modules
```bash
$ mkdir -p ~/Projects/myweather
$ cd ~/Projects/myweather
$ go mod init myweather
   > go: creating new go.mod: module myweather
```
Create a file named 'myweather.go' that looks like:
```golang
package main
import "github.com/hspaay/myzone.golang"
import "fmt"
func main() {
  fmt.Printf("hello, myweather\n")
}
```
4. Build
```bash
$ go build
```
A file go.mod contains the module info include dependencies and versions of the dependencies. 
Make sure go.mod and go.sum are added to your version control.

5. Initialize the publisher
The publisher is hooked into the MyZone library. MyZone handles the launching arguments, sets-up logging, and loads the configuration. Change the example to look like this:
```golang
import "github.com/hspaay/myzone.golang"

class MyWeather extends MyZonePublisher {
}

func main() {
  myzone.run(MyWeather)
}
```

... work in progress ...
