package taxonomy

import (
	"path/filepath"
)

const (
	App        = "cani"
	DsFile     = App + "db.json"
	CfgFile    = App + ".yml"
	CfgDir     = "." + App
	SessionExt = ".session"
)

var (
	DsPath  = filepath.Join(CfgDir, DsFile)
	CfgPath = filepath.Join(CfgDir, CfgFile)
)
