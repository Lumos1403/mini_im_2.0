package logger

import "go.uber.org/zap"

var global = zap.NewNop()

func Init(mode string) {
	var log *zap.Logger
	var err error

	if mode == "release" {
		log, err = zap.NewProduction()
	} else {
		log, err = zap.NewDevelopment()
	}
	if err != nil {
		global = zap.NewNop()
		return
	}

	global = log
}

func L() *zap.Logger {
	return global
}

func Sync() {
	_ = global.Sync()
}
