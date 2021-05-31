package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/adamplansky/go-bridge-mentoring/http-upload-server/http"
)

func main() {
	aa := zap.NewDevelopmentEncoderConfig()
	aa.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(aa),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))

	//logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	var maxUploadRequests int
	flag.IntVar(&maxUploadRequests, "max-upload-requests", 1, "max upload requests concurrently")

	s := http.NewServer(logger, maxUploadRequests)
	logger.Info("server running on :8080")
	if err := s.Run(ctx, ":8080"); err != nil {
		log.Fatal(err)
	}
}
