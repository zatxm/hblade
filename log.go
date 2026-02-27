package hblade

import "go.uber.org/zap"

var log *zap.Logger

func initLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	return logger
}

func Log() *zap.Logger {
	if log != nil {
		return log
	}
	log = initLogger()
	return log
}

func Logger(l *zap.Logger) {
	log = l
}
