package hblade

import (
	"sync/atomic"

	"go.uber.org/zap"
)

// Lock-free so that hot paths calling Log() don't contend on a shared mutex.
var logger atomic.Pointer[zap.Logger]

func Log() *zap.Logger {
	if l := logger.Load(); l != nil {
		return l
	}
	l, _ := zap.NewDevelopment()
	if logger.CompareAndSwap(nil, l) {
		return l
	}
	return logger.Load()
}

func Logger(l *zap.Logger) {
	logger.Store(l)
}
