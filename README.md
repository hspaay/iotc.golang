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

1. Install golang 
  If you are new to golang, check out their website: https://golang.org/doc/install
  This depends on your linux distribution. On Ubuntu:
  ```bash
  $ sudo apt install golang
  ```
  
2. Create a project folder, lets call it mysensor 
```bash
$ mkdir -p ~/go/src/mysensor
$ cd ~/go/src/mysensor
```  
  If you use a different project root folder then set GOPATH to the project home folder. 
  See also https://github.com/golang/go/wiki/SettingGOPATH
  In Go 1.13 and newer use "go env":
```bash
$ go env -w GOPATH=~/Projects
```  

3. Start with hello mysensor
  Create a file named ${GOPATH}/src/mysensor/mysensor.go that looks like:
```golang
package main
import "fmt"
func main() {
  fmt.Printf("hello, mysensor\n")
}
```

4. Add the myzone.golang package as a dependency (using dep)
```bash
$ dep ensure -add github.com/hspaay/myzone.golang
```  
