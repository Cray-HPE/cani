package ngsm

import (
	stdlog "log"
	"os"

	"github.com/Cray-HPE/cani/internal/provider"
)

var ngsm *Ngsm
var log = stdlog.New(os.Stdout, "[ngsm] ", stdlog.LstdFlags)

func init() {
	ngsm = New()
	provider.Register("ngsm", ngsm)
}
