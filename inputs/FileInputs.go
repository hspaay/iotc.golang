// Package inputs for using a file as input
package inputs

import (
	"fmt"
	"os/user"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
)

// FileInputs contains the file watches of inputs that listens for changes in files
// This supports multiple watchers for the same file.
type FileInputs struct {
	watcher       *fsnotify.Watcher
	isRunning     bool
	subscriptions map[string][]func(string) // list of handler addresses
	updateMutex   *sync.Mutex               // mutex for async updating of inputs
}

// Start listening for file changes
func (fil *FileInputs) Start() {
	fil.isRunning = true
	fil.watcherLoop()
}

// Stop listening for file changes
func (fil *FileInputs) Stop() {
	fil.isRunning = false
}

// SubscribeToFile listens for changes in file and invokes the handler
// This supports multiple subscriptions
func (fil *FileInputs) SubscribeToFile(fileName string, handler func(fileName string)) string {
	fil.updateMutex.Lock()
	defer fil.updateMutex.Unlock()
	path := fil.watchFile(fileName)
	if path != "" {
		sub := fil.subscriptions[path]
		if sub == nil {
			sub = make([]func(string), 0)
		}
		sub = append(sub, handler)
		fil.subscriptions[path] = sub
	}
	return path
}

// UnsubscribeFromFile removes subscription to changes in file
func (fil *FileInputs) UnsubscribeFromFile(fileName string, handler func(filename string)) {
	fil.updateMutex.Lock()
	defer fil.updateMutex.Unlock()
	subList := fil.subscriptions[fileName]
	if subList != nil {
		// compare function addresses by converting them to string
		handlerStr := fmt.Sprintf("%p", handler)
		for index, subHandler := range subList {
			subHandlerStr := fmt.Sprintf("%p", subHandler)
			if subHandlerStr == handlerStr {
				copy(subList[index:], subList[index+1:])
				// fil.subscriptions[fileName] = subList
				break
			}
		}
		if len(subList) == 0 {
			fil.watcher.Remove(fileName)
		}
	}
}

// Handle update to file. Invoked by a file watcher
func (fil *FileInputs) handleFileInput(fileName string) {
	fil.updateMutex.Lock()
	subList := fil.subscriptions[fileName]
	fil.updateMutex.Unlock()
	if subList != nil {
		for _, handler := range subList {
			handler(fileName)
		}
	}

}

// SubscribeToFile subscribes to changes in file
// If filename starts with ~ then the current user home directory is used
// If filename contains . or .. then it will be cleaned
// Returns the actual pathname or "" if the path is invalid
func (fil *FileInputs) watchFile(fileName string) string {
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
	fil.watcher.Add(newPath)
	return newPath
}

// loop watching for changes to file
func (fil *FileInputs) watcherLoop() {
	for fil.isRunning {
		select {
		case event, ok := <-fil.watcher.Events:
			if !ok {
				return
			}
			logrus.Infof("watcherLoop:event: %s", event)
			if event.Op&fsnotify.Write == fsnotify.Write {
				logrus.Infof("watcherLoop:modified file %s", event.Name)
				fil.handleFileInput(event.Name)
			}
		case err, ok := <-fil.watcher.Errors:
			if !ok {
				return
			}
			logrus.Warnf("watcherLoop:error: %s", err)
		}
	}
}

// NewFileInputs creates a new file watcher input list
func NewFileInputs() *FileInputs {
	watcher, _ := fsnotify.NewWatcher()
	fil := &FileInputs{
		watcher:       watcher,
		subscriptions: make(map[string][]func(string)),
		updateMutex:   &sync.Mutex{},
	}
	return fil
}
