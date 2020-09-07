package log

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type fLog struct {
	level    Level
	maxSize  int64
	filename string
	osFile   *os.File
}

func NewFlog(l, filename string, maxSize int64) Logger {
	lv := level(l)
	osFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Openfile failed,err:%v,filename:%s", err, filename)
		panic(err)
	}

	return &fLog{
		level:    lv,
		maxSize:  maxSize,
		filename: filename,
		osFile:   osFile,
	}
}

func (f *fLog) Debug(msg ...string) {
	f.write(DebugLevel, msg)
}

func (f *fLog) Info(msg ...string) {
	f.write(InfoLevel, msg)
}

func (f *fLog) Warn(msg ...string) {
	f.write(WarnLevel, msg)
}

func (f *fLog) Error(msg ...string) {
	f.write(ErrorLevel, msg)
}

func (f *fLog) Dpanic(msg ...string) {
	f.write(DPanicLevel, msg)
}

func (f *fLog) Panic(msg ...string) {
	f.write(PanicLevel, msg)
}

func (f *fLog) Fatal(msg ...string) {
	f.write(FatalLevel, msg)
}

func (f *fLog) write(l Level, msg []string) {
	if l >= f.level {
		lSring := l.String()
		_, file, line := caller(3)
		f.compareFileMaxSize()
		t := time.Now().Format("2006-01-02 15:04:05")
		fmt.Fprintf(f.osFile, "%s | %s | %s:%d | %s\n", t, lSring, file, line, strings.Join(msg, ""))
	}
}

func (f *fLog) compareFileMaxSize() {
	fInfo, _ := f.osFile.Stat()
	size := fInfo.Size()
	if size >= f.maxSize {
		f.osFile.Close()
		t := time.Now().Format("20060102150405")
		bakFileName := f.filename + "." + t
		os.Rename(f.filename, bakFileName)
		osFile, _ := os.OpenFile(f.filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		f.osFile = osFile
	}
}
