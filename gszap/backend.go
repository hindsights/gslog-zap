package gszap

import (
	"fmt"
	"time"

	"github.com/hindsights/gslog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	zapLevels map[gslog.LogLevel]zapcore.Level
	gsLevels  map[zapcore.Level]gslog.LogLevel
)

func init() {
	zapLevels = make(map[gslog.LogLevel]zapcore.Level)
	zapLevels[gslog.LogLevelDebug] = zapcore.DebugLevel
	zapLevels[gslog.LogLevelInfo] = zapcore.InfoLevel
	zapLevels[gslog.LogLevelWarn] = zapcore.WarnLevel
	zapLevels[gslog.LogLevelError] = zapcore.ErrorLevel
	zapLevels[gslog.LogLevelFatal] = zapcore.FatalLevel

	gsLevels = make(map[zapcore.Level]gslog.LogLevel)
	gsLevels[zapcore.DebugLevel] = gslog.LogLevelDebug
	gsLevels[zapcore.InfoLevel] = gslog.LogLevelInfo
	gsLevels[zapcore.WarnLevel] = gslog.LogLevelWarn
	gsLevels[zapcore.ErrorLevel] = gslog.LogLevelError
	gsLevels[zapcore.FatalLevel] = gslog.LogLevelFatal
}

func FromGSLogLevel(level gslog.LogLevel) zapcore.Level {
	if zapLevel, ok := zapLevels[level]; ok {
		return zapLevel
	}
	if level > gslog.LogLevelFatal {
		return zap.FatalLevel
	}
	return zap.DebugLevel
}

func ToGSLogLevel(level zapcore.Level) gslog.LogLevel {
	if gsLevel, ok := gsLevels[level]; ok {
		return gsLevel
	}
	if level > zapcore.FatalLevel {
		return gslog.LogLevelFatal
	}
	return gslog.LogLevelDebug
}

type zapBackend struct {
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func (backend *zapBackend) GetLogger(name string) gslog.Logger {
	return fieldLogger{backend: backend, fields: []zap.Field{zap.String("ctx", name)}}
}

func (backend *zapBackend) GetSimpleLogger(name string) gslog.SimpleLogger {
	return simpleLogger{backend: backend, name: name}
}

func NewBackend(logger *zap.Logger) gslog.Backend {
	return &zapBackend{logger: logger, sugarLogger: logger.WithOptions(zap.AddCallerSkip(1)).Sugar()}
}

type fieldLogger struct {
	backend *zapBackend
	fields  []zap.Field
}

func (logger fieldLogger) NeedLog(level gslog.LogLevel) bool {
	return logger.backend.logger.Core().Enabled(FromGSLogLevel(level))
}

func (logger fieldLogger) Log(level gslog.LogLevel, msg string) {
	if !logger.NeedLog(level) {
		return
	}
	if level <= gslog.LogLevelDebug {
		logger.backend.logger.Debug(msg, logger.fields...)
	} else if level == gslog.LogLevelInfo {
		logger.backend.logger.Info(msg, logger.fields...)
	} else if level == gslog.LogLevelWarn {
		logger.backend.logger.Warn(msg, logger.fields...)
	} else if level == gslog.LogLevelError {
		logger.backend.logger.Error(msg, logger.fields...)
	} else if level >= gslog.LogLevelFatal {
		logger.backend.logger.Fatal(msg, logger.fields...)
	}
}

func (logger fieldLogger) Trace(msg string) {
	logger.Debug(msg)
}

func (logger fieldLogger) Debug(msg string) {
	logger.backend.logger.Debug(msg, logger.fields...)
}

func (logger fieldLogger) Info(msg string) {
	logger.backend.logger.Info(msg, logger.fields...)
}

func (logger fieldLogger) Warn(msg string) {
	logger.backend.logger.Warn(msg, logger.fields...)
}

func (logger fieldLogger) Error(msg string) {
	logger.backend.logger.Error(msg, logger.fields...)
}

func (logger fieldLogger) Fatal(msg string) {
	logger.backend.logger.Fatal(msg, logger.fields...)
}

func (logger fieldLogger) Fields(fields gslog.Fields) gslog.Logger {
	newFields := make([]zap.Field, len(logger.fields)+len(fields))
	copy(newFields, logger.fields)
	i := len(logger.fields)
	for k, v := range fields {
		newFields[i] = zap.Any(k, v)
		i++
	}
	return fieldLogger{backend: logger.backend, fields: newFields}
}

func (logger fieldLogger) Field(key string, val interface{}) gslog.Logger {
	return fieldLogger{backend: logger.backend, fields: append(logger.fields, zap.Any(key, val))}
}

func (logger fieldLogger) Str(key string, val string) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Int(key string, val int) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Uint(key string, val uint) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Bool(key string, val bool) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Int64(key string, val int64) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Int32(key string, val int32) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Int16(key string, val int16) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Int8(key string, val int8) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Uint64(key string, val uint64) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Uint32(key string, val uint32) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Uint16(key string, val uint16) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Uint8(key string, val uint8) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Float32(key string, val float32) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Float64(key string, val float64) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Err(key string, val error) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Time(key string, val time.Time) gslog.Logger {
	return logger.Field(key, val)
}

func (logger fieldLogger) Duration(key string, val time.Duration) gslog.Logger {
	return logger.Field(key, val)
}

type simpleLogger struct {
	backend *zapBackend
	name    string
}

func (logger simpleLogger) formatLoggerName() string {
	return fmt.Sprintf("[%s]", logger.name)
}

func (logger simpleLogger) prepareArgs(args []interface{}) []interface{} {
	newArgs := make([]interface{}, len(args)*2+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		// add extra space separator
		newArgs[i*2+1] = " "
		newArgs[i*2+2] = arg
	}
	return newArgs
}

func (logger simpleLogger) prepareFormatArgs(format string, args []interface{}) (string, []interface{}) {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		newArgs[i+1] = arg
	}
	return "%s " + format, newArgs
}

