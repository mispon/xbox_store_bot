package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	xsbot "github.com/mispon/xbox-store-bot/bot"
	"github.com/mispon/xbox-store-bot/bot/cache"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	token         = flag.String("token", "", "-token=qwerty")
	sellerId      = flag.String("seller-id", "", "-seller-id=12345")
	debug         = flag.Bool("debug", false, "-debug=true")
	loadCache     = flag.Bool("load-cache", true, "-load-cache=false")
	chatsFilePath = flag.String("chats-file", "chats.txt", "-chats=bot/chats.txt")
)

func init() {
	flag.Parse()
}

func main() {
	if *token == "" {
		log.Fatal("bot token is not specified")
	}

	if *sellerId == "" {
		log.Fatal("seller id is not specified")
	}

	if *chatsFilePath == "" {
		log.Fatal("chats file path is not specified")
	}

	chatsFile, err := os.OpenFile(*chatsFilePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("failed to open chats file: %v", err)
	}

	logger := mustLogger(*debug)

	botCache, err := cache.New(logger, *sellerId, *loadCache)
	if err != nil {
		log.Fatal(err)
	}

	bot, err := xsbot.New(logger, botCache, chatsFile, *token,
		xsbot.WithSeller(*sellerId),
		xsbot.WithDebug(*debug),
	)
	if err != nil {
		log.Fatal("failed to create bot", err)
	}

	go bot.Run()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)

	<-stopCh

	bot.Stop()
	chatsFile.Close()

	logger.Info("Bot gracefully stopped")
}

func mustLogger(debug bool) *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.DisableStacktrace = true

	logLevel := zapcore.InfoLevel
	if debug {
		logLevel = zapcore.DebugLevel
	}
	cfg.Level = zap.NewAtomicLevelAt(logLevel)

	logger, err := cfg.Build()
	if err != nil {
		log.Fatal(err)
	}

	return logger
}
