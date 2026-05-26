/*
 *
 *  MIT License
 *
 *  (C) Copyright 2024 Hewlett Packard Enterprise Development LP
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
package commands

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewImportCommand creates the "import" command for the Nautobot provider.
// The caller is responsible for setting RunE on the returned command.
func NewImportCommand(base *cobra.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{}

	cmd.Flags().String("default-location", "", "Default location for imported devices")
	cmd.Flags().String("default-role", "", "Default role for imported devices")
	cmd.Flags().String("default-status", "Active", "Default status for imported devices")

	_ = viper.BindPFlag("nautobot.import.default_location", cmd.Flags().Lookup("default-location"))
	_ = viper.BindPFlag("nautobot.import.default_role", cmd.Flags().Lookup("default-role"))
	_ = viper.BindPFlag("nautobot.import.default_status", cmd.Flags().Lookup("default-status"))

	return cmd, nil
}

// NewExportCommand creates the "export" command for the Nautobot provider.
// The caller is responsible for setting RunE on the returned command.
func NewExportCommand(base *cobra.Command) (*cobra.Command, error) {
	cmd := &cobra.Command{}

	cmd.Flags().Bool("create-device-types", true, "Create missing device types in Nautobot")
	cmd.Flags().Bool("create-location-types", true, "Create missing location types in Nautobot")
	cmd.Flags().Bool("create-module-types", true, "Create missing module types in Nautobot")
	cmd.Flags().Bool("create-locations", true, "Create missing locations in Nautobot")
	cmd.Flags().Bool("create-statuses", true, "Create missing statuses in Nautobot")
	cmd.Flags().Bool("create-roles", true, "Create missing roles in Nautobot")
	cmd.Flags().Bool("merge", false, "Merge with existing devices instead of skipping conflicts")
	cmd.Flags().Bool("dry-run", false, "Log planned actions without making API calls")

	_ = viper.BindPFlag("nautobot.export.create_device_types", cmd.Flags().Lookup("create-device-types"))
	_ = viper.BindPFlag("nautobot.export.create_location_types", cmd.Flags().Lookup("create-location-types"))
	_ = viper.BindPFlag("nautobot.export.create_module_types", cmd.Flags().Lookup("create-module-types"))
	_ = viper.BindPFlag("nautobot.export.create_locations", cmd.Flags().Lookup("create-locations"))
	_ = viper.BindPFlag("nautobot.export.create_statuses", cmd.Flags().Lookup("create-statuses"))
	_ = viper.BindPFlag("nautobot.export.create_roles", cmd.Flags().Lookup("create-roles"))
	_ = viper.BindPFlag("nautobot.export.merge", cmd.Flags().Lookup("merge"))
	_ = viper.BindPFlag("nautobot.export.dry_run", cmd.Flags().Lookup("dry-run"))

	return cmd, nil
}
