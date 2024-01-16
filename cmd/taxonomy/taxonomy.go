/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package taxonomy

import (
	"path/filepath"
	"sort"

	"github.com/rs/zerolog/log"
)

const (
	App              = "cani"
	CSM              = "csm"
	HPCM             = "hpcm"
	DsFile           = App + "db.json"
	DsFileCSM        = App + "db.json"
	LogFile          = App + "db.log"
	CfgFile          = App + ".yml"
	CfgDir           = "." + App
	ShortDescription = "From subfloor to top-of-rack, manage your HPC cluster's inventory!"
	LongDescription  = `From subfloor to top-of-rack, manage your HPC cluster's inventory!`
	InitShort        = `Initialize a session and import from the chosen provider.`
	InitLong         = `Initialize a session. This will perform an import using provider-defined logic.`
)

var (
	DsPath             = filepath.Join(CfgDir, DsFile)
	CfgPath            = filepath.Join(CfgDir, CfgFile)
	SupportedProviders = []string{CSM, HPCM}
)

func Init() {
	log.Trace().Msgf("%+v", "github.com/Cray-HPE/cani/cmd/taxonomy.init")
	sort.Strings(SupportedProviders)
}
