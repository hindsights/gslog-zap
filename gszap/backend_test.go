package gszap

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/hindsights/gslog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logLevelChecker struct {
	level zapcore.Level
}

func (checker logLevelChecker) Enabled(l zapcore.Level) bool {
	return l >= checker.level
}

type inttype int

const (
	intval1 inttype = 1
)

func TestLog(t *testing.T) {
	consoleWriter := zapcore.Lock(os.Stdout)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(consoleEncoder, consoleWriter, logLevelChecker{level: zapcore.DebugLevel})

	tempLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	gslog.SetBackend(NewBackend(tempLogger))
	gslog.Info("gs-hello")
	gslog.Warn("start")
	logger := gslog.GetSugaredLogger("app")
	flogger := gslog.GetLogger("app")
	flogger.Debug("custom type", "intval1", intval1)
	for {
		flogger.Debug("debug", 1, "str")
		flogger.Info("info", "abc")
		flogger.Warn("warn", true)
		flogger.Error("error", false)
		flogger.Error("testerr", errors.New("testerr"))
		flogger.Int("int", 1).Debug("debug", "skey1", "sval1", "skey2", 2)
		flogger.Str("str", "abc").Info("info")
		flogger.Bool("bool", true).Warn("warn")
		flogger.Bool("bool", false).Error("error")
		flogger.Str("key1", "val1").Log(gslog.LogLevelDebug, "log debug", "skey1", "sval1", "skey2", 2)
		flogger.With("key2", "val2", "keyint", 12345).Log(gslog.LogLevelDebug, "log debug with", "skey1", "sval1", "skey2", 2)
		flogger.WithAttrs(gslog.String("key2", "val2"), gslog.Int("keyint", 12345)).Log(gslog.LogLevelDebug, "log attrs", "skey1", "sval1", "skey2", 2)
		flogger.Log(gslog.LogLevelError, "log-msg", "abckey", "abcval", "num", 30)
		time.Sleep(time.Second * 2)
		logger.Debug("debug", 1, "str")
		logger.Info("info", "abc")
		logger.Warn("warn", true)
		logger.Error("error", false)
		logger.Debugf("debugf %s", "name")
		logger.Infof("infof %s", "value")
		logger.Warnf("warnf %d", 20)
		logger.Errorf("errorf %v", 100)
		break
	}
}
