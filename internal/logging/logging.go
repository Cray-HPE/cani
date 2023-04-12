package logging

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func NewLogger(logBaseDirectory, logName string) zerolog.Logger {
	// Create directory to persist data from this run like logs and backups!
	timestamp := strings.Replace(time.Now().UTC().Format(time.RFC3339), ":", "-", -1)
	logDirectory := path.Join(logBaseDirectory, fmt.Sprintf("hardware-topology-assistant_%s", timestamp))
	log.Printf("Log directory is at %s", logDirectory)
	if err := os.MkdirAll(logDirectory, 0700); err != nil {
		log.Fatal().Str("logDirectory", logDirectory).Err(err).Msg("Failed to create log directory")
	}

	// Setup the log package to write to both stdout and a log file
	logFilePath := path.Join(logDirectory, logName)
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0600)
	if err != nil {
		log.Fatal().Str("logFilePath", logFilePath).Err(err).Msg("Failed to open log file")
	}

	// Write to both stdout and a log file
	logWriter := io.MultiWriter(os.Stdout, logFile)

	return zerolog.New(logWriter).With().Timestamp().Logger()
}
