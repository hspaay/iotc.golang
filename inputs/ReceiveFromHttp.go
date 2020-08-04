// Package inputs with inputs from polling http services
package inputs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// ReceiveFromHTTP with inputs to periodically poll HTTP
// Only a single handler per URL can be used.
type ReceiveFromHTTP struct {
	isRunning        bool              // flag, polling is active
	pollDelay        map[string]int    // seconds until next poll for each input
	pollInterval     int               // default poll interval
	registeredInputs *RegisteredInputs // inputs of this publisher
	subscriptions    map[string]string // http subscriptions of inputs [inputID]source
	updateMutex      *sync.Mutex       // mutex for async updating of inputs
}

// CreateHttpInput creates a new input that periodically polls a URL address. If a login and password
// is provided then it will be used for http basic authentication.
// If an input of the given nodeID, type and instance already exists it will be replaced.
// pollInterval is in seconds
func (rxFromHttp *ReceiveFromHTTP) CreateHttpInput(
	deviceID string, inputType types.InputType, instance string,
	url string, login string, password string, pollInterval int,
	handler func(input *types.InputDiscoveryMessage, sender string, path string)) *types.InputDiscoveryMessage {

	inputID := MakeInputID(deviceID, inputType, instance)
	// create the input then add it to the list of addresses to poll
	input := rxFromHttp.registeredInputs.CreateInputWithSource(deviceID, inputType, instance, url, handler)
	input.Attr[types.NodeAttrURL] = url
	input.Attr[types.NodeAttrPollInterval] = strconv.Itoa(pollInterval)
	input.Attr[types.NodeAttrLoginName] = login
	input.Attr[types.NodeAttrPassword] = password
	rxFromHttp.registeredInputs.UpdateInput(input)

	rxFromHttp.updateMutex.Lock()
	defer rxFromHttp.updateMutex.Unlock()
	rxFromHttp.subscriptions[inputID] = url
	return input
}

// Start polling inputs for changes
func (rxFromHttp *ReceiveFromHTTP) Start() {
	rxFromHttp.isRunning = true
	go rxFromHttp.pollLoop()
}

// Stop polling for inputs
func (rxFromHttp *ReceiveFromHTTP) Stop() {
	rxFromHttp.isRunning = false
}

// DeleteInput deletes the input and stops polling the url
func (rxFromHttp *ReceiveFromHTTP) DeleteInput(inputID string) {
	rxFromHttp.updateMutex.Lock()
	defer rxFromHttp.updateMutex.Unlock()
	existingInput := rxFromHttp.registeredInputs.GetInputByID(inputID)
	if existingInput == nil {
		logrus.Errorf("DeleteInput: input %s not found", inputID)
		return
	}
	delete(rxFromHttp.subscriptions, inputID)
	rxFromHttp.registeredInputs.DeleteInput(inputID)
}

// Send a request to the URL and read the response
// This supports basic authentication
func (rxFromHttp *ReceiveFromHTTP) readInput(input *types.InputDiscoveryMessage) (string, error) {
	var err error
	url := input.Source

	logrus.Debugf("InputFromHTTP.readInput: Reading from URL %s", url)
	startTime := time.Now()
	var req *http.Request
	var resp *http.Response
	var loginName = input.Attr[types.NodeAttrLoginName]
	var password = input.Attr[types.NodeAttrPassword]

	if loginName == "" {
		// No auth
		resp, err = http.Get(url)
	} else {
		// basic auth
		client := &http.Client{}
		req, err = http.NewRequest("GET", url, nil)
		if err == nil {
			req.SetBasicAuth(loginName, password)
			resp, err = client.Do(req)
		}
	}
	// handle failure to load the image
	if err != nil {
		logrus.Errorf("InputFromHTTP.readInput: Error opening URL %s: %s", url, err)
		return "", err
	}
	defer resp.Body.Close()
	// was it a good response?
	if resp.StatusCode > 299 {
		msg := fmt.Sprintf("InputFromHTTP.readInput: Failed opening URL %s: %s", url, resp.Status)
		err := errors.New(msg)
		return "", err
	}
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("InputFromHTTP.readInput: Error reading from %s: %s", url, err)
		return "", err
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime).Round(time.Millisecond)
	// fixme: update as input status and don't modify input directly
	input.Attr["latency"] = duration.String()
	return string(payload), nil
}

// Poll the source of the input and notify subscribers with the result
func (rxFromHttp *ReceiveFromHTTP) pollInputAndNotify(input *types.InputDiscoveryMessage) {
	payload, err := rxFromHttp.readInput(input)
	if err == nil && payload != "" {
		inputID := MakeInputID(input.DeviceID, input.InputType, input.Instance)
		rxFromHttp.registeredInputs.NotifyInputHandler(inputID, "", string(payload))
	}
}

// Poll handler called on a 1 second interval after calling Start()
// Each input has its own interval.
func (rxFromHttp *ReceiveFromHTTP) pollInputs() {
	// Each second check which cameras need to be polled and poll
	for inputID := range rxFromHttp.subscriptions {

		// Each input can have its own poll interval. The fallback value is the global poll interval
		input := rxFromHttp.registeredInputs.GetInputByID(inputID)
		pollDelay, _ := rxFromHttp.pollDelay[inputID]
		if pollDelay <= 0 {
			pollDelay, _ = strconv.Atoi(input.Attr[types.NodeAttrPollInterval])
			logrus.Debugf("readInput: reading input %s at interval of %d seconds", inputID, pollDelay)
			go rxFromHttp.pollInputAndNotify(input)
		}
		rxFromHttp.pollDelay[inputID] = pollDelay - 1
	}
}

// Poll loop,
// Each input has its own interval.
func (rxFromHttp *ReceiveFromHTTP) pollLoop() {

	for rxFromHttp.isRunning {
		time.Sleep(time.Second)
		rxFromHttp.pollInputs()
	}
}

// NewReceiveFromHTTP creates a new instance of HTTP based inputs
// Inputs must be created through CreateInput
func NewReceiveFromHTTP(registeredInputs *RegisteredInputs) *ReceiveFromHTTP {

	httpInput := &ReceiveFromHTTP{
		pollDelay:        make(map[string]int),
		pollInterval:     3600,
		registeredInputs: registeredInputs,
		subscriptions:    make(map[string]string),
		updateMutex:      &sync.Mutex{},
	}
	return httpInput
}
