package main

import (
	"flag"
	"fmt"
	"os"

	beebot "code.chrissexton.org/cws/BeeBot"
	"github.com/jzelinskie/geddit"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "modernc.org/sqlite"
)

const version = 1.0

var userAgent = fmt.Sprintf("BeeBot:%.2f (by u/phlyingpenguin)", version)
var scopes = "identity read edit"

var debug = flag.Bool("debug", false, "Turn debug printing on")
var dbFilePath = flag.String("db", "beebot.db", "Database file path")
var logFilePath = flag.String("log", "beebot.json", "Log file path")

var o *geddit.OAuthSession

func main() {
	flag.Parse()

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	logFile, err := os.OpenFile(*logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}

	multi := zerolog.MultiLevelWriter(consoleWriter, logFile)

	log.Logger = zerolog.New(multi).
		With().Timestamp().Caller().Stack().
		Logger()

	log.Info().Msgf("BeeBot v%.2f", version)

	b, err := beebot.New(*dbFilePath, *logFilePath, *debug)
	if err != nil {
		log.Fatal().Err(err).Msg("beebot died")
	}

	// b.Run()
	b.ServeWeb()
}
