package inventory

import (
	"os"

	"github.com/rs/zerolog"
)

var (
	tl     *os.File
	logger zerolog.Logger
)

func init() {
	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
