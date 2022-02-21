package db

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

func NewLog(l *zap.Logger, cfg glogger.Config) glogger.Interface {
	var (
		infoStr      = "%s\n[info] "
		warnStr      = "%s\n[warn] "
		errStr       = "%s\n[error] "
		traceStr     = "%s\n[%.3fms] [rows:%v] %s"
		traceWarnStr = "%s %s\n[%.3fms] [rows:%v] %s"
		traceErrStr  = "%s %s\n[%.3fms] [rows:%v] %s"
	)

	if cfg.Colorful {
		infoStr = glogger.Green + "%s\n" + glogger.Reset + glogger.Green + "[info] " + glogger.Reset
		warnStr = glogger.BlueBold + "%s\n" + glogger.Reset + glogger.Magenta + "[warn] " + glogger.Reset
		errStr = glogger.Magenta + "%s\n" + glogger.Reset + glogger.Red + "[error] " + glogger.Reset
		traceStr = glogger.Green + "%s\n" + glogger.Reset + glogger.Yellow + "[%.3fms] " + glogger.BlueBold +
			"[rows:%v]" + glogger.Reset + " %s"
		traceWarnStr = glogger.Green + "%s " + glogger.Yellow + "%s\n" + glogger.Reset + glogger.RedBold + "[%.3fms] " +
			glogger.Yellow + "[rows:%v]" + glogger.Magenta + " %s" + glogger.Reset
		traceErrStr = glogger.RedBold + "%s " + glogger.MagentaBold + "%s\n" + glogger.Reset + glogger.Yellow +
			"[%.3fms] " + glogger.BlueBold + "[rows:%v]" + glogger.Reset + " %s"
	}
	return &Log{
		logger:       l,
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
	logger *zap.Logger
	glogger.Config
	infoStr, warnStr, errStr            string
	traceStr, traceErrStr, traceWarnStr string
}

func (l *Log) LogMode(level glogger.LogLevel) glogger.Interface {
	log := *l
	log.LogLevel = level
	return &log
}

func (l *Log) Info(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel < glogger.Info {
		return
	}
	l.logger.Sugar().Infof(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

func (l *Log) Warn(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel < glogger.Warn {
		return
	}
	l.logger.Sugar().Warnf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

func (l *Log) Error(_ context.Context, msg string, data ...interface{}) {
	if l.LogLevel < glogger.Error {
		return
	}
	l.logger.Sugar().Errorf(l.infoStr+msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
}

const NanosecondPerMillisecond = 1e6

func (l *Log) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= glogger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= glogger.Error:
		s, rows := fc()
		if rows == -1 {
			l.logger.Sugar().Errorf(l.traceErrStr, utils.FileWithLineNum(), err,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			l.logger.Sugar().Errorf(l.traceErrStr, utils.FileWithLineNum(), err,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= glogger.Warn:
		s, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			l.logger.Sugar().Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			l.logger.Sugar().Warnf(l.traceWarnStr, utils.FileWithLineNum(), slowLog,
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	case l.LogLevel == glogger.Info:
		s, rows := fc()
		if rows == -1 {
			l.logger.Sugar().Infof(l.traceStr, utils.FileWithLineNum(),
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, "-", s)
		} else {
			l.logger.Sugar().Infof(l.traceStr, utils.FileWithLineNum(),
				float64(elapsed.Nanoseconds())/NanosecondPerMillisecond, rows, s)
		}
	}
}
