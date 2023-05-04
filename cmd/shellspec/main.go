/*
 * Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 */
package main

import (
	"internal/shellspec"
	"os"

	cani "github.com/Cray-HPE/cani/cmd"
	"github.com/rs/zerolog/log"
	//these imports are needed to make sure the modules get built=
)

func main() {
	// Get all cobra commands and flags from the root command and subcommands
	commandInfo := shellspec.GetCommandInfo(cani.RootCmd)
	for _, command := range commandInfo {
		// Write test files to spec/ for each command
		err := command.GenerateSpecfile()
		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}
	}
}
