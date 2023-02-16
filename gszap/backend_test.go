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
	gslog.SetBackend(NewBackend(tempLogger))
	gslog.Info("gs-hello")
	gslog.Warn("start")
	logger := gslog.GetSugaredLogger("app")
	flogger := gslog.GetLogger("app")
	for {
		flogger.Int("int", 1).Debug("debug", "skey1", "sval1", "skey2", 2)
		flogger.Str("str", "abc").Info("info")
		flogger.Bool("bool", true).Warn("warn")
		flogger.Bool("bool", false).Error("error")
		flogger.Str("key1", "val1").Log(gslog.LogLevelDebug, "log debug", "skey1", "sval1", "skey2", 2)
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
