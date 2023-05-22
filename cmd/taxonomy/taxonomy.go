package taxonomy

import (
	"path/filepath"
)

const (
	App              = "cani"
	DsFile           = App + "db.json"
	LogFile          = App + "db.log"
	CfgFile          = App + ".yml"
	CfgDir           = "." + App
	SessionExt       = ".session"
	ShortDescription = "From subfloor to top-of-rack, manage your HPC cluster's inventory!"
	LongDescription  = `From subfloor to top-of-rack, manage your HPC cluster's inventory!`
)

var (
	DsPath  = filepath.Join(CfgDir, DsFile)
	CfgPath = filepath.Join(CfgDir, CfgFile)
)
