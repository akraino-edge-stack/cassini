package main

import (
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitLog() {
	// lumberjack write log in multi files
	loggerIO := &lumberjack.Logger{
		Filename:   Conf.Log.FileName,
		MaxSize:    Conf.Log.MaxSize,
		MaxBackups: Conf.Log.MaxBackups,
		MaxAge:     Conf.Log.MaxAge,
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
	if lv.UnmarshalText([]byte(Conf.Log.Level)) != nil {
		panic("Error, unkonw debug level")
	}

	core := zapcore.NewCore(encoder, writeSyncer, lv)
	logger := zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(logger)
}
