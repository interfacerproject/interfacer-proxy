// SPDX-FileCopyrightText: 2023 Dyne.org foundation
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package logger

import (
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

var Log = logrus.New()

func InitLog(logPath string) {
	Log.SetFormatter(&logrus.JSONFormatter{})

	Log.Out = os.Stdout

	file, err := os.OpenFile(filepath.Join(logPath, "proxy.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.WithFields(logrus.Fields{"err": err.Error()}).
			Info("Failed to log to file, using default stderr")
	} else {
		Log.Out = file
	}
}
