package taxonomy

import (
	"path/filepath"
)

const (
	App     = "cani"
	DbFile  = App + "db.json"
	CfgFile = App + ".yml"
	CfgDir  = "." + App
)

var (
	DbPath  = filepath.Join(CfgDir, DbFile)
	CfgPath = filepath.Join(CfgDir, CfgFile)
)
