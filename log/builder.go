package log

import (
	"context"
	"os"
)

type Builder interface {
	Opt() Option
	With(...Field) Builder
	Debugf(format string, v ...interface{})
	Debug(message string)
	Infof(format string, v ...interface{})
	Info(message string)
	Warnf(format string, v ...interface{})
	Warn(message string)
	Errorf(format string, v ...interface{})
	Error(message string)
	Fatalf(format string, v ...interface{})
	Fatal(message string)
	Sync() error
}

type Field struct {
	Key   string
	Value interface{}
}

type loggerKey struct{}

// WithContext Adds fields.
func WithContext(ctx context.Context, log Builder) context.Context {
	return context.WithValue(ctx, loggerKey{}, log)
}

// FromContext Gets the log from context.
func FromContext(ctx context.Context) Builder {
	l, ok := ctx.Value(loggerKey{}).(Builder)
	if !ok {
		return NoLogger{}
	}
	return l
}

type NoLogger struct {
}

func (n NoLogger) Opt() Option {
	return Option{}
}

func (n NoLogger) With(...Field) Builder {
	return n
}

func (n NoLogger) Debugf(string, ...interface{}) {
}

func (n NoLogger) Debug(string) {
}

func (n NoLogger) Infof(string, ...interface{}) {
}

func (n NoLogger) Info(string) {
}

func (n NoLogger) Warnf(string, ...interface{}) {
}

func (n NoLogger) Warn(string) {
}

func (n NoLogger) Errorf(string, ...interface{}) {
}

func (n NoLogger) Error(string) {
}

func (n NoLogger) Fatalf(string, ...interface{}) {
	os.Exit(1)
}

func (n NoLogger) Fatal(string) {
	os.Exit(1)
}

func (n NoLogger) Sync() error {
	return nil
}
