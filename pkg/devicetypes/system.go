package devicetypes

import "github.com/google/uuid"

func NewSystem() *CaniDeviceType {
	return &CaniDeviceType{
		Type:     "system", //FIXME
		Name:     "SystemZero",
		ID:       uuid.New(),
		Children: []uuid.UUID{},
		Parent:   uuid.Nil,
	}
}
