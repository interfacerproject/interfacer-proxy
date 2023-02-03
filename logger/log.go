package logger

import (
    "os"
    "path/filepath"
    "github.com/sirupsen/logrus"
)

var Log = logrus.New()

func InitLog(logPath string) {
    Log.SetFormatter(&logrus.JSONFormatter{})

    Log.Out = os.Stdout

    if file, err := os.OpenFile(
                filepath.Join(logPath, "proxy.log"),
                os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666);
            err != nil {
        Log.WithFields(logrus.Fields{
            "err": err.Error(),
        }).Info("Failed to log to file, using default stderr")
    } else {
        Log.Out = file
    }
}