func (logger simpleLogger) doLog(level gslog.LogLevel, f func(...interface{}), args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newArgs := logger.prepareArgs(args)
	f(newArgs...)
}

func (logger simpleLogger) doLogf(level gslog.LogLevel, f func(string, ...interface{}), format string, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newFormat, newArgs := logger.prepareFormatArgs(format, args)
	f(newFormat, newArgs...)
}

func (logger simpleLogger) NeedLog(level gslog.LogLevel) bool {
	return logger.backend.logger.Core().Enabled(FromGSLogLevel(level))
}

func (logger simpleLogger) Logf(level gslog.LogLevel, format string, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	if level <= gslog.LogLevelDebug {
		logger.doLogf(gslog.LogLevelDebug, logger.backend.sugarLogger.Debugf, format, args...)
	} else if level == gslog.LogLevelInfo {
		logger.doLogf(gslog.LogLevelInfo, logger.backend.sugarLogger.Infof, format, args...)
	} else if level == gslog.LogLevelWarn {
		logger.doLogf(gslog.LogLevelWarn, logger.backend.sugarLogger.Warnf, format, args...)
	} else if level >= gslog.LogLevelError {
		logger.doLogf(gslog.LogLevelError, logger.backend.sugarLogger.Errorf, format, args...)
	}
}

func (logger simpleLogger) Log(level gslog.LogLevel, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	if level <= gslog.LogLevelDebug {
		logger.doLog(gslog.LogLevelDebug, logger.backend.sugarLogger.Debug, args...)
	} else if level == gslog.LogLevelInfo {
		logger.doLog(gslog.LogLevelInfo, logger.backend.sugarLogger.Info, args...)
	} else if level == gslog.LogLevelWarn {
		logger.doLog(gslog.LogLevelWarn, logger.backend.sugarLogger.Warn, args...)
	} else if level >= gslog.LogLevelError {
		logger.doLog(gslog.LogLevelError, logger.backend.sugarLogger.Error, args...)
	}
}

func (logger simpleLogger) Trace(args ...interface{}) {
	logger.doLog(gslog.LogLevelDebug, logger.backend.sugarLogger.Debug, args...)
}

func (logger simpleLogger) Debug(args ...interface{}) {
	logger.doLog(gslog.LogLevelDebug, logger.backend.sugarLogger.Debug, args...)
}

func (logger simpleLogger) Info(args ...interface{}) {
	logger.doLog(gslog.LogLevelInfo, logger.backend.sugarLogger.Info, args...)
}

func (logger simpleLogger) Warn(args ...interface{}) {
	logger.doLog(gslog.LogLevelWarn, logger.backend.sugarLogger.Warn, args...)
}

func (logger simpleLogger) Error(args ...interface{}) {
	logger.doLog(gslog.LogLevelError, logger.backend.sugarLogger.Error, args...)
}

func (logger simpleLogger) Fatal(args ...interface{}) {
	logger.doLog(gslog.LogLevelFatal, logger.backend.sugarLogger.Fatal, args...)
}

func (logger simpleLogger) Tracef(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func (logger simpleLogger) Debugf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelDebug, logger.backend.sugarLogger.Debugf, format, args...)
}

func (logger simpleLogger) Infof(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelInfo, logger.backend.sugarLogger.Infof, format, args...)
}

func (logger simpleLogger) Warnf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelWarn, logger.backend.sugarLogger.Warnf, format, args...)
}

func (logger simpleLogger) Errorf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelError, logger.backend.sugarLogger.Errorf, format, args...)
}

func (logger simpleLogger) Fatalf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelFatal, logger.backend.sugarLogger.Fatalf, format, args...)
}
