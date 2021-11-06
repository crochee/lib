package log

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

// SetLoggerWriter return a io.Writer
func SetLoggerWriter(path string) io.Writer {
	if path == "" {
		return os.Stdout
	}
	return &lumberjack.Logger{
		Filename:   filepath.Clean("./log/" + path),
		MaxSize:    DefaultLogSizeM, // 单个日志文件最大MaxSize*M大小
		MaxAge:     MaxLogDays,      // days
		MaxBackups: DefaultMaxZip,   // 备份数量
		Compress:   false,           // 不压缩
		LocalTime:  true,            // 备份名采用本地时间
	}
}

// JudgeLevel return level by mode
func JudgeLevel(level, mode string) string {
	level = strings.ToUpper(level)
	if level == DEBUG && !strings.EqualFold(mode, "debug") {
		level = INFO
	}
	return level
}
