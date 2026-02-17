package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.SugaredLogger

type Config struct {
	Level      string // debug, info, warn, error
	Filename   string // 日志文件路径
	MaxSize    int    // 单文件最大大小(MB)
	MaxBackups int    // 保留旧文件数量
	MaxAge     int    // 保留天数
	Compress   bool   // 是否压缩
}

func Init(cfg Config) error {
	level := parseLevel(cfg.Level)

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)

	cores := []zapcore.Core{
		zapcore.NewCore(consoleEncoder, consoleWriter, level),
	}

	// 文件输出（如果配置了）
	if cfg.Filename != "" {
		if err := ensureDir(cfg.Filename); err != nil {
			return err
		}

		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})

		fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(fileEncoder, fileWriter, level))
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Log = logger.Sugar()

	return nil
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

func ensureDir(filename string) error {
	dir := filepath.Dir(filename)
	return os.MkdirAll(dir, 0755)
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
