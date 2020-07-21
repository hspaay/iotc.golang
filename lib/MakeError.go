// Package lib with common functions
package lib

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
)

// MakeErrorf logs an warning and return an error with the given text
func MakeErrorf(formatStr string, a ...interface{}) error {
	errText := fmt.Sprintf(formatStr, a...)
	logrus.Warning(errText)
	return errors.New(errText)
}
