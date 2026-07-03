package hblade

import (
	"sync"

	"go.uber.org/zap"
)

var (
	logMu sync.RWMutex
	log   *zap.Logger
)

func initLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func Log() *zap.Logger {
	logMu.RLock()
	logger := log
	logMu.RUnlock()
	if logger != nil {
		return logger
	}

	logMu.Lock()
	defer logMu.Unlock()
	if log == nil {
		log = initLogger()
	}
	return log
}

func Logger(l *zap.Logger) {
	logMu.Lock()
	defer logMu.Unlock()
	log = l
}
