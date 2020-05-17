// Package messenger - Publish and Subscribe to message using the MQTT message bus
package messenger

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/hspaay/iotc.golang/iotc"
	log "github.com/sirupsen/logrus"
)

// ConnectionTimeoutSec constant with connection and reconnection timeouts
const ConnectionTimeoutSec = 20

// TLSPort is the default secure port to connect to mqtt
const TLSPort = 8883

// MqttMessenger that implements IMessenger
type MqttMessenger struct {
	config              *MessengerConfig    // connect information
	pahoClient          pahomqtt.Client     // Paho MQTT Client
	subscriptions       []TopicSubscription // list of TopicSubscription for re-subscribing after reconnect
	Logger              *log.Logger         // Logger provided by user
	isRunning           bool                // listen for messages while running
	tlsVerifyServerCert bool                // verify the server certificate, this requires a Root CA signed cert
	tlsCACertFile       string              // path to CA certificate
	updateMutex         *sync.Mutex         // mutex for async updating of subscriptions
}

// TopicSubscription holds subscriptions to restore after disconnect
type TopicSubscription struct {
	address string
	handler func(address string, publication *iotc.Publication)
	token   pahomqtt.Token // for debugging
	client  *MqttMessenger //
	log     *log.Logger
}

// Connect to the MQTT broker and set the LWT
// If a previous connection exists then it is disconnected first.
// This publishes the LWT on the address baseTopic/deviceID/$state.
// @param lastWillTopic optional last will and testament address for publishing device state on accidental disconnect.
//                       Use "" to ignore LWT feature.
// @param lastWillValue to use as the last will
func (messenger *MqttMessenger) Connect(lastWillAddress string, lastWillValue string) error {
	config := messenger.config

	// close existing connection
	if messenger.pahoClient != nil && messenger.pahoClient.IsConnected() {
		messenger.pahoClient.Disconnect(10 * ConnectionTimeoutSec)
	}

	// set config defaults
	// ClientID defaults to hostname-secondsSinceEpoc
	hostName, _ := os.Hostname()
	if config.ClientID == "" {
		config.ClientID = fmt.Sprintf("%s-%d", hostName, time.Now().Unix())
	}

	// Connect using TLS
	port := config.Port
	if port == 0 {
		port = TLSPort
	}

	brokerURL := fmt.Sprintf("tls://%s:%d/", config.Server, port) // tcp://host:1883 ws://host:1883 tls://host:8883, tcps://awshost:8883/mqtt
	// brokerURL := fmt.Sprintf("tls://mqtt.eclipse.org:8883/")
	opts := pahomqtt.NewClientOptions()
	opts.AddBroker(brokerURL)
	opts.SetClientID(config.ClientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetMaxReconnectInterval(60 * time.Second) // max wait 1 minute for a reconnect
	// Do not use MQTT persistence as not all brokers support it, and it causes problems on the broker if the client ID is
	// randomly generated. CleanSession disables persistence.
	opts.SetCleanSession(true)
	opts.SetKeepAlive(ConnectionTimeoutSec * time.Second) // pings to detect a disconnect. Use same as reconnect interval
	//opts.SetKeepAlive(60) // keepalive causes deadlock in v1.1.0. See github issue #126

	opts.SetOnConnectHandler(func(client pahomqtt.Client) {
		messenger.Logger.Warningf("MqttMessenger.onConnect: Connected to server at %s. Connected=%v. ClientId=%s",
			brokerURL, client.IsConnected(), config.ClientID)
		// Subscribe to addresss already registered by the app on connect or reconnect
		messenger.resubscribe()
	})
	opts.SetConnectionLostHandler(func(client pahomqtt.Client, err error) {
		log.Warningf("MqttMessenger.onConnectionLost: Disconnected from server %s. Error %s, ClientId=%s",
			brokerURL, err, config.ClientID)
	})
	if lastWillAddress != "" {
		//lastWillTopic := fmt.Sprintf("%s/%s/$state", messenger.config.Base, deviceId)
		opts.SetWill(lastWillAddress, lastWillValue, 1, false)
	}
	// Use TLS if a CA certificate is given
	var rootCA *x509.CertPool
	if messenger.tlsCACertFile != "" {
		rootCA = x509.NewCertPool()
		caFile, err := ioutil.ReadFile(messenger.tlsCACertFile)
		if err != nil {
			messenger.Logger.Errorf("MqttMessenger.Connect: Unable to read CA certificate chain: %s", err)
		}
		rootCA.AppendCertsFromPEM([]byte(caFile))
	}
	opts.SetTLSConfig(&tls.Config{
		InsecureSkipVerify: !messenger.tlsVerifyServerCert,
		RootCAs:            rootCA, // include the zcas cert in the host root ca set
		// https://opium.io/blog/mqtt-in-go/
		ServerName: "", // hostname on the server certificate. How to get this?
	})

	messenger.Logger.Infof("MqttMessenger.Connect: Connecting to MQTT server: %s with clientID %s"+
		" AutoReconnect and CleanSession are set.",
		brokerURL, config.ClientID)

	// FIXME: PahoMqtt disconnects when sending a lot of messages, like on startup of some adapters.
	messenger.pahoClient = pahomqtt.NewClient(opts)

	// start listening for messages
	messenger.isRunning = true
	//go messenger.messageChanLoop()

	// Auto reconnect doesn't work for initial attempt: https://github.com/eclipse/paho.mqtt.golang/issues/77
	retryDelaySec := 1
	for {
		token := messenger.pahoClient.Connect()
		token.Wait()
		// Wait to give connection time to settle. Sending a lot of messages causes the connection to fail. Bug?
		time.Sleep(1000 * time.Millisecond)
		err := token.Error()
		if err == nil {
			break
		}

		messenger.Logger.Errorf("MqttMessenger.Connect: Connecting to broker on %s failed: %s. retrying in %d seconds.",
			brokerURL, token.Error(), retryDelaySec)
		time.Sleep(time.Duration(retryDelaySec) * time.Second)
		// slowly increment wait time
		if retryDelaySec < 120 {
			retryDelaySec++
		}
	}
	return nil
}

// Disconnect from the MQTT broker and unsubscribe from all addresss and set
// device state to disconnected
func (messenger *MqttMessenger) Disconnect() {
	messenger.updateMutex.Lock()
	messenger.isRunning = false
	messenger.updateMutex.Unlock()

	if messenger.pahoClient != nil {
		messenger.Logger.Warningf("MqttMessenger.Disconnect: Set state to disconnected and close connection")
		//messenger.publish("$state", "disconnected")
		time.Sleep(time.Second / 10) // Disconnect doesn't seem to wait for all messages. A small delay ahead helps
		messenger.pahoClient.Disconnect(10 * ConnectionTimeoutSec * 1000)
		messenger.pahoClient = nil

		messenger.subscriptions = nil
		//close(messenger.messageChannel)     // end the message handler loop
	}
}

// GetZone returns the zone in which this messenger operates
// This is provided via the messenger config file or defaults to iotc.LocalZoneID
func (messenger *MqttMessenger) GetZone() string {
	zone := messenger.config.Zone
	if zone == "" {
		return iotc.LocalZoneID
	}
	return zone
}

// Publish value using the device address as base
// address to publish on.
// retained to have the broker retain the address value
// payload is converted to string if it isn't a byte array, as Paho doesn't handle int and bool
func (messenger *MqttMessenger) Publish(address string, retained bool, publication *iotc.Publication) error {
	var err error

	//fullTopic := fmt.Sprintf("%s/%s/%s", messenger.config.Base, messenger.deviceId, addressLevels)
	if messenger.pahoClient == nil || !messenger.pahoClient.IsConnected() {
		messenger.Logger.Warnf("MqttMessenger.Publish: Unable to publish. No connection with server.")
		return errors.New("no connection with server")
	}
	payload, err := json.MarshalIndent(publication, " ", " ")
	if err != nil {
		messenger.Logger.Errorf("MqttMessenger.Publish:  Error marshalling publication: %s", err)
		return err
	}
	messenger.Logger.Debugf("MqttMessenger.Publish []byte: address=%s, qos=%d, retained=%v",
		address, messenger.config.PubQos, retained)
	token := messenger.pahoClient.Publish(address, messenger.config.PubQos, retained, payload)

	err = token.Error()
	if err != nil {
		// TODO: confirm that with qos=1 the message is sent after reconnect
		messenger.Logger.Warnf("MqttMessenger.Publish: Error during publish on address %s: %v", address, err)
		//return err
	}
	return err
}

// PublishRaw message
func (messenger *MqttMessenger) PublishRaw(address string, retained bool, message json.RawMessage) error {
	if messenger.pahoClient == nil || !messenger.pahoClient.IsConnected() {
		messenger.Logger.Warnf("MqttMessenger.PublishRaw: Unable to publish. No connection with server.")
		return errors.New("MqttMessenger.PublishRaw: no connection with server")
	}
	// publication := Publication{Message: message}
	// payload, err := json.Marshal(publication)
	token := messenger.pahoClient.Publish(address, messenger.config.PubQos, retained, []byte(message))

	err := token.Error()
	if err != nil {
		// TODO: confirm that with qos=1 the message is sent after reconnect
		messenger.Logger.Warnf("MqttMessenger.PublishRaw: Error during publish on address %s: %v", address, err)
		//return err
	}
	return err
}

// Wrapper for message handling.
// Use a channel to handle the message in a gorouting.
// This fixes a problem with losing context in callbacks. Not sure what is going on though.
func (subscription *TopicSubscription) onMessage(c pahomqtt.Client, msg pahomqtt.Message) {
	// NOTE: Scope in this callback is not always retained. Pipe notifications through a channel and handle in goroutine
	address := msg.Topic()
	rawPayload := msg.Payload()
	var publication iotc.Publication
	err := json.Unmarshal(rawPayload, &publication)
	if err != nil {
		subscription.log.Infof("MqttMessenger.onMessage: Unable to unmarshal payload on address %s. Error: %s", address, err)
		return
	}
	subscription.log.Infof("MqttMessenger.onMessage. address=%s, subscription=%s, retained=%v",
		address, subscription.address, msg.Retained())
	subscription.handler(address, &publication)
	//message := &IncomingMessage{msgTopic, payload, subscription}
	//subscription.client.messageChannel <- message
}

// subscribe to addresss after establishing connection
// The application can already subscribe to addresss before the connection is established. If connection is lost then
// this will re-subscribe to those addresss as PahoMqtt drops the subscriptions after disconnect.
//
func (messenger *MqttMessenger) resubscribe() {
	// prevent simultaneous access to subscriptions
	messenger.updateMutex.Lock()
	defer messenger.updateMutex.Unlock()

	messenger.Logger.Infof("MqttMessenger.resubscribe to %d addresss", len(messenger.subscriptions))
	for _, subscription := range messenger.subscriptions {
		// clear existing subscription
		messenger.pahoClient.Unsubscribe(subscription.address)

		messenger.Logger.Infof("MqttMessenger.resubscribe: address %s", subscription.address)
		// create a new variable to hold the subscription in the closure
		newSubscr := subscription
		token := messenger.pahoClient.Subscribe(newSubscr.address, messenger.config.PubQos, newSubscr.onMessage)
		//token := messenger.pahoClient.Subscribe(newSubscr.address, newSubscr.qos, func (c pahomqtt.Client, msg pahomqtt.Message) {
		//messenger.Logger.Infof("mqtt.resubscribe.onMessage: address %s, subscription %s", msg.Topic(), newSubscr.address)
		//newSubscr.onMessage(c, msg)
		//})
		newSubscr.token = token
	}
	messenger.Logger.Infof("MqttMessenger.resubscribe complete")
}

// Subscribe to a address
// Subscribers are automatically resubscribed after the connection is restored
// If no connection exists, then subscriptions are stored until a connection is established.
// address: address to subscribe to. This can contain wildcards.
// qos: Quality of service for subscription: 0, 1, 2
// handler: callback handler.
func (messenger *MqttMessenger) Subscribe(
	address string, onMessage func(address string, publication *iotc.Publication)) {
	subscription := TopicSubscription{
		address: address,
		handler: onMessage,
		token:   nil,
		client:  messenger,
		log:     messenger.Logger,
	}
	messenger.updateMutex.Lock()
	defer messenger.updateMutex.Unlock()
	messenger.subscriptions = append(messenger.subscriptions, subscription)

	messenger.Logger.Infof("MqttMessenger.Subscribe: address %s, qos %d", address, messenger.config.SubQos)
	//messenger.pahoClient.Subscribe(address, qos, addressSubscription.onMessage) //func(c pahomqtt.Client, msg pahomqtt.Message) {
	if messenger.pahoClient != nil {
		messenger.pahoClient.Subscribe(address, messenger.config.SubQos, subscription.onMessage) //func(c pahomqtt.Client, msg pahomqtt.Message) {
	}
	// return nil
}

// NewMqttMessenger creates a new MQTT messenger instance
func NewMqttMessenger(config *MessengerConfig, logger *log.Logger) *MqttMessenger {
	if logger == nil {
		logger = log.New()
	}
	messenger := &MqttMessenger{
		pahoClient: nil,
		Logger:     logger,
		config:     config,
		//messageChannel: make(chan *IncomingMessage),
		tlsCACertFile:       "/etc/mosquitto/certs/zcas_ca.crt",
		tlsVerifyServerCert: true,
		updateMutex:         &sync.Mutex{},
	}
	return messenger
}
