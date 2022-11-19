package main

import (
	_ "embed"
	"flag"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	SessionTimeoutMinutes = 10
	TelegramBotTokenEnv   = "HOMEAPP_TELEGRAM_BOT_TOKEN"
	TelegramChannelIdEnv  = "HOMEAPP_TELEGRAM_CHANNEL_ID"
)

//go:generate sh -c "head -1 CHANGELOG.md > VERSION.txt"
//go:generate sh -c "printf %s $(git rev-parse HEAD) >> VERSION.txt"
//go:embed VERSION.txt
var versionFile string

type Config struct {
	Port                  int
	DatabasePath          string
	UseTelegram2FA        bool
	Telegram              *TelegramConfig
	SessionTimeoutMinutes int
	HttpClientTimeout     time.Duration
	AppVersion            string
	CurrentCommitSHA      string
	PublishViewsAfter     time.Duration
}

type TelegramConfig struct {
	BotToken  string
	ChannelId string
}

// Parse or fail.
func ParseConfigFlags() Config {
	telegram2fa := flag.Bool("telegram2fa", false, "Use Telegram for two-factor authentication")
	dbPath := flag.String("dbPath", "test.db", "Path to SQLite Home DB")
	port := flag.Int("port", 8080, "Port on which HomeApp is listening")
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

	appVersion, commitSha := parseVersions(versionFile)

	return Config{
		Port:           *port,
		DatabasePath:   *dbPath,
		UseTelegram2FA: *telegram2fa,
		Telegram:       telegramConfig,

		SessionTimeoutMinutes: SessionTimeoutMinutes,
		HttpClientTimeout:     60 * time.Second,

		PublishViewsAfter: 1 * time.Hour,

		AppVersion:       appVersion,
		CurrentCommitSHA: commitSha,
	}
}

func parseVersions(input string) (string, string) {
	const firstSHAChars = 8

	lines := strings.Split(input, "\n")
	if len(lines) != 2 {
		log.Fatal().Msg("[config] VERSION.txt file should contain exactly 2 lines - app version and commit SHA")
	}
	_, appVersion, appVerOk := strings.Cut(lines[0], " ")
	if !appVerOk {
		log.Fatal().Msgf("[config] incorrect app version. Should be taken from CHANGELOG in form of '# v0.1.0', got: %s",
			lines[0])
	}
	commitSha := lines[1]
	if len(commitSha) <= firstSHAChars {
		log.Fatal().Msgf("[config] incorrect commit SHA length, got: %d", len(commitSha))
	}
	shortSha := lines[1][:firstSHAChars] + "..."

	return appVersion, shortSha
}

func setupZerolog() {
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.SetGlobalLevel(zerolog.DebugLevel)                                               // TODO: make flag for this
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}) // TODO: make flag for this
}
