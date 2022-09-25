package main

import (
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"inv-client-extension/ivt/client"
	"time"

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
	u, err := client.NewUser("@ProxyDemhackBot", logger)
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
		return
	}
	go u.Run()
	headers := make(map[string][]string)
	headers["Accept"] = []string{"*/*"}
	headers["Upgrade-Insecure-Requests"] = []string{"1"}
	headers["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/"}
	err = u.SendRequests(client.NewRequestWrapper("http://neverssl.com/", "GET", nil, nil))
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
		return
	}
	time.Sleep(30 * time.Second)
}
