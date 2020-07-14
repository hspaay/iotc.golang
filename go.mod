module github.com/iotdomain/iotdomain-go

go 1.13

// replace github.com/iotdomain/iotdomain-go => ../iotdomain-go

require (
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/google/go-cmp v0.5.0 // indirect
	github.com/sirupsen/logrus v1.6.0
	github.com/square/go-jose v2.5.1+incompatible
	github.com/stretchr/testify v1.6.0
	golang.org/x/net v0.0.0-20200528225125-3c3fba18258b // indirect
	gopkg.in/square/go-jose.v2 v2.5.1
	gopkg.in/yaml.v2 v2.3.0
)

// replace github.com/iotdomain/iotdomain-go/standard => ../standard
