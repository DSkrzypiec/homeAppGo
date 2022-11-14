package main

import (
	"flag"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	TelegramBotTokenEnv  = "HOMEAPP_TELEGRAM_BOT_TOKEN"
	TelegramChannelIdEnv = "HOMEAPP_TELEGRAM_CHANNEL_ID"
)

type Config struct {
	DatabasePath   string
	UseTelegram2FA bool
	Telegram       *TelegramConfig
}

type TelegramConfig struct {
	BotToken  string
	ChannelId string
}

// Parse or fail.
func ParseConfigFlags() Config {
	telegram2fa := flag.Bool("telegram2fa", false, "Use Telegram for two-factor authentication")
	dbPath := flag.String("dbPath", "test.db", "Path to SQLite Home DB")
	flag.Parse()

	var telegramConfig *TelegramConfig

	if *telegram2fa {
		telegramBotToken := os.Getenv(TelegramBotTokenEnv)
		telegramChannelId := os.Getenv(TelegramChannelIdEnv)
		if telegramBotToken == "" {
			log.Fatal().Msgf("[config] Telegram 2FA is on, %s env variable should be set", TelegramBotTokenEnv)
		}
		if telegramChannelId == "" {
			log.Fatal().Msgf("[config] Telegram 2FA is on, %s env variable should be set", TelegramChannelIdEnv)
		}
		telegramConfig = &TelegramConfig{
			BotToken:  telegramBotToken,
			ChannelId: telegramChannelId,
		}
	}

	return Config{
		DatabasePath:   *dbPath,
		UseTelegram2FA: *telegram2fa,
		Telegram:       telegramConfig,
	}
}

func setupZerolog() {
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.SetGlobalLevel(zerolog.DebugLevel)                                               // TODO: make flag for this
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}) // TODO: make flag for this
}
