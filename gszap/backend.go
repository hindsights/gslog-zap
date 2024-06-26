package gszap

import (
	"fmt"
	"time"

	"github.com/hindsights/gslog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	ErrorKey      = "error"
	LoggerNameKey = "ctx"
)

const badKey = "<badkey>"

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
	getter      func() *zap.Logger
	logger      *zap.Logger
	sugarLogger *zap.SugaredLogger
}

func (backend *zapBackend) NewLogger() gslog.Logger {
	return fieldLogger{backend: backend}
}

func (backend *zapBackend) GetLogger(name string) gslog.Logger {
	var fields []zap.Field
	if name != "" {
		fields = append(fields, zap.String(LoggerNameKey, name))
	}
	return fieldLogger{backend: backend, fields: fields}
}

func (backend *zapBackend) getZapLogger() *zap.Logger {
	if backend.logger != nil {
		return backend.logger
	}
	if backend.getter != nil {
		return backend.getter()
	}
	return nil
}

func (backend *zapBackend) getZapSugarLogger() *zap.SugaredLogger {
	if backend.sugarLogger != nil {
		return backend.sugarLogger
	}
	if backend.logger != nil {
		return backend.logger.Sugar()
	}
	if backend.getter != nil {
		return backend.getter().Sugar()
	}
	return nil
}

func (backend *zapBackend) GetSugaredLogger(name string) gslog.SugaredLogger {
	return sugaredLogger{backend: backend, name: name}
}

func NewBackend(logger *zap.Logger) gslog.Backend {
	zlogger := logger.WithOptions(zap.AddCallerSkip(1))
	return &zapBackend{logger: zlogger, sugarLogger: zlogger.Sugar()}
}

func NewBackendWith(f func() *zap.Logger) gslog.Backend {
	return &zapBackend{getter: f}
}

type fieldLogger struct {
	backend *zapBackend
	fields  []zap.Field
}

func (logger fieldLogger) NeedLog(level gslog.LogLevel) bool {
	return logger.backend.getZapLogger().Core().Enabled(FromGSLogLevel(level))
}

func extractAttr(args []interface{}) (zap.Field, []interface{}) {
	switch x := args[0].(type) {
	case string:
		if len(args) >= 2 {
			return zap.Any(x, args[1]), args[2:]
		}
		return zap.String(badKey, x), nil
	case gslog.Attr:
		return zap.Any(x.Key, x.Value), args[1:]
	case *gslog.Attr:
		return zap.Any(x.Key, x.Value), args[1:]
	case error:
		return zap.NamedError(ErrorKey, x), args[1:]
	default:
		return zap.Any(badKey, x), args[1:]
	}
}

func joinFields(fields []zap.Field, args []interface{}) []zap.Field {
	if len(args) == 0 {
		return fields
	}
	ret := make([]zap.Field, 0, len(fields)+len(args))
	ret = append(ret, fields...)
	var field zap.Field
	for {
		if len(args) == 0 {
			break
		}
		field, args = extractAttr(args)
		ret = append(ret, field)
	}
	return ret
}

func (logger fieldLogger) LogDirect(level gslog.LogLevel, msg string, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	zlogger := logger.backend.getZapLogger()
	fields := joinFields(logger.fields, args)
	if level <= gslog.LogLevelDebug {
		zlogger.Debug(msg, fields...)
	} else if level == gslog.LogLevelInfo {
		zlogger.Info(msg, fields...)
	} else if level == gslog.LogLevelWarn {
		zlogger.Warn(msg, fields...)
	} else if level == gslog.LogLevelError {
		zlogger.Error(msg, fields...)
	} else if level >= gslog.LogLevelFatal {
		zlogger.Fatal(msg, fields...)
	}
}

func (logger fieldLogger) Log(level gslog.LogLevel, msg string, args ...interface{}) {
	logger.LogDirect(level, msg, args...)
}

func (logger fieldLogger) Trace(msg string, args ...interface{}) {
	logger.LogDirect(gslog.LogLevelDebug, msg, args...)
}

func (logger fieldLogger) Debug(msg string, args ...interface{}) {
	logger.LogDirect(gslog.LogLevelDebug, msg, args...)
}

func (logger fieldLogger) Info(msg string, args ...interface{}) {
	logger.LogDirect(gslog.LogLevelInfo, msg, args...)
}

func (logger fieldLogger) Warn(msg string, args ...interface{}) {
	logger.LogDirect(gslog.LogLevelWarn, msg, args...)
}

func (logger fieldLogger) Error(msg string, args ...interface{}) {
	logger.LogDirect(gslog.LogLevelError, msg, args...)
}

func (logger fieldLogger) Fatal(msg string, args ...interface{}) {
	logger.LogDirect(gslog.LogLevelFatal, msg, args...)
}

func (logger fieldLogger) With(args ...interface{}) gslog.Logger {
	return fieldLogger{backend: logger.backend, fields: joinFields(logger.fields, args)}
}

