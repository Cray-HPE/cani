/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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
package nautobot

import (
	"fmt"
	"strings"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/export"
)

// The export option that lists the provider metadata to send as custom
// fields, and the flag that replaces it for one export.
const (
	providerMetadataFieldsKey  = "provider_metadata_fields"
	providerMetadataFieldsFlag = "provider-metadata-field"
)

// providerMetadataFields resolves the option for one export: the entries
// the flag gives, or those of the config file when the flag is not given.
// The flag replaces the configured list rather than adding to it, so a
// config value the flag replaces is not read.
func (p *Nautobot) providerMetadataFields(cmd *cli.Command) ([]export.ProviderMetadataField, error) {
	entries, err := p.exportStringsFromConfig(providerMetadataFieldsKey)
	if cmd.Flags().Changed(providerMetadataFieldsFlag) {
		entries, err = cmd.Flags().GetStringSlice(providerMetadataFieldsFlag)
	}
	if err != nil {
		return nil, err
	}
	p.Options.Export.ProviderMetadataFields = entries
	return export.ParseProviderMetadataFields(entries)
}

// splitList splits a comma-separated value, dropping the spaces around each
// entry and any empty one.
func splitList(value string) []string {
	entries := []string{}
	for _, entry := range strings.Split(value, ",") {
		if entry = strings.TrimSpace(entry); entry != "" {
			entries = append(entries, entry)
		}
	}
	return entries
}

// exportStringsFromConfig returns an export-scoped list from the config
// file: a YAML list of text entries, or one comma-separated string. A key
// that is absent or empty is no entries. Any other value is an error, so a
// list written as a mapping is reported rather than read as nothing.
func (p *Nautobot) exportStringsFromConfig(key string) ([]string, error) {
	value, ok := config.GetNestedValue(providerSlug, "export", key)
	if !ok || value == nil {
		return nil, nil
	}
	switch list := value.(type) {
	case string:
		return splitList(list), nil
	case []string:
		return list, nil
	case []any:
		return textEntries(key, list)
	default:
		return nil, fmt.Errorf("%s in the config file must be a list of provider.key or provider.key=custom_field entries, not %v", key, value)
	}
}

// textEntries returns the entries of a decoded YAML list, all of which must
// be text. "- csm.xname: node_id" decodes to a mapping, not to the entry it
// resembles.
func textEntries(key string, list []any) ([]string, error) {
	entries := make([]string, 0, len(list))
	for _, item := range list {
		text, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s in the config file: entry %v is not provider.key or provider.key=custom_field", key, item)
		}
		entries = append(entries, text)
	}
	return entries, nil
}
