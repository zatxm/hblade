package hblade

import (
	"github.com/zatxm/hblade/log"
)

var Log log.Logger = log.NewPlog("debug")

func NewPlog(l string) log.Logger {
	Log = log.NewPlog(l)
	return Log
}

func NewFlog(l, filename string, maxSize int64) log.Logger {
	Log = log.NewFlog(l, filename, maxSize)
	return Log
}
