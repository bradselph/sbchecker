package logger

import (
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func Initialize() {
	Log = logrus.New()
	Log.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	logFileName := "logs/" + time.Now().Format("2006-01-02") + ".txt"
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.WithError(err).Error("Failed to open log file")
	}

	mw := io.MultiWriter(os.Stdout, logFile)
	Log.SetOutput(mw)
}
