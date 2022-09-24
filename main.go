package main

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"inv-client-extension/ivt/client"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, // days
	})
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		w,
		zap.InfoLevel,
	)
	logger := zap.New(core)
	u, err := client.NewUser(0, logger)
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
		return
	}
	err = u.SendRequests(client.NewRequestWrapper("test", "test", nil))
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
		return
	}
}
