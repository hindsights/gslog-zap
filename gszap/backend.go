package gszap

import (
	"fmt"

	"github.com/hindsights/gslog"
	"go.uber.org/zap"
)

type zapBackend struct {
	logLevel    gslog.LogLevel
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func (backend *zapBackend) GetLogger(name string) gslog.Logger {
	return &zapLogger{backend: backend, name: name}
}

func NewBackend(logLevel gslog.LogLevel, logger *zap.Logger) gslog.Backend {
	return &zapBackend{logLevel: logLevel, logger: logger, sugarLogger: logger.Sugar()}
}

type zapLogger struct {
	backend *zapBackend
	name    string
}

// func (logger *zapLogger) Name() string {
// 	return logger.name
// }

func (logger *zapLogger) formatLoggerName() string {
	return fmt.Sprintf("[%6s]", logger.name)
}

func (logger *zapLogger) prepareArgs(args []interface{}) []interface{} {
	newArgs := make([]interface{}, len(args)*2+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		// add extra space separator
		newArgs[i*2+1] = " "
		newArgs[i*2+2] = arg
	}
	return newArgs
}

func (logger *zapLogger) prepareFormatArgs(format string, args []interface{}) (string, []interface{}) {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		newArgs[i+1] = arg
	}
	return "%s " + format, newArgs
}

func (logger *zapLogger) doLog(level gslog.LogLevel, f func(...interface{}), args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newArgs := logger.prepareArgs(args)
	f(newArgs...)
}

func (logger *zapLogger) doLogf(level gslog.LogLevel, f func(string, ...interface{}), format string, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newFormat, newArgs := logger.prepareFormatArgs(format, args)
	f(newFormat, newArgs...)
}

func (logger *zapLogger) NeedLog(level gslog.LogLevel) bool {
	return level >= logger.backend.logLevel
}

func (logger *zapLogger) Logf(level gslog.LogLevel, format string, args ...interface{}) {
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

func (logger *zapLogger) Log(level gslog.LogLevel, args ...interface{}) {
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

func (logger *zapLogger) Trace(args ...interface{}) {
	logger.Debug(args...)
}

func (logger *zapLogger) Debug(args ...interface{}) {
	logger.doLog(gslog.LogLevelDebug, logger.backend.sugarLogger.Debug, args...)
}

func (logger *zapLogger) Info(args ...interface{}) {
	logger.doLog(gslog.LogLevelInfo, logger.backend.sugarLogger.Info, args...)
}

func (logger *zapLogger) Warn(args ...interface{}) {
	logger.doLog(gslog.LogLevelWarn, logger.backend.sugarLogger.Warn, args...)
}

func (logger *zapLogger) Error(args ...interface{}) {
	logger.doLog(gslog.LogLevelError, logger.backend.sugarLogger.Error, args...)
}

func (logger *zapLogger) Fatal(args ...interface{}) {
	logger.doLog(gslog.LogLevelFatal, logger.backend.sugarLogger.Fatal, args...)
}

func (logger *zapLogger) Tracef(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func (logger *zapLogger) Debugf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelDebug, logger.backend.sugarLogger.Debugf, format, args...)
}

func (logger *zapLogger) Infof(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelInfo, logger.backend.sugarLogger.Infof, format, args...)
}

func (logger *zapLogger) Warnf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelWarn, logger.backend.sugarLogger.Warnf, format, args...)
}

func (logger *zapLogger) Errorf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelError, logger.backend.sugarLogger.Errorf, format, args...)
}

func (logger *zapLogger) Fatalf(format string, args ...interface{}) {
	logger.doLogf(gslog.LogLevelFatal, logger.backend.sugarLogger.Fatalf, format, args...)
}

func (logger *zapLogger) WithFields(fields gslog.Fields) gslog.Logger {
	return gslog.NewFieldsLogger(logger, fields)
}
