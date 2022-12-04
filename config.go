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
	Logger                LoggerConfig
	AppVersion            string
	CurrentCommitSHA      string
	PublishViewsAfter     time.Duration
}

type TelegramConfig struct {
	BotToken  string
	ChannelId string
}

type LoggerConfig struct {
	UseDebugLevel    bool
	UseConsoleWriter bool
}

// Parse or fail.
func ParseConfigFlags() Config {
	telegram2fa := flag.Bool("telegram2fa", false, "Use Telegram for two-factor authentication")
	dbPath := flag.String("dbPath", "test.db", "Path to SQLite Home DB")
	port := flag.Int("port", 8080, "Port on which HomeApp is listening")
	publishViewsAfter := flag.Int("publishViewsAfter", 300,
		"After each 'x' minutes endpoints views statistics will be published")

	logDebugLevel := flag.Bool("logDebug", true,
		"Log events on at least debug level. Otherwise info level is assumed.")
	logUseConsoleWriter := flag.Bool("logConsole", true,
		"Use ConsoleWriter within zerolog - pretty but not efficient, mostly for development")
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

	loggerConfig := LoggerConfig{
		UseDebugLevel:    *logDebugLevel,
		UseConsoleWriter: *logUseConsoleWriter,
	}

	appVersion, commitSha := parseVersions(versionFile)

	return Config{
		Port:           *port,
		DatabasePath:   *dbPath,
		UseTelegram2FA: *telegram2fa,
		Telegram:       telegramConfig,

		SessionTimeoutMinutes: SessionTimeoutMinutes,
		HttpClientTimeout:     60 * time.Second,
		Logger:                loggerConfig,

		PublishViewsAfter: time.Duration(*publishViewsAfter) * time.Minute,

		AppVersion:       appVersion,
		CurrentCommitSHA: commitSha,
	}
}

func (c *Config) setupZerolog() {
	zerolog.DurationFieldUnit = time.Millisecond
	if c.Logger.UseDebugLevel {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
	if c.Logger.UseConsoleWriter {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		zerolog.TimeFieldFormat = time.RFC3339
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
