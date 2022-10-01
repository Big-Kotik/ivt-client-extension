package main

import (
	"inv-client-extension/ivt/client"
	"inv-client-extension/ivt/proxy"
	"log"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	log.Print("Start")
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
	go pr.ListenAndServe()
	go u.Run(pr.GetChan())

	<-stop
}
