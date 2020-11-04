package config

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	FileLogger *log.Logger
}

func Logs() *Logger {
	if flogger == nil {
		flogger = log.New(LogFile(), fmt.Sprintf("%s ", logFile), log.Ldate)
	}
	return &Logger{
		FileLogger: flogger,
	}
}

func LogFile() *os.File {
	year, month, day := time.Now().Local().Date()
	file, err := os.OpenFile(fmt.Sprintf("%s_%d-%s-%d.log", Cfg.LogFile, year, month.String(), day), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	CheckError(err)
	return file
}

func (l *Logger) Info(info string) {
	l.FileLogger.Printf("Info: %s\n", info)
	log.Printf("Info: %s\n", info)
}

func (l *Logger) Error(err string) {
	l.FileLogger.Printf("Error: %s\n", err)
	log.Printf("Error: %s\n", err)
}

func (l *Logger) Fatal(fatal string) {
	l.FileLogger.Printf("Fatal: %s\n", fatal)
	log.Fatalf("Fatal: %s\n", fatal)
}

var logFile string = fmt.Sprintf("ChatServer_%d", time.Now().Unix())
var flogger *log.Logger
