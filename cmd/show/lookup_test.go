package show

import (
	"testing"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/google/uuid"
)

func TestFindInterfacesByNameOrUUIDReturnsEveryNameMatch(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	inv := &devicetypes.Inventory{
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
			firstID:  {ID: firstID, Name: "lag256", DeviceID: uuid.New()},
			secondID: {ID: secondID, Name: "LAG256", DeviceID: uuid.New()},
		},
	}

	matches, err := findInterfacesByNameOrUUID("lag256", inv)
	if err != nil {
		t.Fatalf("findInterfacesByNameOrUUID() error = %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("findInterfacesByNameOrUUID() returned %d matches, want 2", len(matches))
	}
}

func TestFindInterfacesByNameOrUUIDReturnsOnlyUUIDMatch(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	inv := &devicetypes.Inventory{
		Interfaces: map[uuid.UUID]*devicetypes.CaniInterface{
			firstID:  {ID: firstID, Name: "lag256", DeviceID: uuid.New()},
			secondID: {ID: secondID, Name: "lag256", DeviceID: uuid.New()},
		},
	}

	matches, err := findInterfacesByNameOrUUID(firstID.String(), inv)
	if err != nil {
		t.Fatalf("findInterfacesByNameOrUUID() error = %v", err)
	}
	if len(matches) != 1 || matches[0].ID != firstID {
		t.Fatalf("findInterfacesByNameOrUUID() = %v, want only %s", matches, firstID)
	}
}
