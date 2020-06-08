module github.com/hspaay/iotc.golang

go 1.13

// replace github.com/hspaay/iotc.golang => ../iotc.golang

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/square/go-jose v2.5.1+incompatible
	github.com/stretchr/testify v1.6.0
	golang.org/x/net v0.0.0-20200528225125-3c3fba18258b // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.3.0
)

// replace github.com/hspaay/iotc.golang/standard => ../standard
