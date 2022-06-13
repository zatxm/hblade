package hblade

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const LogKey = "hblade_log_key"

var Log *zap.Logger = initLog()

func initLog() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()
	return logger
}

func LogWithCtr(c *Context) *zap.Logger {
	c.SetKey(LogKey, Log)
	signId := uuid.New().String()
	Log = Log.With(zap.String("ReqId", signId))
	return Log
}

func LogReleaseCtr(c *Context) *zap.Logger {
	if c == nil {
		return Log
	}
	l, ok := c.GetKey(LogKey)
	if ok {
		return l.(*zap.Logger)
	}
	return Log
}
