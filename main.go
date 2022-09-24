package main

import (
	"fmt"
	"inv-client-extension/ivt/client"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		nil,
		zap.InfoLevel,
	)
	logger := zap.New(core)
	u, err := client.NewUser(0, logger)
	if err != nil {
		fmt.Printf("Failed: %s", err.Error())
		return
	}
	u.SendRequests(client.NewRequestsWrapper(client.NewRequestWrapper("test", "test", nil)))
}
