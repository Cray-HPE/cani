package inventory

import (
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

var transactionLogFile *os.File

// logTransaction logs a transaction to logger
func logTransaction(operation string, key string, value interface{}, err error) {

	// Get the current timestamp
	timestamp := time.Now()

	// Determine the operation status
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
	}

	// Log the transaction
	logEvent := log.With().
		Timestamp().
		Str("timestamp", timestamp.Format(time.RFC3339)).
		Str("operation", operation).
		Str("key", key).
		Interface("value", value).
		Str("status", status).
		Logger()

	if err != nil {
		logEvent.Err(err).Msg("Transaction")
	} else {
		logEvent.Info().Msg("Transaction")
	}
}

// CloseTransactionLog closes the transaction log file
func CloseTransactionLog() {
	if transactionLogFile != nil {
		transactionLogFile.Close()
	}
}
