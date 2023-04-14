package logging

import (
	"fmt"

	"go.uber.org/zap"
)

var badgerCallerSkip = 2 //nolint:gochecknoglobals

type badger interface {
	Errorf(string, ...interface{})
	Warningf(string, ...interface{})
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
}

type BadgerLogger struct {
	log *zap.Logger
}

func NewBadgerLogger(log *zap.Logger) *BadgerLogger {
	return &BadgerLogger{
		log: log.WithOptions(zap.AddCallerSkip(badgerCallerSkip)).With(FieldPkg("badger")),
	}
}

func (l *BadgerLogger) Errorf(s string, i ...interface{}) {
	l.log.Error(fmt.Sprintf(s, i...))
}

func (l *BadgerLogger) Warningf(s string, i ...interface{}) {
	l.log.Warn(fmt.Sprintf(s, i...))
}

func (l *BadgerLogger) Infof(s string, i ...interface{}) {
	l.log.Info(fmt.Sprintf(s, i...))
}

func (l *BadgerLogger) Debugf(s string, i ...interface{}) {
	l.log.Debug(fmt.Sprintf(s, i...))
}

var _ badger = &BadgerLogger{}
