// Package inputs for using a file as input
package inputs

import (
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// ReceiveFromFiles receives updates from file watchers of inputs that listens for changes in files
// This supports multiple watchers for the same file.
type ReceiveFromFiles struct {
	isRunning        bool
	registeredInputs *RegisteredInputs // registered inputs of this publisher
	updateMutex      *sync.Mutex       // mutex for async updating of inputs
	watcher          *fsnotify.Watcher // the watcher to monitor files
}

// CreateInput creates an input that triggers when a file is written to and invokes the given handler.
// The file must exist when creating the input (the file watcher requires it).
// If the input already exists, the existing input is returned.
func (iffile *ReceiveFromFiles) CreateInput(
	deviceID string, inputType types.InputType, instance string,
	path string, handler func(input *types.InputDiscoveryMessage, sender string, path string)) *types.InputDiscoveryMessage {

	iffile.updateMutex.Lock()
	defer iffile.updateMutex.Unlock()

	// prevent input to be added multiple times
	existingInput := iffile.registeredInputs.GetInputByDevice(deviceID, inputType, instance)
	if existingInput != nil {
		logrus.Errorf("AddInput: Input %s already exists. Ignored.", existingInput.InputID)
		return existingInput
	}
	// use full path as this is what the watcher uses in file change notification
	fullPath := iffile.watchFile(path)

	if fullPath == "" {
		inputID := MakeInputID(deviceID, inputType, instance)
		logrus.Errorf("AddInput: Source path '%s' for input '%s' is invalid. Ignored.", path, inputID)
		return nil
	}
	input := iffile.registeredInputs.CreateInputWithSource(
		deviceID, inputType, instance, fullPath, handler)

	return input
}

// DeleteInput deletes the input and unsubscribes from the file watcher
func (iffile *ReceiveFromFiles) DeleteInput(deviceID string, inputType types.InputType, instance string) {
	iffile.updateMutex.Lock()
	iffile.updateMutex.Unlock()

	inputID := MakeInputID(deviceID, inputType, instance)
	existingInput := iffile.registeredInputs.GetInputByID(inputID)
	if existingInput == nil {
		logrus.Errorf("DeleteInput: input %s not found", inputID)
		return
	}
	fullPath := existingInput.Source
	if fullPath == "" {
		// not file path, just delete the input
		logrus.Errorf("DeleteInput: input %s does not have a source path", inputID)
		iffile.registeredInputs.DeleteInput(inputID)
		return
	}

	// now we have a valid path, remove it from the watcher
	err := iffile.watcher.Remove(fullPath)
	if err != nil {
		logrus.Errorf("DeleteInput: error removing full path %s from file watcher: %s", fullPath, err)
	}
	iffile.registeredInputs.DeleteInput(inputID)
}

// Start listening for file changes
func (iffile *ReceiveFromFiles) Start() {
	iffile.isRunning = true
	go iffile.watcherLoop()
}

// Stop listening for file changes
func (iffile *ReceiveFromFiles) Stop() {
	iffile.isRunning = false
}

// Invoked by a file watcher when a file changes
// This looks up the corresponding input(s) and notifies their subscriber
func (iffile *ReceiveFromFiles) onFileWatcherEvent(fullPath string) {
	sourceInputs := iffile.registeredInputs.GetInputsWithSource(fullPath)
	for _, input := range sourceInputs {
		inputID := MakeInputID(input.DeviceID, input.InputType, input.Instance)
		iffile.registeredInputs.NotifyInputHandler(inputID, "", fullPath)
	}
}

// SubscribeToFile subscribes to changes in file
// If filename starts with ~ then the current user home directory is used
// If filename contains . or .. then it will be cleaned
// Returns the actual pathname or "" if the path is invalid
func (iffile *ReceiveFromFiles) watchFile(fileName string) string {
	newPath := fileName
	var err error
	// Replace ~ prefix with the user home directory
	if strings.HasPrefix(fileName, "~") {
		currentUser, _ := user.Current()
		newPath = filepath.Join(currentUser.HomeDir, fileName[1:])
		newPath = filepath.Clean(newPath)
	} else if strings.HasPrefix(fileName, ".") {
		newPath, err = filepath.Abs(fileName)
	}
	err = iffile.watcher.Add(newPath)
	if err != nil {
		logrus.Errorf("SubscribeToFile: File ignored. Unable to watch file %s: %s", newPath, err)
		return ""
	}
	return newPath
}

// loop watching for writing to file
func (iffile *ReceiveFromFiles) watcherLoop() {
	for iffile.isRunning {
		select {
		case event, ok := <-iffile.watcher.Events:
			if !ok {
				return
			}
			logrus.Infof("watcherLoop:event: %s", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				logrus.Infof("watcherLoop:modified file %s", event.Name)
				iffile.onFileWatcherEvent(event.Name)
			}
		case err, ok := <-iffile.watcher.Errors:
			if !ok {
				return
			}
			logrus.Warnf("watcherLoop:error: %s", err)
		}
	}
}

// NewReceiveFromFiles creates a new file watcher input list
func NewReceiveFromFiles(regInputs *RegisteredInputs) *ReceiveFromFiles {
	watcher, _ := fsnotify.NewWatcher()
	fil := &ReceiveFromFiles{
		registeredInputs: regInputs,
		updateMutex:      &sync.Mutex{},
		watcher:          watcher,
	}
	return fil
}
