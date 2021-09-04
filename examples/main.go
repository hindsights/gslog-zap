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
	tempLogger = tempLogger.WithOptions(zap.AddCallerSkip(1))

	gslog.SetBackend(gszap.NewBackend(tempLogger))

	gslog.Info("gs-zap-hello")
	gslog.Warn("zap-start")

	logger := gslog.GetFieldLogger("app")
	logger.Debug("debug", gslog.Fields{"val": 1})
	logger.Info("info", gslog.Fields{"strval": "abc"})
	logger.Warn("warn", gslog.Fields{"boolval": true, "intval": 123})
	logger.Error("error", gslog.Fields{"boolval": false, "name": "hello"})
	logger.WithFields(gslog.Fields{"key1": 1, "key2": "val2"}).Error("field output")
	logger.WithFields(gslog.Fields{"key1": 1, "key2": "val2"}).Info("field output", gslog.Fields{"val": 567})

	gslog.Debugf("debugf %s", "name")
	gslog.Infof("infof %s", "value")
	gslog.Warnf("warnf %d", 20)
	gslog.Errorf("errorf %v", 100)
	logger.Info("output to zap")
}
