package gszap

import (
	"fmt"

	"github.com/hindsights/gslog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func FromGSLogLevel(level gslog.LogLevel) zapcore.Level {
	switch {
	// case level <= gslog.LogLevelTrace:
	// 	return zap.DebugLevel
	case level == gslog.LogLevelDebug:
		return zap.DebugLevel
	case level == gslog.LogLevelInfo:
		return zap.InfoLevel
	case level == gslog.LogLevelWarn:
		return zap.WarnLevel
	case level == gslog.LogLevelError:
		return zap.ErrorLevel
	case level >= gslog.LogLevelFatal:
		return zap.FatalLevel
	}
	return zap.FatalLevel
}

func ToGSLogLevel(level zapcore.Level) gslog.LogLevel {
	switch {
	case level <= zap.DebugLevel:
		return gslog.LogLevelDebug
	case level == zap.InfoLevel:
		return gslog.LogLevelInfo
	case level == zap.WarnLevel:
		return gslog.LogLevelWarn
	case level == zap.ErrorLevel:
		return gslog.LogLevelError
	case level >= zap.FatalLevel:
		return gslog.LogLevelFatal
	}
	return gslog.LogLevelFatal
}

type zapBackend struct {
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func (backend *zapBackend) GetLogger(name string) gslog.Logger {
	return zapLogger{backend: backend, name: name, fields: gslog.Fields{"ctx": name}}
}

func (backend *zapBackend) GetFieldLogger(name string) gslog.FieldLogger {
	return sLogger{backend: backend, name: name, fields: gslog.Fields{"ctx": name}}
}

func NewBackend(logger *zap.Logger) gslog.Backend {
	return &zapBackend{logger: logger, sugarLogger: logger.Sugar()}
}

type sLogger struct {
	backend *zapBackend
	name    string
	fields  gslog.Fields
}

func (logger sLogger) doLog(level gslog.LogLevel, f func(string, ...interface{}), msg string, fields ...gslog.Fields) {
	if !logger.NeedLog(level) {
		return
	}
	allFields := append([]gslog.Fields{logger.fields}, fields...)
	args := make([]interface{}, 0, gslog.GetFieldCount(allFields...)*2)
	for _, fs := range allFields {
		for k, v := range fs {
			args = append(args, k, v)
		}
	}
	f(msg, args...)
}

func (logger sLogger) NeedLog(level gslog.LogLevel) bool {
	return logger.backend.logger.Core().Enabled(FromGSLogLevel(level))
}

func (logger sLogger) Log(level gslog.LogLevel, msg string, fields ...gslog.Fields) {
	if !logger.NeedLog(level) {
		return
	}
	if level <= gslog.LogLevelDebug {
		logger.doLog(gslog.LogLevelDebug, logger.backend.sugarLogger.Debugw, msg, fields...)
	} else if level == gslog.LogLevelInfo {
		logger.doLog(gslog.LogLevelInfo, logger.backend.sugarLogger.Infow, msg, fields...)
	} else if level == gslog.LogLevelWarn {
		logger.doLog(gslog.LogLevelWarn, logger.backend.sugarLogger.Warnw, msg, fields...)
	} else if level >= gslog.LogLevelError {
		logger.doLog(gslog.LogLevelError, logger.backend.sugarLogger.Errorw, msg, fields...)
	}
}

func (logger sLogger) Trace(msg string, fields ...gslog.Fields) {
	logger.Debug(msg, fields...)
}

func (logger sLogger) Debug(msg string, fields ...gslog.Fields) {
	logger.doLog(gslog.LogLevelDebug, logger.backend.sugarLogger.Debugw, msg, fields...)
}

func (logger sLogger) Info(msg string, fields ...gslog.Fields) {
	logger.doLog(gslog.LogLevelInfo, logger.backend.sugarLogger.Infow, msg, fields...)
}

func (logger sLogger) Warn(msg string, fields ...gslog.Fields) {
	logger.doLog(gslog.LogLevelWarn, logger.backend.sugarLogger.Warnw, msg, fields...)
}

func (logger sLogger) Error(msg string, fields ...gslog.Fields) {
	logger.doLog(gslog.LogLevelError, logger.backend.sugarLogger.Errorw, msg, fields...)
}

func (logger sLogger) Fatal(msg string, fields ...gslog.Fields) {
	logger.doLog(gslog.LogLevelFatal, logger.backend.sugarLogger.Fatalw, msg, fields...)
}

func (logger sLogger) WithFields(fields gslog.Fields) gslog.FieldLogger {
	return sLogger{backend: logger.backend, name: logger.name, fields: gslog.JoinFields(logger.fields, fields)}
}

type zapLogger struct {
	backend *zapBackend
	name    string
	fields  gslog.Fields
}

func (logger zapLogger) formatLoggerName() string {
	return fmt.Sprintf("[%s]", logger.name)
}

func (logger zapLogger) prepareArgs(args []interface{}) []interface{} {
	newArgs := make([]interface{}, len(args)*2+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		// add extra space separator
		newArgs[i*2+1] = " "
		newArgs[i*2+2] = arg
	}
	return newArgs
}

func (logger zapLogger) prepareFormatArgs(format string, args []interface{}) (string, []interface{}) {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		newArgs[i+1] = arg
	}
	return "%s " + format, newArgs
}

func (logger zapLogger) doLog(level gslog.LogLevel, f func(...interface{}), args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newArgs := logger.prepareArgs(args)
	f(newArgs...)
}

func (logger zapLogger) doLogf(level gslog.LogLevel, f func(string, ...interface{}), format string, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newFormat, newArgs := logger.prepareFormatArgs(format, args)
	f(newFormat, newArgs...)
}

func (logger zapLogger) NeedLog(level gslog.LogLevel) bool {
	return logger.backend.logger.Core().Enabled(FromGSLogLevel(level))
}

func (logger zapLogger) Logf(level gslog.LogLevel, format string, args ...interface{}) {
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

func (logger zapLogger) Log(level gslog.LogLevel, args ...interface{}) {
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

func (logger zapLogger) Trace(args ...interface{}) {
	logger.Debug(args...)
}

func (logger zapLogger) Debug(args ...interface{}) {
	logger.doLog(gslog.LogLevelDebug, logger.backend.sugarLogger.Debug, args...)
}

func (logger zapLogger) Info(args ...interface{}) {
	logger.doLog(gslog.LogLevelInfo, logger.backend.sugarLogger.Info, args...)
}

func (logger zapLogger) Warn(args ...interface{}) {
	logger.doLog(gslog.LogLevelWarn, logger.backend.sugarLogger.Warn, args...)
}

func (logger zapLogger) Error(args ...interface{}) {
	logger.doLog(gslog.LogLevelError, logger.backend.sugarLogger.Error, args...)
}

func (logger zapLogger) Fatal(args ...interface{}) {
	logger.doLog(gslog.LogLevelFatal, logger.backend.sugarLogger.Fatal, args...)
}

func (logger zapLogger) Tracef(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func (logger zapLogger) Debugf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelDebug, logger.backend.sugarLogger.Debugf, format, args...)
}

func (logger zapLogger) Infof(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelInfo, logger.backend.sugarLogger.Infof, format, args...)
}

func (logger zapLogger) Warnf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelWarn, logger.backend.sugarLogger.Warnf, format, args...)
}

func (logger zapLogger) Errorf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelError, logger.backend.sugarLogger.Errorf, format, args...)
}

func (logger zapLogger) Fatalf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelFatal, logger.backend.sugarLogger.Fatalf, format, args...)
}
