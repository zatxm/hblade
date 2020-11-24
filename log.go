package hblade

import (
	"go.uber.org/zap"
)

var Log *zap.Logger = initLog()

func initLog() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	return logger
}
