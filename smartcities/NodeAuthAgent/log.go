package main

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLog() {
	// lumberjack write log in multi files
	loggerIO := &lumberjack.Logger{
		Filename:   "/var/log/NodeAuthAgent.log",
		MaxSize:    10,
		MaxBackups: 2,
		MaxAge:     30,
	}
	// use lumberjack instead the default writer in zapcore
	writeSyncer := zapcore.AddSync(loggerIO)

	// get a zap encoder config, can special output format here
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	// custom lv
	var lv = new(zapcore.Level)
	lv.Set("debug")

	core := zapcore.NewCore(encoder, writeSyncer, lv)
	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)
}