func (logger fieldLogger) WithAttrs(attrs ...gslog.Attr) gslog.Logger {
	newFields := make([]zap.Field, len(logger.fields)+len(attrs))
	copy(newFields, logger.fields)
	i := len(logger.fields)
	for _, attr := range attrs {
		newFields[i] = zap.Any(attr.Key, attr.Value)
		i++
	}
	return fieldLogger{backend: logger.backend, fields: newFields}
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
	newFields := make([]zap.Field, len(logger.fields)+1)
	copy(newFields, logger.fields)
	newFields[len(logger.fields)] = zap.Any(key, val)
	return fieldLogger{backend: logger.backend, fields: newFields}
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

type sugaredLogger struct {
	backend *zapBackend
	name    string
}

func (logger sugaredLogger) formatLoggerName() string {
	return fmt.Sprintf("[%s]", logger.name)
}

func (logger sugaredLogger) prepareArgs(args []interface{}) []interface{} {
	newArgs := make([]interface{}, len(args)*2+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		// add extra space separator
		newArgs[i*2+1] = " "
		newArgs[i*2+2] = arg
	}
	return newArgs
}

func (logger sugaredLogger) prepareFormatArgs(format string, args []interface{}) (string, []interface{}) {
	newArgs := make([]interface{}, len(args)+1)
	newArgs[0] = logger.formatLoggerName()
	for i, arg := range args {
		newArgs[i+1] = arg
	}
	return "%s " + format, newArgs
}

// func (logger sugaredLogger) doLog(level gslog.LogLevel, f func(...interface{}), args ...interface{}) {
// 	if !logger.NeedLog(level) {
// 		return
// 	}
// 	newArgs := logger.prepareArgs(args)
// 	f(newArgs...)
// }

// func (logger sugaredLogger) doLogf(level gslog.LogLevel, f func(string, ...interface{}), format string, args ...interface{}) {
// 	if !logger.NeedLog(level) {
// 		return
// 	}
// 	newFormat, newArgs := logger.prepareFormatArgs(format, args)
// 	f(newFormat, newArgs...)
// }

func (logger sugaredLogger) NeedLog(level gslog.LogLevel) bool {
	return logger.backend.getZapLogger().Core().Enabled(FromGSLogLevel(level))
}

func (logger sugaredLogger) Logf(level gslog.LogLevel, format string, args ...interface{}) {
	logger.LogfDirect(level, format, args...)
}

func (logger sugaredLogger) LogfDirect(level gslog.LogLevel, format string, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newFormat, newArgs := logger.prepareFormatArgs(format, args)
	if level <= gslog.LogLevelDebug {
		logger.backend.getZapSugarLogger().Debugf(newFormat, newArgs...)
	} else if level == gslog.LogLevelInfo {
		logger.backend.getZapSugarLogger().Infof(newFormat, newArgs...)
	} else if level == gslog.LogLevelWarn {
		logger.backend.getZapSugarLogger().Warnf(newFormat, newArgs...)
	} else if level >= gslog.LogLevelError {
		logger.backend.getZapSugarLogger().Errorf(newFormat, newArgs...)
	}
}

func (logger sugaredLogger) LogDirect(level gslog.LogLevel, args ...interface{}) {
	if !logger.NeedLog(level) {
		return
	}
	newArgs := logger.prepareArgs(args)
	if level <= gslog.LogLevelDebug {
		logger.backend.getZapSugarLogger().Debug(newArgs...)
	} else if level == gslog.LogLevelInfo {
		logger.backend.getZapSugarLogger().Info(newArgs...)
	} else if level == gslog.LogLevelWarn {
		logger.backend.getZapSugarLogger().Warn(newArgs...)
	} else if level >= gslog.LogLevelError {
		logger.backend.getZapSugarLogger().Error(newArgs...)
	}
}

func (logger sugaredLogger) Log(level gslog.LogLevel, args ...interface{}) {
	logger.LogDirect(level, args...)
}

func (logger sugaredLogger) Trace(args ...interface{}) {
	logger.LogDirect(gslog.LogLevelDebug, args...)
}

func (logger sugaredLogger) Debug(args ...interface{}) {
	logger.LogDirect(gslog.LogLevelDebug, args...)
}

func (logger sugaredLogger) Info(args ...interface{}) {
	logger.LogDirect(gslog.LogLevelInfo, args...)
}

func (logger sugaredLogger) Warn(args ...interface{}) {
	logger.LogDirect(gslog.LogLevelWarn, args...)
}

func (logger sugaredLogger) Error(args ...interface{}) {
	logger.LogDirect(gslog.LogLevelError, args...)
}

func (logger sugaredLogger) Fatal(args ...interface{}) {
	logger.LogDirect(gslog.LogLevelFatal, args...)
}

func (logger sugaredLogger) Tracef(format string, args ...interface{}) {
	logger.Debugf(format, args...)
}

func (logger sugaredLogger) Debugf(format string, args ...interface{}) {
	logger.LogfDirect(gslog.LogLevelDebug, format, args...)
}

func (logger sugaredLogger) Infof(format string, args ...interface{}) {
	logger.LogfDirect(gslog.LogLevelInfo, format, args...)
}

func (logger sugaredLogger) Warnf(format string, args ...interface{}) {
	logger.LogfDirect(gslog.LogLevelWarn, format, args...)
}

func (logger sugaredLogger) Errorf(format string, args ...interface{}) {
	logger.LogfDirect(gslog.LogLevelError, format, args...)
}

func (logger sugaredLogger) Fatalf(format string, args ...interface{}) {
	logger.LogfDirect(gslog.LogLevelFatal, format, args...)
}
