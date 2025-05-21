package logging

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	// BusinessLogger 业务日志记录器
	BusinessLogger *zap.Logger
	// DBLogger 数据库日志记录器
	DBLogger *zap.Logger
)

// Config 日志配置
type Config struct {
	// 日志文件路径
	LogDir string
	// 业务日志文件名
	BusinessLogFile string
	// 数据库日志文件名
	DBLogFile string
	// 日志级别
	Level string
	// 是否输出到控制台
	Console bool
	// 日志轮转配置
	MaxSize    int  // 单个日志文件最大尺寸，单位MB
	MaxBackups int  // 保留的旧日志文件最大数量
	MaxAge     int  // 保留的旧日志文件最大天数
	Compress   bool // 是否压缩旧日志文件
}

// Init 初始化日志系统
func Init(cfg *Config) error {
	// 确保日志目录存在
	if err := os.MkdirAll(cfg.LogDir, 0755); err != nil {
		return fmt.Errorf("create log directory error: %v", err)
	}

	// 设置日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return fmt.Errorf("parse log level error: %v", err)
	}

	// 创建编码器配置
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

	// 创建业务日志记录器
	businessLogger, err := createLogger(
		filepath.Join(cfg.LogDir, cfg.BusinessLogFile),
		level,
		encoderConfig,
		cfg,
	)
	if err != nil {
		return fmt.Errorf("create business logger error: %v", err)
	}
	BusinessLogger = businessLogger

	// 创建数据库日志记录器
	dbLogger, err := createLogger(
		filepath.Join(cfg.LogDir, cfg.DBLogFile),
		level,
		encoderConfig,
		cfg,
	)
	if err != nil {
		return fmt.Errorf("create db logger error: %v", err)
	}
	DBLogger = dbLogger

	return nil
}

// createLogger 创建日志记录器
func createLogger(logFile string, level zapcore.Level, encoderConfig zapcore.EncoderConfig, cfg *Config) (*zap.Logger, error) {
	// 创建日志轮转器
	rotator := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    cfg.MaxSize, // MB
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge, // days
		Compress:   cfg.Compress,
	}

	// 创建文件输出
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(rotator),
		level,
	)

	// 如果配置了控制台输出，创建控制台输出
	var cores []zapcore.Core
	if cfg.Console {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			level,
		)
		cores = append(cores, consoleCore)
	}
	cores = append(cores, fileCore)

	// 创建多输出核心
	core := zapcore.NewTee(cores...)

	// 创建日志记录器
	logger := zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return logger, nil
}

// Sync 同步所有日志记录器
func Sync() {
	if BusinessLogger != nil {
		_ = BusinessLogger.Sync()
	}
	if DBLogger != nil {
		_ = DBLogger.Sync()
	}
}

// 以下是便捷的日志记录方法

// Business 业务日志记录方法
func Business() *zap.Logger {
	if BusinessLogger == nil {
		panic("business logger not initialized")
	}
	return BusinessLogger
}

// DB 数据库日志记录方法
func DB() *zap.Logger {
	if DBLogger == nil {
		panic("db logger not initialized")
	}
	return DBLogger
}
