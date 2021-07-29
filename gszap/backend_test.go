package gszap

import (
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

func TestLog(t *testing.T) {
	consoleWriter := zapcore.Lock(os.Stdout)

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	var core zapcore.Core
	core = zapcore.NewCore(consoleEncoder, consoleWriter, logLevelChecker{level: zapcore.DebugLevel})

	tempLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	tempLogger = tempLogger.WithOptions(zap.AddCallerSkip(1))
	gslog.SetBackend(NewBackend(gslog.LogLevelAll, tempLogger))
	gslog.Info("gs-hello")
	gslog.Warn("start")
	logger := gslog.GetLogger("app")
	for {
		logger.Debug("debug", 1)
		logger.Info("info", "abc")
		logger.Warn("warn", true)
		logger.Error("error", false)
		// logger.WithFields(gslog.Fields{"key1": 1, "key2": "val2"}).Error("field output")
		logger.Log(gslog.LogLevelDebug, "log debug", "value1", "value2")
		logger.Logf(gslog.LogLevelDebug, "log debug format key1=%s key2=%d", "value1", 123)
		// time.Sleep(time.Second)
		// gslog.Debugf("debugf %s", "name")
		// gslog.Infof("infof %s", "value")
		// gslog.Warnf("warnf %d", 20)
		// gslog.Errorf("errorf %v", 100)
		time.Sleep(time.Second * 3)
		break
	}
}
