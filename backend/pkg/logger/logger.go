package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 日志接口
type Logger struct {
	zap *zap.Logger
}

// Field 日志字段
type Field = zap.Field

// Config 日志配置
type Config struct {
	Level      string // 日志级别: debug, info, warn, error
	OutputPath string // 输出路径（为空则只输出到控制台）
	MaxSize    int    // 单个文件最大大小(MB)
	MaxBackups int    // 保留的旧文件最大数量
	MaxAge     int    // 保留的旧文件最大天数
	Compress   bool   // 是否压缩旧文件
	Console    bool   // 是否同时输出到控制台
}

// New 创建日志实例
func New(cfg *Config) (*Logger, error) {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch cfg.Level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 彩色输出
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006/01/02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 配置输出
	var cores []zapcore.Core
	
	// 控制台输出（彩色）
	if cfg.Console || cfg.OutputPath == "" {
		consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
		consoleCore := zapcore.NewCore(
			consoleEncoder,
			zapcore.AddSync(os.Stdout),
			zapLevel,
		)
		cores = append(cores, consoleCore)
	}
	
	// 文件输出（JSON格式）
	if cfg.OutputPath != "" {
		fileEncoderConfig := encoderConfig
		fileEncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder // 文件不需要颜色
		fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
		
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.OutputPath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		})
		
		fileCore := zapcore.NewCore(
			fileEncoder,
			fileWriter,
			zapLevel,
		)
		cores = append(cores, fileCore)
	}

	// 创建核心
	core := zapcore.NewTee(cores...)

	// 创建 logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{zap: zapLogger}, nil
}

// Debug 调试日志
func (l *Logger) Debug(msg string, fields ...Field) {
	l.zap.Debug(msg, fields...)
}

// Info 信息日志
func (l *Logger) Info(msg string, fields ...Field) {
	l.zap.Info(msg, fields...)
}

// Warn 警告日志
func (l *Logger) Warn(msg string, fields ...Field) {
	l.zap.Warn(msg, fields...)
}

// Error 错误日志
func (l *Logger) Error(msg string, fields ...Field) {
	l.zap.Error(msg, fields...)
}

// Fatal 致命错误日志
func (l *Logger) Fatal(msg string, fields ...Field) {
	l.zap.Fatal(msg, fields...)
}

// With 添加字段
func (l *Logger) With(fields ...Field) *Logger {
	return &Logger{zap: l.zap.With(fields...)}
}

// Sync 同步日志
func (l *Logger) Sync() error {
	return l.zap.Sync()
}

// 便捷字段函数
func String(key, val string) Field {
	return zap.String(key, val)
}

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

func Uint(key string, val uint) Field {
	return zap.Uint(key, val)
}

func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) Field {
	return zap.Duration(key, val)
}

func Error(err error) Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}
