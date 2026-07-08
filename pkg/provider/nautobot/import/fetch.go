package imprt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	nautobotapi "github.com/Cray-HPE/cani/pkg/nautobot"
)

// pageSize is the number of items to request per API page.
const pageSize = 100

func intPtr(v int) *int { return &v }

// FetchLocations retrieves all locations from the Nautobot API.
func FetchLocations(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Location, error) {
	var all []nautobotapi.Location
	offset := 0
	for {
		resp, err := client.DcimLocationsListWithResponse(ctx, &nautobotapi.DcimLocationsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list locations: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list locations: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchRacks retrieves all racks from the Nautobot API.
func FetchRacks(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Rack, error) {
	var all []nautobotapi.Rack
	offset := 0
	for {
		resp, err := client.DcimRacksListWithResponse(ctx, &nautobotapi.DcimRacksListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list racks: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list racks: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchDevices retrieves all devices from the Nautobot API.
func FetchDevices(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Device, error) {
	var all []nautobotapi.Device
	offset := 0
	for {
		resp, err := client.DcimDevicesListWithResponse(ctx, &nautobotapi.DcimDevicesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list devices: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list devices: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchDeviceTypes retrieves all device types from the Nautobot API.
func FetchDeviceTypes(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.DeviceType, error) {
	var all []nautobotapi.DeviceType
	offset := 0
	for {
		resp, err := client.DcimDeviceTypesListWithResponse(ctx, &nautobotapi.DcimDeviceTypesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list device types: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list device types: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchInterfaces retrieves all interfaces from the Nautobot API.
func FetchInterfaces(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Interface, error) {
	var all []nautobotapi.Interface
	offset := 0
	for {
		resp, err := client.DcimInterfacesListWithResponse(ctx, &nautobotapi.DcimInterfacesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list interfaces: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list interfaces: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchModules retrieves all modules from the Nautobot API.
func FetchModules(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Module, error) {
	var all []nautobotapi.Module
	offset := 0
	for {
		resp, err := client.DcimModulesListWithResponse(ctx, &nautobotapi.DcimModulesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list modules: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list modules: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchModuleBays retrieves all module bays from the Nautobot API.
func FetchModuleBays(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.ModuleBay, error) {
	var all []nautobotapi.ModuleBay
	offset := 0
	for {
		resp, err := client.DcimModuleBaysListWithResponse(ctx, &nautobotapi.DcimModuleBaysListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list module bays: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list module bays: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchCables retrieves all cables from the Nautobot API.
func FetchCables(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Cable, error) {
	var all []nautobotapi.Cable
	offset := 0
	for {
		resp, err := client.DcimCablesListWithResponse(ctx, &nautobotapi.DcimCablesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list cables: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list cables: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchInventoryItems retrieves all inventory items (FRUs) from the Nautobot API.
func FetchInventoryItems(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.InventoryItem, error) {
	var all []nautobotapi.InventoryItem
	offset := 0
	for {
		resp, err := client.DcimInventoryItemsListWithResponse(ctx, &nautobotapi.DcimInventoryItemsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list inventory items: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list inventory items: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchStatuses retrieves all statuses from the Nautobot API.
func FetchStatuses(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Status, error) {
	var all []nautobotapi.Status
	offset := 0
	for {
		resp, err := client.ExtrasStatusesListWithResponse(ctx, &nautobotapi.ExtrasStatusesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list statuses: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list statuses: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchRoles retrieves all roles from the Nautobot API.
func FetchRoles(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Role, error) {
	var all []nautobotapi.Role
	offset := 0
	for {
		resp, err := client.ExtrasRolesListWithResponse(ctx, &nautobotapi.ExtrasRolesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list roles: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list roles: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchVRFs retrieves all VRFs from the Nautobot API.
func FetchVRFs(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.VRF, error) {
	var all []nautobotapi.VRF
	offset := 0
	for {
		resp, err := client.IpamVrfsListWithResponse(ctx, &nautobotapi.IpamVrfsListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list vrfs: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list vrfs: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchVLANs retrieves all VLANs from the Nautobot API.
func FetchVLANs(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.VLAN, error) {
	var all []nautobotapi.VLAN
	offset := 0
	for {
		resp, err := client.IpamVlansListWithResponse(ctx, &nautobotapi.IpamVlansListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list vlans: %w", err)
		}
		if resp.StatusCode() != http.StatusOK {
			return nil, fmt.Errorf("list vlans: status %d", resp.StatusCode())
		}
		if resp.JSON200 == nil || len(resp.JSON200.Results) == 0 {
			break
		}
		all = append(all, resp.JSON200.Results...)
		if resp.JSON200.Next == nil || *resp.JSON200.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// The generated client types Nautobot's network/broadcast/host address fields
// as []byte, which Go's encoding/json decodes from base64. Nautobot returns them
// as plain dotted-decimal strings (e.g. "10.0.0.0"), so decoding the generated
// Prefix/IPAddress directly fails with "illegal base64 data". These shadow types
// re-map those keys to json.RawMessage (which the transform ignores) so the rest
// of the object decodes normally. Field names must match the embedded struct so
// the shallower field wins during unmarshal.
type prefixNoBinary struct {
	nautobotapi.Prefix
	Network   json.RawMessage `json:"network"`
	Broadcast json.RawMessage `json:"broadcast"`
}

type ipAddressNoBinary struct {
	nautobotapi.IPAddress
	Host json.RawMessage `json:"host"`
}

// decodeListPage reads a raw list response body into dest, closing the body and
// verifying a 200 status. It is used for endpoints whose generated response
// parser mishandles Nautobot's []byte address fields.
func decodeListPage(resp *http.Response, dest any) error {
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, dest)
}

// FetchPrefixes retrieves all prefixes from the Nautobot API.
func FetchPrefixes(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.Prefix, error) {
	var all []nautobotapi.Prefix
	offset := 0
	for {
		resp, err := client.IpamPrefixesList(ctx, &nautobotapi.IpamPrefixesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list prefixes: %w", err)
		}
		var page struct {
			Next    *string          `json:"next"`
			Results []prefixNoBinary `json:"results"`
		}
		if err := decodeListPage(resp, &page); err != nil {
			return nil, fmt.Errorf("list prefixes: %w", err)
		}
		if len(page.Results) == 0 {
			break
		}
		for i := range page.Results {
			all = append(all, page.Results[i].Prefix)
		}
		if page.Next == nil || *page.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}

// FetchIPAddresses retrieves all IP addresses from the Nautobot API.
func FetchIPAddresses(ctx context.Context, client *nautobotapi.ClientWithResponses) ([]nautobotapi.IPAddress, error) {
	var all []nautobotapi.IPAddress
	offset := 0
	for {
		resp, err := client.IpamIpAddressesList(ctx, &nautobotapi.IpamIpAddressesListParams{
			Limit:  intPtr(pageSize),
			Offset: &offset,
		})
		if err != nil {
			return nil, fmt.Errorf("list ip addresses: %w", err)
		}
		var page struct {
			Next    *string             `json:"next"`
			Results []ipAddressNoBinary `json:"results"`
		}
		if err := decodeListPage(resp, &page); err != nil {
			return nil, fmt.Errorf("list ip addresses: %w", err)
		}
		if len(page.Results) == 0 {
			break
		}
		for i := range page.Results {
			all = append(all, page.Results[i].IPAddress)
		}
		if page.Next == nil || *page.Next == "" {
			break
		}
		offset += pageSize
	}
	return all, nil
}
