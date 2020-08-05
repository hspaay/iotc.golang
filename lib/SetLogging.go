package lib

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// SetLogging sets the logging level and output file.
// Intended for setting logging from configuration
//  levelName is the requested logging level: error, warning, info, debug
//  filename is the output log file full name including path, use "" for stderr
func SetLogging(levelName string, filename string) error {
	loggingLevel := logrus.DebugLevel
	var err error

	if levelName != "" {
		switch strings.ToLower(levelName) {
		case "error":
			loggingLevel = logrus.ErrorLevel
		case "warn":
		case "warning":
			loggingLevel = logrus.WarnLevel
		case "info":
			loggingLevel = logrus.InfoLevel
		case "debug":
			loggingLevel = logrus.DebugLevel
		}
	}
	logOut := os.Stderr
	if filename != "" {
		logFileHandle, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
		if err != nil {
			err = MakeErrorf("Publisher.SetLogging: Unable to open logfile: %s", err)
		} else {
			logrus.Warnf("Publisher.SetLogging: Send logging output to %s", filename)
			logOut = logFileHandle
		}
	}

	logrus.SetFormatter(
		&logrus.TextFormatter{
			// LogFormat: "",
			// DisableColors:   true,
			// DisableLevelTruncation: true,
			// PadLevelText:    true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			FullTimestamp:   true,
			// ForceFormatting: true,
		})
	logrus.SetOutput(logOut)
	logrus.SetLevel(loggingLevel)

	logrus.SetReportCaller(false) // publisher logging includes caller and file:line#
	return err
}
