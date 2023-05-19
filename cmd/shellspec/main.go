/*
 * Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 */
package main

import (
	"os"

	"github.com/Cray-HPE/cani/internal/shellspec"

	cani "github.com/Cray-HPE/cani/cmd"
	_ "github.com/Cray-HPE/cani/cmd/blade"
	_ "github.com/Cray-HPE/cani/cmd/cabinet"
	_ "github.com/Cray-HPE/cani/cmd/cdu"
	_ "github.com/Cray-HPE/cani/cmd/chassis"
	_ "github.com/Cray-HPE/cani/cmd/config"
	_ "github.com/Cray-HPE/cani/cmd/node"
	_ "github.com/Cray-HPE/cani/cmd/pdu"
	_ "github.com/Cray-HPE/cani/cmd/session"
	_ "github.com/Cray-HPE/cani/cmd/switch"
	_ "github.com/Cray-HPE/cani/cmd/taxonomy"
	"github.com/rs/zerolog/log"
	//these imports are needed to make sure the modules get built=
)

func main() {
	// Get all cobra commands and flags from the root command and subcommands
	commandInfo := shellspec.GetCommandInfo(cani.RootCmd)
	for _, command := range commandInfo {
		// Write --help output as a fixture for each command
		err := command.GenerateHelpFixtures()
		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}

		// Write test files to spec/ for each command
		err = command.GenerateSpecfile()
		if err != nil {
			log.Error().Msg(err.Error())
			os.Exit(1)
		}
	}
}
