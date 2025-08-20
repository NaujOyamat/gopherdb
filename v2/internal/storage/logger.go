package storage

import "log"

type BadgerLogger struct{}

func (l *BadgerLogger) Errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *BadgerLogger) Warningf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *BadgerLogger) Infof(format string, args ...interface{}) {
	log.Printf(format, args...)
}

func (l *BadgerLogger) Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
