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

// InputFromFiles contains the file watches of inputs that listens for changes in files
// This supports multiple watchers for the same file.
type InputFromFiles struct {
	isRunning        bool
	registeredInputs *RegisteredInputs   // registered inputs of this publisher
	updateMutex      *sync.Mutex         // mutex for async updating of inputs
	watcher          *fsnotify.Watcher   // the watcher to monitor files
	sourceToInputMap map[string][]string // source to inputAddress map (1:N)
}

// CreateInput creates an input that triggers on file or folder changes and invokes the given handler.
// This returns the input address
func (iffile *InputFromFiles) CreateInput(
	nodeID string, inputType types.InputType, instance string,
	path string, handler func(inputAddress string, sender string, path string)) string {

	iffile.updateMutex.Lock()
	defer iffile.updateMutex.Unlock()

	// prevent input to be added multiple times
	existingInput := iffile.registeredInputs.GetInput(nodeID, inputType, instance)
	if existingInput != nil {
		logrus.Errorf("AddInput: Input %s already exists. Ignored.", existingInput.Address)
		return existingInput.Address
	}
	// use full path as this is what the watcher uses in file change notification
	fullPath := iffile.watchFile(path)

	if fullPath == "" {
		inputAddress := MakeInputDiscoveryAddress(
			iffile.registeredInputs.domain, iffile.registeredInputs.publisherID, nodeID, inputType, instance)
		logrus.Errorf("AddInput: Source path '%s' for input '%s' is invalid. Ignored.", path, inputAddress)
		return ""
	}
	input := iffile.registeredInputs.CreateInputWithSource(
		nodeID, inputType, instance, fullPath, handler)

	// multiple inputs can use the same source so append this input's address
	inputAddresses := iffile.sourceToInputMap[fullPath]
	if inputAddresses == nil {
		inputAddresses = make([]string, 0)
	}
	inputAddresses = append(inputAddresses, input.Address)
	iffile.sourceToInputMap[fullPath] = inputAddresses
	return input.Address
}

// DeleteInput deletes the input and unsubscribes from the file watcher
func (iffile *InputFromFiles) DeleteInput(nodeID string, inputType types.InputType, instance string) {
	iffile.updateMutex.Lock()
	iffile.updateMutex.Unlock()

	existingInput := iffile.registeredInputs.GetInput(nodeID, inputType, instance)
	if existingInput == nil {
		logrus.Errorf("DeleteInput: input for node %s, type %s, instance %s not found", nodeID, inputType, instance)
		return
	}
	fullPath := existingInput.Source
	if fullPath == "" {
		logrus.Errorf("DeleteInput: input %s does not have a source path", existingInput.Address)
		return
	}
	// now we have a valid path, remove it
	err := iffile.watcher.Remove(fullPath)
	if err != nil {
		logrus.Errorf("DeleteInput: error removing full path %s from file watcher: %s", fullPath, err)
	}

	// cleanup the inputs
	inputAddresses := iffile.sourceToInputMap[fullPath]
	if inputAddresses == nil {
		logrus.Errorf("DeleteInput: input %s with source path %s is not a file input", existingInput.Address, fullPath)
		return
	}
	// delete the input from the subscription
	for index, a := range inputAddresses {
		if a == existingInput.Address {
			if index == len(inputAddresses)-1 {
				inputAddresses = inputAddresses[:index]
			} else {
				inputAddresses = append(inputAddresses[:index], inputAddresses[index+1:]...)
			}
			iffile.sourceToInputMap[fullPath] = inputAddresses
		}
	}
	iffile.registeredInputs.DeleteInput(nodeID, inputType, instance)
}

// Start listening for file changes
func (iffile *InputFromFiles) Start() {
	iffile.isRunning = true
	go iffile.watcherLoop()
}

// Stop listening for file changes
func (iffile *InputFromFiles) Stop() {
	iffile.isRunning = false
}

// Invoked by a file watcher when a file changes
// This looks up the corresponding input(s) and notifies their subscriber
func (iffile *InputFromFiles) onFileWatcherEvent(fullPath string) {
	iffile.updateMutex.Lock()
	inputAddresses := iffile.sourceToInputMap[fullPath]
	iffile.updateMutex.Unlock()
	if inputAddresses != nil {
		for _, inputAddress := range inputAddresses {
			iffile.registeredInputs.NotifyInputHandler(inputAddress, "", fullPath)
		}
	}
}

// SubscribeToFile subscribes to changes in file
// If filename starts with ~ then the current user home directory is used
// If filename contains . or .. then it will be cleaned
// Returns the actual pathname or "" if the path is invalid
func (iffile *InputFromFiles) watchFile(fileName string) string {
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
	if err != nil {
		logrus.Errorf("SubscribeToFile: File ignored. Unable to watch file %s: %s", fileName, err)
		return ""
	}
	iffile.watcher.Add(newPath)
	return newPath
}

// loop watching for changes to file
func (iffile *InputFromFiles) watcherLoop() {
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

// NewInputFromFiles creates a new file watcher input list
func NewInputFromFiles() *InputFromFiles {
	watcher, _ := fsnotify.NewWatcher()
	fil := &InputFromFiles{
		watcher:          watcher,
		sourceToInputMap: make(map[string][]string),
		updateMutex:      &sync.Mutex{},
	}
	return fil
}
