package main

import (
	"flag"
	"inv-client-extension/ivt/client"
	"inv-client-extension/ivt/proxy"
	"log"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	port  = flag.Int("port", 8080, "port to listen on")
	debug = flag.Bool("debug", false, "enable debug logging")
)

func main() {
	log.Print("Start")
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "log.json",
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
		logger.Sugar().Errorf("Failed: %v", err)
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	pr := proxy.NewProxy(*port, logger.Named("proxy"))
	go func() {
		err := pr.ListenAndServe()
		if err != nil {
			stop <- os.Interrupt
		}
	}()
	go u.Run(pr.GetChan())

	<-stop

	err = pr.Shutdown()
	if err != nil {
		logger.Sugar().Errorf("Failed: %v", err)
		return
	}
}
