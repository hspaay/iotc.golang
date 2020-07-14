// Package inputs with inputs from polling http services
package inputs

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// HTTPInputs with inputs to periodically poll
// Only a single handler per URL can be used.
type HTTPInputs struct {
	Inputs       map[string]*HTTPInput // input to poll
	PollInterval int                   // default poll interval
	//
	errorHandler func(url string, err error) // invoke when error reading input
	isRunning    bool                        // flag, polling is active
	logger       logrus.Logger               //
	pollDelay    map[string]int              // seconds until next poll for each input
	updateMutex  *sync.Mutex                 // mutex for async updating of inputs
}

// HTTPInput defines a single input
type HTTPInput struct {
	Enabled  bool   `json:"enabled"`            // polling is enabled
	Interval int    `json:"interval"`           // interval to poll in seconds
	Login    string `json:"login,omitempty"`    // optional basic auth
	Password string `json:"password,omitempty"` // optional
	URL      string `json:"url"`                // input url with http/https to read from
	handler  func(url string, payload []byte, duration time.Duration)
}

// Start polling inputs for changes
func (hil *HTTPInputs) Start() {
	hil.isRunning = true
	go hil.pollLoop()
}

// Stop polling for inputs
func (hil *HTTPInputs) Stop() {
	hil.isRunning = false
}

// AddInput will add the input. If the url exists its input will be replaced
// This supports basic authentication
func (hil *HTTPInputs) AddInput(url string, login string, password string, interval int,
	handler func(url string, payload []byte, duration time.Duration)) {
	input := &HTTPInput{
		Enabled:  true,
		Interval: interval,
		Login:    login,
		Password: password,
		URL:      url,
		handler:  handler,
	}
	hil.updateMutex.Lock()
	defer hil.updateMutex.Unlock()
	hil.Inputs[url] = input
}

// RemoveInput by url
func (hil *HTTPInputs) RemoveInput(url string) {
	hil.updateMutex.Lock()
	defer hil.updateMutex.Unlock()
	delete(hil.Inputs, url)
}

// read the response downloaded from the URL and invoke the input handler
// This supports basic authentication
func (hil *HTTPInputs) readInput(input *HTTPInput) {
	var err error
	url := input.URL

	hil.logger.Debugf("HTTPInputList.read: Reading from URL %s", url)
	startTime := time.Now()
	var req *http.Request
	var resp *http.Response

	if input.Login == "" {
		// No auth
		resp, err = http.Get(url)
	} else {
		// basic auth
		client := &http.Client{}
		req, err = http.NewRequest("GET", url, nil)
		if err == nil {
			req.SetBasicAuth(input.Login, input.Password)
			resp, err = client.Do(req)
		}
	}
	// handle failure to load the image
	if err != nil {
		hil.logger.Errorf("readCameraImage: Error opening URL %s: %s", url, err)
		return
	}
	defer resp.Body.Close()
	// was it a good response?
	if resp.StatusCode > 299 {
		msg := fmt.Sprintf("readCameraImage: Failed opening URL %s: %s", url, resp.Status)
		hil.logger.Errorf(msg)
		return
	}
	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		hil.logger.Errorf("readCameraImage: Error reading camera image from %s: %s", url, err)
		return
	}
	endTime := time.Now()
	duration := endTime.Sub(startTime).Round(time.Millisecond)
	input.handler(url, payload, duration)
}

// Poll handler called on a 1 second interval after calling Start()
// Each input has its own interval.
func (hil *HTTPInputs) poll() {
	// Each second check which cameras need to be polled and poll

	for _, input := range hil.Inputs {

		// Each input can have its own poll interval. The fallback value is the global poll interval
		pollDelay, _ := hil.pollDelay[input.URL]
		if pollDelay <= 0 {
			pollDelay = input.Interval
			hil.logger.Debugf("readInput: reading input from %s at interval of %d seconds", input.URL, input.Interval)
			go hil.readInput(input)
		}
		hil.pollDelay[input.URL] = pollDelay - 1
	}
}

// Poll loop,
// Each input has its own interval.
func (hil *HTTPInputs) pollLoop() {

	for hil.isRunning {
		time.Sleep(time.Second)
		hil.poll()
	}
}

// NewHTTPInputs creates a new instance of HTTP inputs collection
func NewHTTPInputs(errorHandler func(url string, err error)) *HTTPInputs {
	hil := &HTTPInputs{}
	hil.errorHandler = errorHandler
	hil.logger = logrus.Logger{}
	return hil
}
