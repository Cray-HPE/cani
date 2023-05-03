/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package cabinet

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/rs/zerolog/log"
)

// writeHelperScriptsToDisk writes the helper scripts for adding a cabinet to disk
func writeHelperScriptsToDisk() error {
	// loop through all files in the helperScripts embed.FS
	files, err := fs.ReadDir(helperScripts, "scripts")
	if err != nil {
		return err
	}
	for _, file := range files {
		src := fmt.Sprintf("scripts/%s", file.Name())
		content, err := helperScripts.ReadFile(src)
		if err != nil {
			return err
		}
		log.Info().Msg(fmt.Sprintf("Unpacking %s...\n", file.Name()))
		dest := fmt.Sprintf("/tmp/%s", file.Name())
		err = os.WriteFile(dest, content, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}
