package db

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"

	"github.com/crochee/lirity/log"
)

func NewLog(l log.Interface, cfg logger.Config) logger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if cfg.Colorful {
		infoStr = logger.Green + "%s\n" + logger.Reset + logger.Green + "[info] " + logger.Reset
		warnStr = logger.BlueBold + "%s\n" + logger.Reset + logger.Magenta + "[warn] " + logger.Reset
		errStr = logger.Magenta + "%s\n" + logger.Reset + logger.Red + "[error] " + logger.Reset
		traceStr = logger.Green + "%s\n" + logger.Reset + logger.Yellow + "[%.3fms] " + logger.BlueBold +
			"[rows:%v]" + logger.Reset + " %s"
		traceWarnStr = logger.Green + "%s " + logger.Yellow + "%s\n" + logger.Reset + logger.RedBold + "[%.3fms] " +
			logger.Yellow + "[rows:%v]" + logger.Magenta + " %s" + logger.Reset
		traceErrStr = logger.RedBold + "%s " + logger.MagentaBold + "%s\n" + logger.Reset + logger.Yellow +
			"[%.3fms] " + logger.BlueBold + "[rows:%v]" + logger.Reset + " %s"
	}
	return &Log{
		Interface:    l,
		Config:       cfg,
		infoStr:      infoStr,
		warnStr:      warnStr,
		errStr:       errStr,
		traceStr:     traceStr,
		traceWarnStr: traceWarnStr,
		traceErrStr:  traceErrStr,
	}
}

type Log struct {
	log.Interface
	logger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func (l *Log) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

func (l *Log) Info(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		l.Interface.Infof(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Log) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		l.Interface.Warnf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

func (l *Log) Error(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		l.Interface.Errorf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

const NanosecondPerMillisecond = 1e6

func (l *Log) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error:
		s, rows := fc()
		if rows == -1 {
			l.Interface.Errorf(l.traceErrStr, utils.FileWithLineNum(), err,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			l.Interface.Errorf(l.traceErrStr, utils.FileWithLineNum(), err,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		s, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.Interface.Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			l.Interface.Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case l.LogLevel == logger.Info:
		s, rows := fc()
		if rows == -1 {
			l.Interface.Infof(l.traceStr, utils.FileWithLineNum(),
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			l.Interface.Infof(l.traceStr, utils.FileWithLineNum(),
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	}
}
