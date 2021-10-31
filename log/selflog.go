package log

import (
	"io"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/crochee/lirity/e"
)

const (
	DefaultLogSizeM int = 20
	DefaultMaxZip   int = 50
	MaxLogDays      int = 30
)

type Option struct {
	Path       string
	Level      string
	Skip       int
	JSONEnable bool
	SetWriter  func(string) io.Writer
}

// NewLogger 初始化日志对象
//
// @param: path 日志路径
// @param: level 日志等级
func NewLogger(opts ...func(*Option)) *Logger {
	l := &Logger{
		Option: Option{
			Path:       "",
			Level:      "INFO",
			Skip:       1,
			JSONEnable: true,
			SetWriter:  SetLoggerWriter,
		},
	}
	for _, opt := range opts {
		opt(&l.Option)
	}
	l.Level = strings.ToUpper(l.Level)
	var encode encoder = zapcore.NewConsoleEncoder
	if l.Option.JSONEnable && l.Option.Path != "" {
		encode = zapcore.NewJSONEncoder
	}
	l.Logger = newZap(l.Option.Level, encode, l.Option.Skip, l.SetWriter(l.Option.Path))
	l.LoggerSugar = l.Logger.Sugar()

	return l
}

type Logger struct {
	Logger      *zap.Logger
	LoggerSugar *zap.SugaredLogger
	Option
}

func (l *Logger) Opt() Option {
	return l.Option
}

func (l *Logger) With(fields ...Field) Interface {
	fieldList := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		fieldList = append(fieldList, zap.Any(field.Key, field.Value))
	}
	cpLog := l.Logger.With(fieldList...)
	return &Logger{
		Logger:      cpLog,
		LoggerSugar: cpLog.Sugar(),
		Option:      l.Option,
	}
}

// Debugf 打印Debug信息
//
// @param: format 格式信息
// @param: v 参数信息
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.LoggerSugar.Debugf(format, v...)
}

// Debug 打印Debug信息
//
// @param: message 格式信息
func (l *Logger) Debug(message string) {
	l.Logger.Debug(message)
}

// Infof 打印Info信息
//
// @param: format 格式信息
// @param: v 参数信息
func (l *Logger) Infof(format string, v ...interface{}) {
	l.LoggerSugar.Infof(format, v...)
}

// Info 打印Info信息
//
// @param: message 格式信息
func (l *Logger) Info(message string) {
	l.Logger.Info(message)
}

// Warnf 打印Warn信息
//
// @param: format 格式信息
// @param: v 参数信息
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.LoggerSugar.Warnf(format, v...)
}

// Warn 打印Warn信息
//
// @param: message 信息
func (l *Logger) Warn(message string) {
	l.Logger.Warn(message)
}

// Errorf 打印Error信息
//
// @param: format 格式信息
// @param: v 参数信息
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.LoggerSugar.Errorf(format, v...)
}

// Error 打印Error信息
//
// @param: message 信息
func (l *Logger) Error(message string) {
	l.Logger.Error(message)
}

// Fatalf 打印Fatalf信息
//
// @param: format 格式信息
// @param: v 参数信息
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.LoggerSugar.Fatalf(format, v...)
}

// Fatal 打印Fatal信息
//
// @param: message 信息
func (l *Logger) Fatal(message string) {
	l.Logger.Fatal(message)
}

func (l *Logger) Sync() error {
	var errs e.Errors
	if err := l.Logger.Sync(); err != nil {
		errs = append(errs, err)
	}
	if err := l.LoggerSugar.Sync(); err != nil {
		errs = append(errs, err)
	}
	return errs
}
