package main

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"inv-client-extension/ivt/client"
	"inv-client-extension/ivt/proxy"
	"os"
	"os/signal"
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
		zap.DebugLevel,
	)
	logger := zap.New(core)
	u, err := client.NewUser("@ProxyDemhackBot", logger.Named("client"))
	if err != nil {
		logger.Sugar().Errorf("Failed: %w", err)
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	pr := proxy.NewProxy(logger.Named("proxy"))
	go u.Run(pr.GetChan())
	go pr.ListenAndServe()

	<-stop
}

func testSend(u *client.User) error {
	headers := make(map[string][]string)
	headers["Accept"] = []string{"*/*"}
	headers["Upgrade-Insecure-Requests"] = []string{"1"}
	headers["User-Agent"] = []string{"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/"}
	err := u.SendRequests(client.NewRequestWrapper("http://neverssl.com/", "GET", nil, headers))
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
		return err
	}
	return nil
}
