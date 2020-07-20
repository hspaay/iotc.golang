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

// InputFromHTTP with inputs to periodically poll HTTP
// Only a single handler per URL can be used.
type InputFromHTTP struct {
	isRunning        bool              // flag, polling is active
	pollDelay        map[string]int    // seconds until next poll for each input
	pollInterval     int               // default poll interval
	registeredInputs *RegisteredInputs // inputs of this publisher
	subscriptions    map[string]string // http subscriptions of inputs [inputAddr]source
	updateMutex      *sync.Mutex       // mutex for async updating of inputs
}

// CreateInput creates a new input that periodically polls a URL address. If a login and password
// is provided then it will be used for http basic authentication.
// If an input of the given nodeID, type and instance already exists it will be replaced.
// pollInterval is in seconds
// This returns the discovery address of the new input
func (ifhttp *InputFromHTTP) CreateInput(
	nodeID string, inputType types.InputType, instance string,
	url string, login string, password string, pollInterval int,
	handler func(inputAddress string, sender string, path string)) string {

	// create the input then add it to the list of addresses to poll
	input := ifhttp.registeredInputs.CreateInputWithSource(nodeID, inputType, instance, url, handler)
	input.Attr[types.NodeAttrURL] = url
	input.Attr[types.NodeAttrPollInterval] = strconv.Itoa(pollInterval)
	input.Attr[types.NodeAttrLoginName] = login
	input.Attr[types.NodeAttrPassword] = password
	ifhttp.registeredInputs.UpdateInput(input)

	ifhttp.updateMutex.Lock()
	defer ifhttp.updateMutex.Unlock()
	ifhttp.subscriptions[input.Address] = url
	return input.Address
}

// Start polling inputs for changes
func (ifhttp *InputFromHTTP) Start() {
	ifhttp.isRunning = true
	go ifhttp.pollLoop()
}

// Stop polling for inputs
func (ifhttp *InputFromHTTP) Stop() {
	ifhttp.isRunning = false
}

// DeleteInput deletes the input and stops polling the url
func (ifhttp *InputFromHTTP) DeleteInput(nodeID string, inputType types.InputType, instance string) {
	ifhttp.updateMutex.Lock()
	defer ifhttp.updateMutex.Unlock()
	existingInput := ifhttp.registeredInputs.GetInput(nodeID, inputType, instance)
	if existingInput == nil {
		logrus.Errorf("DeleteInput: input for node %s, type %s, instance %s not found", nodeID, inputType, instance)
		return
	}
	delete(ifhttp.subscriptions, existingInput.Address)
	ifhttp.registeredInputs.DeleteInput(nodeID, inputType, instance)
}

// Send a request to the URL and read the response
// This supports basic authentication
func (ifhttp *InputFromHTTP) readInput(input *types.InputDiscoveryMessage) (string, error) {
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
func (ifhttp *InputFromHTTP) pollInputAndNotify(input *types.InputDiscoveryMessage) {
	payload, err := ifhttp.readInput(input)
	if err == nil && payload != "" {
		ifhttp.registeredInputs.NotifyInputHandler(input.Address, "", string(payload))
	}
}

// Poll handler called on a 1 second interval after calling Start()
// Each input has its own interval.
func (ifhttp *InputFromHTTP) pollInputs() {
	// Each second check which cameras need to be polled and poll

	for inputAddr := range ifhttp.subscriptions {

		// Each input can have its own poll interval. The fallback value is the global poll interval
		input := ifhttp.registeredInputs.GetInputByAddress(inputAddr)
		pollDelay, _ := ifhttp.pollDelay[inputAddr]
		if pollDelay <= 0 {
			pollDelay, _ = strconv.Atoi(input.Attr[types.NodeAttrPollInterval])
			logrus.Debugf("readInput: reading input %s at interval of %d seconds", inputAddr, pollDelay)
			go ifhttp.pollInputAndNotify(input)
		}
		ifhttp.pollDelay[inputAddr] = pollDelay - 1
	}
}

// Poll loop,
// Each input has its own interval.
func (ifhttp *InputFromHTTP) pollLoop() {

	for ifhttp.isRunning {
		time.Sleep(time.Second)
		ifhttp.pollInputs()
	}
}

// NewInputFromHTTP creates a new instance of HTTP based inputs
// Inputs must be created through CreateInput
func NewInputFromHTTP(registeredInputs *RegisteredInputs) *InputFromHTTP {

	httpInput := &InputFromHTTP{
		pollDelay:        make(map[string]int),
		pollInterval:     3600,
		registeredInputs: registeredInputs,
		subscriptions:    make(map[string]string),
		updateMutex:      &sync.Mutex{},
	}
	return httpInput
}
