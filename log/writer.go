package log

import (
	"io"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

// SetLoggerWriter return a io.Writer
func SetLoggerWriter(path string) io.Writer {
	if path == "" {
		return os.Stdout
	}
	return &lumberjack.Logger{
		Filename:   path,
		MaxSize:    DefaultLogSizeM, // 单个日志文件最大MaxSize*M大小
		MaxAge:     MaxLogDays,      // days
		MaxBackups: DefaultMaxZip,   // 备份数量
		Compress:   false,           // 不压缩
		LocalTime:  true,            // 备份名采用本地时间
	}
}
