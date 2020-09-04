package log

import (
	"fmt"
	"strings"
	"time"
)

type pLog struct {
	level Level
}

func NewPlog(l string) Logger {
	lv := level(l)
	return &pLog{level: lv}
}

func (p *pLog) Debug(msg ...string) {
	p.write(DebugLevel, msg)
}

func (p *pLog) Info(msg ...string) {
	p.write(InfoLevel, msg)
}

func (p *pLog) Warn(msg ...string) {
	p.write(WarnLevel, msg)
}

func (p *pLog) Error(msg ...string) {
	p.write(ErrorLevel, msg)
}

func (p *pLog) Dpanic(msg ...string) {
	p.write(DPanicLevel, msg)
}

func (p *pLog) Panic(msg ...string) {
	p.write(PanicLevel, msg)
}

func (p *pLog) Fatal(msg ...string) {
	p.write(FatalLevel, msg)
}

func (p *pLog) write(l Level, msg []string) {
	if l >= p.level {
		lSring := l.String()
		_, file, line, _ := caller(2)
		t := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("%s | %s | %s | %d | %s\n", t, lSring, file, line, strings.Join(msg, ""))
	}
}
