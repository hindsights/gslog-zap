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

	core := zapcore.NewCore(consoleEncoder, consoleWriter, logLevelChecker{level: zapcore.DebugLevel})

	tempLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	tempLogger = tempLogger.WithOptions(zap.AddCallerSkip(1))
	gslog.SetBackend(NewBackend(tempLogger))
	gslog.Info("gs-hello")
	gslog.Warn("start")
	logger := gslog.GetLogger("app")
	flogger := gslog.GetFieldLogger("app")
	for {
		flogger.Debug("debug", gslog.Fields{"int": 1})
		flogger.Info("info", gslog.Fields{"str": "abc"})
		flogger.Warn("warn", gslog.Fields{"bool": true})
		flogger.Error("error", gslog.Fields{"bool": false})
		flogger.Log(gslog.LogLevelDebug, "log debug", gslog.Fields{"key1": "value1"})
		time.Sleep(time.Second * 3)
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
