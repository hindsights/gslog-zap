package main

import (
	"fmt"
	"os"

	"github.com/hindsights/gslog"
	"github.com/hindsights/gslog-zap/gszap"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type logLevelChecker struct {
	level zapcore.Level
}

func (checker logLevelChecker) Enabled(l zapcore.Level) bool {
	return l >= checker.level
}

func main() {
	fmt.Println("test")

	consoleWriter := zapcore.Lock(os.Stdout)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	core := zapcore.NewCore(consoleEncoder, consoleWriter, logLevelChecker{level: zapcore.DebugLevel})
	tempLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	gslog.SetBackend(gszap.NewBackend(tempLogger))

	gslog.Info("gs-zap-hello")
	gslog.Warn("zap-start")

	flogger := gslog.GetLogger("app")
	flogger.Int("val", 1).Debug("debug")
	flogger.Str("str", "abc").Info("info")
	flogger.Bool("bool", true).Warn("warn")
	flogger.Bool("bool", false).Str("name", "hello").Error("error")
	flogger.Fields(gslog.Fields{"key1": 1, "key2": "val2"}).Error("field output")
	flogger.Fields(gslog.Fields{"key1": 1, "key2": "val2"}).Field("val", 566).Info("field output")

	gslog.Debugf("debugf %s", "name")
	gslog.Infof("infof %s", "value")
	gslog.Warnf("warnf %d", 20)
	gslog.Errorf("errorf %v", 100)
	logger := gslog.GetSugaredLogger("slog")
	logger.Info("output to zap", 123)
	logger.Debug("debug", 1, "str")
	logger.Info("info", "abc")
	logger.Warn("warn", true)
	logger.Error("error", false)
	logger.Debugf("debugf %s", "name")
	logger.Infof("infof %s", "value")
	logger.Warnf("warnf %d", 20)
	logger.Errorf("errorf %v", 100)
}
